package encryption

import (
	"net"
	"crypto/rc4"
	"errors"
	"github.com/crabkun/DazeClient/util"
)

type PskRc4Md5 struct {
	reserved string
}
type PskRc4Md5Tmp struct {
	Cipher *rc4.Cipher
}
func (this *PskRc4Md5)InitUser(conn net.Conn,param string,client *interface{})(error){
	key,GenKeyErr:=util.Gen16Md5Key(param)
	if GenKeyErr!=nil{
		return GenKeyErr
	}
	t:=PskRc4Md5Tmp{}
	var CipherErr error=nil
	t.Cipher,CipherErr=rc4.NewCipher(key)
	if CipherErr!=nil{
		return CipherErr
	}
	*client=t
	return nil
}
func (this *PskRc4Md5)Encrypt(client *interface{},data []byte)([]byte,error){
	t,flag:=(*client).(PskRc4Md5Tmp)
	if !flag{
		return nil,errors.New("unknown error")
	}
	dst:=make([]byte,len(data))
	t.Cipher.Reset()
	t.Cipher.XORKeyStream(dst,data)
	return dst,nil
}
func (this *PskRc4Md5)Decrypt(client *interface{},data []byte)([]byte,error){
	t,flag:=(*client).(PskRc4Md5Tmp)
	if !flag{
		return nil,errors.New("unknown error")
	}
	dst:=make([]byte,len(data))
	t.Cipher.Reset()
	t.Cipher.XORKeyStream(dst,data)
	return dst,nil
}
