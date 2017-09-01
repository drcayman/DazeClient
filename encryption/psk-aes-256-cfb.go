package encryption

import (
	"net"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"github.com/crabkun/DazeClient/util"
)

type PskAes256Cfb struct {
	reserved string
}
type PskAesCfb256Tmp struct {
	Key []byte
	Block cipher.Block
}
func (this *PskAes256Cfb)InitUser(conn net.Conn,param string,client *interface{})(error){
	key,GenKeyErr:=util.Gen32Md5Key(param)
	if GenKeyErr!=nil{
		return GenKeyErr
	}
	t:=PskAesCfb256Tmp{}

	var CipherErr error=nil
	t.Block,CipherErr=aes.NewCipher(key)
	if CipherErr!=nil{
		return CipherErr
	}
	t.Key=key[:t.Block.BlockSize()]
	*client=t
	return nil
	return nil
}
func (this *PskAes256Cfb)Encrypt(client *interface{},data []byte)([]byte,error){
	t,flag:=(*client).(PskAesCfb256Tmp)
	if !flag{
		return nil,errors.New("unknown error")
	}
	dst:=make([]byte,len(data))
	Crypter:=cipher.NewCFBEncrypter(t.Block,t.Key)
	Crypter.XORKeyStream(dst,data)
	return dst,nil
}
func (this *PskAes256Cfb)Decrypt(client *interface{},data []byte)([]byte,error){
	t,flag:=(*client).(PskAesCfb256Tmp)
	if !flag{
		return nil,errors.New("unknown error")
	}
	dst:=make([]byte,len(data))
	Decrypter:=cipher.NewCFBDecrypter(t.Block,t.Key)
	Decrypter.XORKeyStream(dst,data)
	return dst,nil
}
