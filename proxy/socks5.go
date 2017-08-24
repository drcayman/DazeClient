package proxy

import (
	"io"
	"fmt"
	"net"
	"strings"
	"strconv"
	"DazeClient/client"
	"DazeClient/common"
	"DazeClient/config"
	"log"
	"bytes"
	"github.com/pkg/errors"
)
// Convert a IP:Port string to a byte array in network order.
// e.g.: 74.125.31.104:80 -> [74 125 31 104 0 80]
func packNetAddr4(addr string, buf []byte) {
	ipport := addr
	pair := strings.Split(ipport, ":")
	ipstr, portstr := pair[0], pair[1]
	port, err := strconv.Atoi(portstr)
	if err != nil {
		panic(fmt.Sprintf("invalid address %s", ipport))
	}

	copy(buf[:4], net.ParseIP(ipstr).To4())
	buf[4] = byte(port / 256)
	buf[5] = byte(port % 256)
}
func packNetAddr6(addr string, buf []byte) {
	ipport := addr
	pos:=strings.LastIndex(ipport,":")
	ipstr, portstr := ipport[:pos],ipport[pos+1:]
	port, err := strconv.Atoi(portstr)
	if err != nil {
		panic(fmt.Sprintf("invalid ipv6 address %s(%s)", ipport,err.Error()))
	}

	copy(buf, net.ParseIP(ipstr).To16())
	buf[16] = byte(port / 256)
	buf[17] = byte(port % 256)
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
	//if config.GetDebug(){
	//	log.Println("trying to connect to ", backend)
	//}
	////backconn, err := net.Dial("tcp", backend)
	ProxyClient,err:=client.NewProxyConn(backend,frontconn,true)
	if ProxyClient==nil || err != nil {
		//if config.GetDebug() {
		//	log.Println("failed to connect to ", backend, err)
		//}
		frontconn.Write(errorReplyConnect(0x05))
		return
	}
	backaddr := ProxyClient.RemoteRealAddr
	if config.GetDebug() {
		log.Println("CONNECTED backend", backaddr)
	}
	defer func() {
		ProxyClient.Remote.Close()
		if config.GetDebug() {
			log.Println("DISCONNECTED backend", backaddr)
		}
	}()
	IsIPv6:=IsIPv6Address(backaddr)
	// reply to the CONNECT command
	var buf []byte
	if !IsIPv6{  //IPv4
		buf = make([]byte, 10)
		copy(buf, []byte{0x05, 0x00, 0x00, 0x01})
		packNetAddr4(ProxyClient.RemoteRealAddr, buf[4:])
	}else{  //IPv6
		buf = make([]byte, 22)
		copy(buf, []byte{0x05, 0x00, 0x00, 0x04})
		packNetAddr6(ProxyClient.RemoteRealAddr, buf[4:])
	}
	frontconn.Write(buf)
	//// bridge connection
	//shutdown := make(chan bool, 2)
	//go iobridge(frontconn, backconn, shutdown)
	//go iobridge(backconn, frontconn, shutdown)
	//
	//// wait for either side to close
	//<-shutdown
	go SocksTCPBridgeRemoteToProxy(ProxyClient)
	SocksTCPBridgeProxyToRemote(ProxyClient)
}
func IsIPv6Address(addr string) bool {
	return strings.Count(addr,":")>1
}
func SocksTCPBridgeRemoteToProxy(ProxyClient *common.ProxyClientSturct){
	for{
		buf,err:=client.ReadFromServer(ProxyClient)
		if err!=nil{
			goto quit
		}
		ProxyClient.ProxyUser.Write(buf)
	}
quit:
	ProxyClient.ProxyUser.Close()
	ProxyClient.Remote.Close()
}
func SocksTCPBridgeProxyToRemote(ProxyClient *common.ProxyClientSturct){
	buf:=make([]byte,10240)
	for{
		n,err:=ProxyClient.ProxyUser.Read(buf)
		if err!=nil{
			goto quit
		}
		client.SendPacketToServer(ProxyClient,buf[:n])
	}
quit:
	ProxyClient.ProxyUser.Close()
	ProxyClient.Remote.Close()
}
func handleConnection(frontconn net.Conn) {
	frontaddr := frontconn.RemoteAddr().String()
	if config.GetDebug() {
		log.Println("ACCEPTED frontend", frontaddr)
	}
	defer func() {
		if err := recover(); err != nil {
			if config.GetDebug(){
				log.Println("ERROR frontend", frontaddr, err)
			}
		}
		frontconn.Close()
		if config.GetDebug() {
			log.Println("DISCONNECTED frontend", frontaddr)
		}
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
		log.Println("performUdp error:",err.Error())
		return
	}
	defer func(){
		conn.Close()
	}()

	buf := make([]byte, 10)
	copy(buf, []byte{0x05, 0x00, 0x00, 0x01})
	packNetAddr4(conn.LocalAddr().String(), buf[4:])
	frontconn.Write(buf)
	rd,target,err:=SafeReadBytes(conn)
	if err!=nil{
		return
	}
	reconnect:
	address,headerBuf:=GetAddress(rd)
	ProxyClient,err:=client.NewProxyConn(address,frontconn,false)
	if ProxyClient==nil || err != nil {
		frontconn.Write(errorReplyConnect(0x05))
		return
	}
	ProxyClient.UDPHeader=headerBuf
	ProxyClient.UDPTarget=target
	perBuf:=make([]byte,65507)
	preBufLen,err:=rd.Read(perBuf)
	if err!=nil{
		panic(err)
	}
	client.SendPacketToServer(ProxyClient,perBuf[:preBufLen])
	go SocksUDPBridgeRemoteToProxy(ProxyClient,conn)
	flag,rd:=SocksUDPBridgeProxyToRemote(ProxyClient,conn)
	if flag{
		goto reconnect
	}
}
func GetAddress(rd io.Reader)(string,[]byte){
	headerBuffer:=bytes.NewBuffer(nil)
	pkt:=readBytes(rd,4)
	headerBuffer.Write(pkt)
	if pkt[3]!=0x01 && pkt[3]!=0x03 && pkt[3]!=0x04{
		panic("invalid target addr")
	}
	var address string
	if pkt[3]==0x01{ //ipv4
		buf4 := readBytes(rd, 6)
		headerBuffer.Write(buf4)
		address = fmt.Sprintf("%d.%d.%d.%d:%d", buf4[0], buf4[1],
			buf4[2], buf4[3], int(buf4[4]) * 256 + int(buf4[5]))
	}else if pkt[3]==0x03{ //NAME
		buf4 := readBytes(rd, 1)
		headerBuffer.Write(buf4)
		nmlen := int(buf4[0])  // domain name length
		if nmlen > 253 {
			panic("domain name too long")  // will be recovered
		}

		buf5 := readBytes(rd, nmlen + 2)
		headerBuffer.Write(buf5)
		address = fmt.Sprintf("%s:%d", buf5[0:nmlen],
			int(buf5[nmlen]) * 256 + int(buf5[nmlen+1]))
	}else if pkt[3]==0x04{ //ipv6
		buf4 := readBytes(rd, 18)
		headerBuffer.Write(buf4)
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
	return address,headerBuffer.Bytes()
}
func SafeReadBytes(conn *net.UDPConn) (io.Reader,*net.UDPAddr,error){
	buf:=make([]byte,65507)
	n,addr,err:=conn.ReadFromUDP(buf)
	if err!=nil{
		return nil,nil,errors.New("read error")
	}
	return bytes.NewReader(buf[:n]),addr,nil
}
func SocksUDPBridgeRemoteToProxy(ProxyClient *common.ProxyClientSturct,UdpClient *net.UDPConn){
	defer func(){
		ProxyClient.Remote.Close()
	}()
	for{
		buf,err:=client.ReadFromServer(ProxyClient)
		if err!=nil{
				return
		}
		UdpClient.WriteTo(ProxyClient.UDPHeader,ProxyClient.UDPTarget)
		UdpClient.WriteTo(buf,ProxyClient.UDPTarget)
	}
}
func SocksUDPBridgeProxyToRemote(ProxyClient *common.ProxyClientSturct,UdpClient *net.UDPConn) (bool,*bytes.Buffer){
	defer func(){
		ProxyClient.Remote.Close()
	}()
	buf:=make([]byte,65507)
	for{
		rd,_,err:=SafeReadBytes(UdpClient)
		if err!=nil{
			return false,nil
		}
		_,newHeader:=GetAddress(rd)
		n,err:=rd.Read(buf)
		if err!=nil{
			return false,nil
		}
		if bytes.Compare(ProxyClient.UDPHeader,newHeader)!=0{
			newBuffer:=bytes.NewBuffer(nil)
			newBuffer.Write(newHeader)
			newBuffer.Write(buf[:n])
			return true,newBuffer
		}
		client.SendPacketToServer(ProxyClient,buf[:n])
	}

}
func StartSocks5(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Println("Socks5代理服务端监听失败，原因: ", err)
		return
	}
	log.Println("Socks5代理服务端成功监听于",address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}