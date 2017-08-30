package encryption

import "net"

//none-无加密
type none struct {
	RegArg string
}

func (this *none)InitUser(conn net.Conn,param string,client *interface{})(error){
	return nil
}
func (this *none)Encrypt(client *interface{},data []byte)([]byte,error){
	return data,nil
}
func (this *none)Decrypt(client *interface{},data []byte)([]byte,error){
	return data,nil
}

