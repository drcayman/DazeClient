package encryption

import (
	"net"
	"crypto/rand"
	"strings"
	"errors"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"io"
	"math/big"
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
	pub:=&rsa.PublicKey{N:big.NewInt(0).SetBytes(buf),E:65537}
	if err!=nil{
		return errors.New("无法解析服务器发送过来的公钥！原因："+err.Error())
	}
	newAesKey:=this.GenBytes(16)
	finallyAesKey,err:=rsa.EncryptPKCS1v15(rand.Reader,pub,newAesKey)
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