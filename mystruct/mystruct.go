package mystruct

import (
	"sync"
	"net"
)

type Packet struct {
	Command byte
	Data []byte
}
type ProxyClientSturct struct{
	Conn net.Conn
	Remote net.Conn
	RemoteRealAddr string
	ProxyUser net.Conn
	ProxyBrideChan chan Packet
	ProxyRecvChan chan Packet
	IsKeyExchange bool
	IsConnected bool
	AESKey []byte
	Locker sync.Mutex

}
