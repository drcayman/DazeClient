package encryption

import (
	"net"
	"crypto/aes"
	"crypto/cipher"
	"github.com/pkg/errors"
	"crypto/md5"
	"encoding/hex"
)

type PskAes256Cfb struct {
	reserved string
}
type PskAesCfb256Tmp struct {
	Key []byte
	Block cipher.Block
}
func (this *PskAes256Cfb) GenKey(key string) ([]byte,error){
	test := md5.New()
	_,err:=test.Write([]byte(key))
	if err!=nil{
		return nil,err
	}
	md5src:=test.Sum(nil)
	md5dst:=make([]byte,32)
	hex.Encode(md5dst,md5src)
	return md5dst,nil

}
func (this *PskAes256Cfb) Init(arg string,client *interface{})(error){
	key,GenKeyErr:=this.GenKey(arg)
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
}
func (this *PskAes256Cfb)InitUser(conn net.Conn,client *interface{})(error){
	return nil
}
func (this *PskAes256Cfb)Encrypt(client *interface{},data []byte)([][]byte,error){
	t,flag:=(*client).(PskAesCfb256Tmp)
	if !flag{
		return nil,errors.New("unknown error")
	}
	dst:=make([]byte,len(data))
	Crypter:=cipher.NewCFBEncrypter(t.Block,t.Key)
	Crypter.XORKeyStream(dst,data)
	list:=make([][]byte,0)
	list=append(list,dst)
	return list,nil
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
