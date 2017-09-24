package encryption

import (
	"net"
	"crypto/aes"
	"crypto/cipher"
	"github.com/crabkun/DazeClient/util"
)

type PskAes256Cfb struct {
	Key []byte
	Block cipher.Block
}
func (this *PskAes256Cfb)InitUser(conn net.Conn,param string)(error){
	key,GenKeyErr:=util.Gen32Md5Key(param)
	if GenKeyErr!=nil{
		return GenKeyErr
	}
	var CipherErr error=nil
	this.Block,CipherErr=aes.NewCipher(key)
	if CipherErr!=nil{
		return CipherErr
	}
	this.Key=key[:this.Block.BlockSize()]
	return nil
	return nil
}
func (this *PskAes256Cfb)Encrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	Crypter:=cipher.NewCFBEncrypter(this.Block,this.Key)
	Crypter.XORKeyStream(dst,data)
	return dst,nil
}
func (this *PskAes256Cfb)Decrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	Decrypter:=cipher.NewCFBDecrypter(this.Block,this.Key)
	Decrypter.XORKeyStream(dst,data)
	return dst,nil
}
func init(){
	RegisterEncryption("psk-aes-256-cfb",new(PskAes256Cfb))
}