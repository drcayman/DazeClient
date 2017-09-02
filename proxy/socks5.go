package proxy

import (
	"io"
	"fmt"
	"net"
	"strings"
	"github.com/crabkun/DazeClient/server"
	"github.com/crabkun/DazeClient/common"
	"errors"
	"encoding/json"
	"bytes"
)
// Convert a IP:Port string to a byte array in network order.
// e.g.: 74.125.31.104:80 -> [74 125 31 104 0 80]
func packNetAddr4(addr string, buf []byte) {
	IP,err:=net.ResolveTCPAddr("tcp",addr)
	if err!=nil{
		panic(fmt.Sprintf("invalid address %s", addr))
	}
	copy(buf[:4], IP.IP.To4())
	buf[4] = byte(IP.Port / 256)
	buf[5] = byte(IP.Port % 256)
}
func packNetAddr6(addr string, buf []byte) {
	IP,err:=net.ResolveTCPAddr("tcp",addr)
	if err!=nil{
		panic(fmt.Sprintf("invalid address %s", addr))
	}
	copy(buf,IP.IP.To16())
	buf[16] = byte(IP.Port / 256)
	buf[17] = byte(IP.Port % 256)
}
// Read a specified number of bytes.
func readBytes(conn io.Reader, count int) (buf []byte) {
	buf = make([]byte, count)
	if _, err := io.ReadFull(conn, buf); err != nil {
		panic(err)
	}
	return
}

func protocolCheck(assert bool) {
	if !assert {
		panic("protocol error")
	}
}

