package disguise

import (
	"net"
	"crypto/tls"
)

type TlsHandshake struct {
	reserved string
}

func (this *TlsHandshake) Init(arg string,client *interface{})(error){
	return nil
}

func (this *TlsHandshake) Action(conn net.Conn ,client *interface{}) (error){
	c:=tls.Client(conn,&tls.Config{InsecureSkipVerify:true})
	return c.Handshake()
}
