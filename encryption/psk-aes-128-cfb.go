package encryption

import (
	"net"
	"crypto/aes"
	"crypto/cipher"
	"github.com/crabkun/DazeClient/util"
)

type PskAesCfb struct {
	Key []byte
	Block cipher.Block
}

func (this *PskAesCfb)InitUser(conn net.Conn,param string)(error){
	key,GenKeyErr:=util.Gen16Md5Key(param)
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
}
func (this *PskAesCfb)Encrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	Crypter:=cipher.NewCFBEncrypter(this.Block,this.Key)
	Crypter.XORKeyStream(dst,data)
	return dst,nil
}
func (this *PskAesCfb)Decrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	Decrypter:=cipher.NewCFBDecrypter(this.Block,this.Key)
	Decrypter.XORKeyStream(dst,data)
	return dst,nil
}
func init(){
	RegisterEncryption("psk-aes-128-cfb",new(PskAesCfb))
}