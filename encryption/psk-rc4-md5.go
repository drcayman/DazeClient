package encryption

import (
	"net"
	"crypto/rc4"
	"github.com/crabkun/DazeClient/util"
)

type PskRc4Md5 struct {
	Cipher *rc4.Cipher
}

func (this *PskRc4Md5)InitUser(conn net.Conn,param string)(error){
	key,GenKeyErr:=util.Gen16Md5Key(param)
	if GenKeyErr!=nil{
		return GenKeyErr
	}
	var CipherErr error=nil
	this.Cipher,CipherErr=rc4.NewCipher(key)
	if CipherErr!=nil{
		return CipherErr
	}
	return nil
}
func (this *PskRc4Md5)Encrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	this.Cipher.Reset()
	this.Cipher.XORKeyStream(dst,data)
	return dst,nil
}
func (this *PskRc4Md5)Decrypt(data []byte)([]byte,error){
	dst:=make([]byte,len(data))
	this.Cipher.Reset()
	this.Cipher.XORKeyStream(dst,data)
	return dst,nil
}
func init(){
	RegisterEncryption("psk-rc4-md5",new(PskRc4Md5))
}