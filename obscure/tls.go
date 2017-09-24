package obscure

import (
	"crypto/tls"
	"net"
)

type TlsHandshake struct {
}

func (this *TlsHandshake) Action(conn net.Conn , param string) (error){
	c:=tls.Client(conn,&tls.Config{InsecureSkipVerify:true})
	return c.Handshake()
}
func init(){
	RegisterObscure("tls_handshake",new(TlsHandshake))
}

