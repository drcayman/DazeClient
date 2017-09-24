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
	"io"
)

type KeypairAes struct {
	Key []byte
	Block cipher.Block
}

func (this *KeypairAes)SafeRead(conn net.Conn,length int)([]byte,error){
	buf:=make([]byte,length)
	_,err:=io.ReadFull(conn,buf)
	return buf,err
}
func (this *KeypairAes)InitUser(conn net.Conn,param string)(error){
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
	this.Block,err=aes.NewCipher(newAesKey)
	if err!=nil{
		return err
	}
	this.Key=newAesKey[:this.Block.BlockSize()]
	return nil
}
func (this *KeypairAes)Encrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	Crypter:=cipher.NewCFBEncrypter(this.Block,this.Key)
	Crypter.XORKeyStream(dst,data)
	return dst,nil
}
func (this *KeypairAes)Decrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	Decrypter:=cipher.NewCFBDecrypter(this.Block,this.Key)
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
func init(){
	RegisterEncryption("keypair-aes",new(KeypairAes))
}