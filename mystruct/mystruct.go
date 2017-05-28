package mystruct

import (
	"net"
)
type ProxyClientSturct struct{
	Remote net.Conn
	RemoteRealAddr string
	ProxyUser net.Conn
	IsKeyExchange bool
	IsConnected bool
	AESKey []byte
	IsUDP bool
	Address string
	Protocol string
}
