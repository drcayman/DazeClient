package obscure

import (
	"crypto/tls"
	"net"
)

type TlsHandshake struct {
	RegArg string
}

func (this *TlsHandshake) Action(conn net.Conn , param string) (error){
	c:=tls.Client(conn,&tls.Config{InsecureSkipVerify:true})
	return c.Handshake()
}

