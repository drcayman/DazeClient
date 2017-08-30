package obscure

import "net"

//none-无伪装
type none struct {
	RegArg string
}

func (this *none) Action(conn net.Conn , param string) (error){
	return nil
}
