package encryption

import "net"

//none-无加密
type none struct {
}

func (this *none)InitUser(conn net.Conn,param string)(error){
	return nil
}
func (this *none)Encrypt(data []byte)([]byte,error){
	return data,nil
}
func (this *none)Decrypt(data []byte)([]byte,error){
	return data,nil
}
func init(){
	RegisterEncryption("none",new(none))
}
