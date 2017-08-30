package encryption

import (
	"net"
	"crypto/rand"
	"crypto/x509"

	"strings"

	"crypto/md5"
	"errors"
	"time"
	"strconv"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
)

type KeypairAes struct {
	reserved string
}
type KeypairAesTmp struct {
	Key []byte
	Block cipher.Block
}
func (this *KeypairAes)InitUser(conn net.Conn,param string,client *interface{})(error){
	pos:=0
	buf:=make([]byte,294)
	for pos<294{
		n,err:=conn.Read(buf[pos:])
		if err!=nil{
			return errors.New("服务器在握手期间断开连接"+err.Error())
		}
		pos+=n
	}
	utc:=time.Now().UTC()
	s,ParseDuration:=time.ParseDuration(utc.Format("-15h04m05s"))
	if ParseDuration!=nil{
		return ParseDuration
	}

	utc=utc.Add(s)
	UTCunix:=utc.Unix()
	UTCunixStr:=strconv.FormatInt(UTCunix,10)
	UTCunixStrPadded:=this.StrPadding(UTCunixStr)

	aesKey,GenMd5Err:=this.GenMd5Key(UTCunixStrPadded)
	if GenMd5Err!=nil{
		return GenMd5Err
	}

	Cipher,CipherErr:=aes.NewCipher(aesKey)
	if CipherErr!=nil{
		return CipherErr
	}
	enc:=cipher.NewCFBDecrypter(Cipher,aesKey[:Cipher.BlockSize()])
	keyDecoded:=make([]byte,len(buf))
	enc.XORKeyStream(keyDecoded,buf)

	pub,pubErr:=x509.ParsePKIXPublicKey(keyDecoded)
	if pubErr!=nil{
		return errors.New("无法解析服务器发送过来的公钥！原因："+pubErr.Error())
	}
	publicKey,flag:=pub.(*rsa.PublicKey)
	if !flag{
		return errors.New("无法解析服务器发送过来的公钥！")
	}
	newAesKey:=this.GenBytes(16)
	finallyAesKey,EncryptAesKeyErr:=rsa.EncryptPKCS1v15(rand.Reader,publicKey,newAesKey)
	if EncryptAesKeyErr!=nil{
		return errors.New("无法利用服务器返回的公钥加密！原因："+EncryptAesKeyErr.Error())
	}
	conn.Write(finallyAesKey)

	t:=KeypairAesTmp{}
	t.Block,CipherErr=aes.NewCipher(newAesKey)
	if CipherErr!=nil{
		return CipherErr
	}
	t.Key=newAesKey[:t.Block.BlockSize()]
	*client=t

	return nil
}
func (this *KeypairAes)Encrypt(client *interface{},data []byte)([]byte,error){
	t,flag:=(*client).(KeypairAesTmp)
	if !flag{
		return nil,errors.New("unknown error")
	}
	dst:=make([]byte,len(data))
	Crypter:=cipher.NewCFBEncrypter(t.Block,t.Key)
	Crypter.XORKeyStream(dst,data)
	return dst,nil
}
func (this *KeypairAes)Decrypt(client *interface{},data []byte)([]byte,error){
	t,flag:=(*client).(KeypairAesTmp)
	if !flag{
		return nil,errors.New("unknown error")
	}
	dst:=make([]byte,len(data))
	Decrypter:=cipher.NewCFBDecrypter(t.Block,t.Key)
	Decrypter.XORKeyStream(dst,data)
	return dst,nil
}

func (this *KeypairAes)StrPadding(str string) string {
	l:=16-len(str)
	newstr:=str+strings.Repeat("0",l)
	return newstr
}
func (this *KeypairAes) GenMd5Key(key string) ([]byte,error){
	test := md5.New()
	_,err:=test.Write([]byte(key))
	if err!=nil{
		return nil,err
	}
	return test.Sum(nil),nil
}
func (this *KeypairAes) GenBytes(bytesLen int) ([]byte){
	buf:=make([]byte,bytesLen)
	rand.Read(buf)
	return buf
}