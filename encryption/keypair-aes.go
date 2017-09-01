package encryption

import (
	"net"
	"crypto/rand"
	"crypto/x509"
	"strings"
	"errors"
	"time"
	"strconv"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"github.com/crabkun/DazeClient/util"
)

type KeypairAes struct {
	reserved string
}
type KeypairAesTmp struct {
	Key []byte
	Block cipher.Block
}
func (this *KeypairAes)SafeRead(conn net.Conn,length int)([]byte,error){
	buf:=make([]byte,length)
	for pos:=0;pos<length;{
		n,err:=conn.Read(buf[pos:])
		if err!=nil {
			return nil,err
		}
		pos+=n
	}
	return buf,nil
}
func (this *KeypairAes)InitUser(conn net.Conn,param string,client *interface{})(error){
	var buf []byte
	var err error
	buf,err=this.SafeRead(conn,1)
	if err!=nil{
		return err
	}
	buf,err=this.SafeRead(conn,int(buf[0]))
	if err!=nil{
		return err
	}
	utc:=time.Now().UTC()
	s,err:=time.ParseDuration(utc.Format("-15h04m05s"))
	if err!=nil{
		return err
	}
	utc=utc.Add(s)
	UTCunix:=utc.Unix()
	UTCunixStr:=strconv.FormatInt(UTCunix,10)
	UTCunixStrPadded:=this.StrPadding(UTCunixStr)
	aesKey,err:=util.Gen16Md5Key(UTCunixStrPadded)
	if err!=nil{
		return err
	}
	Cipher,err:=aes.NewCipher(aesKey)
	if err!=nil{
		return err
	}
	enc:=cipher.NewCFBDecrypter(Cipher,aesKey[:Cipher.BlockSize()])
	keyDecoded:=make([]byte,len(buf))
	enc.XORKeyStream(keyDecoded,buf)
	pub,err:=x509.ParsePKIXPublicKey(keyDecoded)
	if err!=nil{
		return errors.New("无法解析服务器发送过来的公钥！原因："+err.Error())
	}
	publicKey,flag:=pub.(*rsa.PublicKey)
	if !flag{
		return errors.New("无法解析服务器发送过来的公钥！")
	}
	newAesKey:=this.GenBytes(16)
	finallyAesKey,err:=rsa.EncryptPKCS1v15(rand.Reader,publicKey,newAesKey)
	if err!=nil{
		return errors.New("无法利用服务器返回的公钥加密！原因："+err.Error())
	}
	conn.Write(finallyAesKey)
	t:=KeypairAesTmp{}
	t.Block,err=aes.NewCipher(newAesKey)
	if err!=nil{
		return err
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

func (this *KeypairAes) GenBytes(bytesLen int) ([]byte){
	buf:=make([]byte,bytesLen)
	rand.Read(buf)
	return buf
}