func errorReplyConnect(reason byte) []byte {
	return []byte{0x05, reason, 0x00, 0x01,
				  0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}

func performConnect(backend string, frontconn net.Conn) {
	ProxyClient:=server.CallProxyServer(frontconn,common.SrvConf,backend,"tcp")
	if ProxyClient==nil {
		frontconn.Write(errorReplyConnect(0x05))
		return
	}
	backaddr := ProxyClient.TargetHostRealAddr
	defer func() {
		ProxyClient.LocalDisconnect=true
		ProxyClient.RemoteServerConn.Close()
	}()
	IsIPv6:=IsIPv6Address(backaddr)
	// reply to the CONNECT command
	var buf []byte
	if !IsIPv6{  //IPv4
		buf = make([]byte, 10)
		copy(buf, []byte{0x05, 0x00, 0x00, 0x01})
		packNetAddr4(backaddr, buf[4:])
	}else{  //IPv6
		buf = make([]byte, 22)
		copy(buf, []byte{0x05, 0x00, 0x00, 0x04})
		packNetAddr6(backaddr, buf[4:])
	}
	frontconn.Write(buf)

	go SocksTCPBridgeRemoteToProxy(ProxyClient)
	SocksTCPBridgeProxyToRemote(ProxyClient)
}
func IsIPv6Address(addr string) bool {
	return strings.Count(addr,":")>1
}
func SocksTCPBridgeRemoteToProxy(client *server.S_Client){
	defer func(){
		recover()
		client.ProxyUserConn.Close()
	}()
	for{
		client.SafeSend(client.Read(),client.ProxyUserConn)
	}
}
func SocksTCPBridgeProxyToRemote(client *server.S_Client){
	defer func(){
		recover()
		client.RemoteServerConn.Close()
	}()
	buf:=make([]byte,56789)
	var n int
	var err error
	for{
		n,err=client.ProxyUserConn.Read(buf)
		if err!=nil{
			return
		}
		client.Write(buf[:n])
	}
}
func Socks5handleConnection(frontconn net.Conn) {
	defer func() {
		recover()
		frontconn.Close()

	}()

	// receive auth packet
	buf1 := readBytes(frontconn, 2)
	protocolCheck(buf1[0] == 0x05)  // VER

	nom := int(buf1[1])  // number of methods
	methods := readBytes(frontconn, nom)

	var support bool
	for _, meth := range methods {
		if meth == 0x00 {
			support = true
			break
		}
	}
	if !support {
		// X'FF' NO ACCEPTABLE METHODS
		frontconn.Write([]byte{0x05, 0xff})
		return
	}

	// X'00' NO AUTHENTICATION REQUIRED
	frontconn.Write([]byte{0x05, 0x00})

	// recv command packet
	buf3 := readBytes(frontconn, 4)
	protocolCheck(buf3[0] == 0x05)  // VER
	protocolCheck(buf3[2] == 0x00)  // RSV

	command := buf3[1]
	if command != 0x01 && command != 0x03{  // 0x01: CONNECT
		// X'07' Command not supported
		frontconn.Write(errorReplyConnect(0x07))
		return
	}

	addrtype := buf3[3]
	if addrtype != 0x01 && addrtype != 0x03 && addrtype != 0x04  {
		// X'08' Address type not supported
		frontconn.Write(errorReplyConnect(0x08))
		return
	}

	var backend string
	switch addrtype {
	case 0x01: //IPv4
		buf4 := readBytes(frontconn, 6)
		backend = fmt.Sprintf("%d.%d.%d.%d:%d", buf4[0], buf4[1],
			buf4[2], buf4[3], int(buf4[4]) * 256 + int(buf4[5]))
	case 0x03: //DOMAINNAME
		buf4 := readBytes(frontconn, 1)
		nmlen := int(buf4[0])  // domain name length
		if nmlen > 253 {
			panic("domain name too long")  // will be recovered
		}

		buf5 := readBytes(frontconn, nmlen + 2)
		backend = fmt.Sprintf("%s:%d", buf5[0:nmlen],
			int(buf5[nmlen]) * 256 + int(buf5[nmlen+1]))
	case 0x04:
		buf4 := readBytes(frontconn, 18)
		backend = fmt.Sprintf("[%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x]:%d",
			buf4[0], buf4[1],
			buf4[2], buf4[3],
			buf4[4], buf4[5],
			buf4[6], buf4[7],
			buf4[8], buf4[9],
			buf4[10], buf4[11],
			buf4[12], buf4[13],
			buf4[14], buf4[15],
			int(buf4[16]) * 256 + int(buf4[17]))
	}
	if command==0x03{  //UDP
		performUdp(frontconn)
	}else{//TCP
		performConnect(backend, frontconn)
	}

}
func performUdp(frontconn net.Conn){
	addr,_:=net.ResolveUDPAddr("udp4","0.0.0.0")
	conn,err:=net.ListenUDP("udp4",addr)
	if err!=nil{
		return
	}
	defer func(){
		conn.Close()
	}()
	buf := make([]byte, 10)
	copy(buf, []byte{0x05, 0x00, 0x00, 0x01})
	packNetAddr4(conn.LocalAddr().String(), buf[4:])
	frontconn.Write(buf)
	client:=server.CallProxyServer(frontconn,common.SrvConf,"","udp")
	if client==nil{
		return
	}
	go SocksUDPBridgeRemoteToProxy(client,conn)
	SocksUDPBridgeProxyToRemote(client,conn)
}
func GetUDPAddress(rd io.Reader)(string){
	pkt:=readBytes(rd,4)
	if pkt[3]!=0x01 && pkt[3]!=0x03 && pkt[3]!=0x04{
		panic("invalid target addr")
	}
	var address string
	if pkt[3]==0x01{ //ipv4
		buf4 := readBytes(rd, 6)
		address = fmt.Sprintf("%d.%d.%d.%d:%d", buf4[0], buf4[1],
			buf4[2], buf4[3], int(buf4[4]) * 256 + int(buf4[5]))
	}else if pkt[3]==0x03{ //NAME
		buf4 := readBytes(rd, 1)
		nmlen := int(buf4[0])  // domain name length
		if nmlen > 253 {
			panic("domain name too long")  // will be recovered
		}

		buf5 := readBytes(rd, nmlen + 2)
		address = fmt.Sprintf("%s:%d", buf5[0:nmlen],
			int(buf5[nmlen]) * 256 + int(buf5[nmlen+1]))
	}else if pkt[3]==0x04{ //ipv6
		buf4 := readBytes(rd, 18)
		address = fmt.Sprintf("[%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x]:%d",
			buf4[0], buf4[1],
			buf4[2], buf4[3],
			buf4[4], buf4[5],
			buf4[6], buf4[7],
			buf4[8], buf4[9],
			buf4[10], buf4[11],
			buf4[12], buf4[13],
			buf4[14], buf4[15],
			int(buf4[16]) * 256 + int(buf4[17]))
	}
	return address
}
func SafeReadUDPBytes(conn *net.UDPConn) ([]byte,*net.UDPAddr,*net.UDPAddr,error){
	var err error
	buf:=make([]byte,65507)
	n,clientAddr,err:=conn.ReadFromUDP(buf)
	if err!=nil{
		return nil,nil,nil,errors.New("read error")
	}
	reader:=bytes.NewBuffer(buf[:n])
	targetAddr,err:=net.ResolveUDPAddr("udp",GetUDPAddress(reader))
	if err!=nil{
		return nil,nil,nil,err
	}
	return reader.Bytes(),targetAddr,clientAddr,nil
}
func SocksUDPBridgeRemoteToProxy(client *server.S_Client,UdpClient *net.UDPConn){
	defer func(){
		recover()
		UdpClient.Close()
	}()
	var UDP common.Json_UDP
	var ADDR *net.UDPAddr
	var LastAddr string
	var err error
	var header []byte
	for{
		buf:=client.Read()
		err=json.Unmarshal(buf,&UDP)
		if err!=nil{
			return
		}
		if LastAddr!=UDP.Host{
			ADDR,err=net.ResolveUDPAddr("udp",UDP.Host)
			if err!=nil{
				return
			}
			if IsIPv6Address(ADDR.String()){
				header=make([]byte,22)
				header[3]=4
				copy(header[4:], ADDR.IP.To16())
				header[20] = byte(ADDR.Port / 256)
				header[21] = byte(ADDR.Port % 256)

			}else{
				header=make([]byte,10)
				header[3]=1
				copy(header[4:], ADDR.IP.To4())
				header[8] = byte(ADDR.Port / 256)
				header[9] = byte(ADDR.Port % 256)
			}
			LastAddr=UDP.Host
		}
		UdpClient.WriteTo(header,client.UDPAddr)
		UdpClient.WriteTo(UDP.Data,client.UDPAddr)
	}
}
func SocksUDPBridgeProxyToRemote(client *server.S_Client,UdpClient *net.UDPConn){
	defer func(){
		recover()
		client.RemoteServerConn.Close()
	}()
	var UDP common.Json_UDP
	var err error
	var jsonBuf []byte
	var targetHost *net.UDPAddr
	for{
		UDP.Data,targetHost,client.UDPAddr,err=SafeReadUDPBytes(UdpClient)
		if err!=nil{
			return
		}
		UDP.Host=targetHost.String()
		jsonBuf,err=json.Marshal(UDP)
		if err!=nil{
			return
		}
		client.Write(jsonBuf)
	}

}
