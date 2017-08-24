package common

import (
	"net"
	"DazeClient/encryption"
	"DazeClient/disguise"
)
type ProxyClientSturct struct{
	Remote net.Conn
	RemoteRealAddr string
	ProxyUser net.Conn
	IsConnected bool
	IsUDP bool
	Address string
	Protocol string
	UDPHeader []byte
	UDPTarget *net.UDPAddr
	Disguise disguise.DisguiseAction
	Encryption encryption.EncryptionAction
	EncReserved interface{}
	DsgReserved interface{}
}
type JsonAuth struct{
	Username string
	Password string
	Net string
	Host string
	Port string
}