package proxy

import (
	"fmt"
	"os"
	"strings"
	"bufio"
	"net/http"
	"net"
	common "DazeClient/mystruct"
	"bytes"
	"DazeClient/client"
)

func SendPacketToProxyUser(ProxyUser net.Conn,data []byte){
	ProxyUser.Write(data)
}
func LocalHttpsProxyHandle(ProxyUser net.Conn){
	var ProxyClientGloba *common.ProxyClientSturct
	defer func(){
		ProxyUser.Close()
		if ProxyClientGloba!=nil{
			ProxyClientGloba.Remote.Close()
			close(ProxyClientGloba.ProxyBrideChan)
		}
	}()
	IsConnected:=false
	P:=""
	defer func(){
		ProxyUser.Close()
	}()
	buf:=make([]byte,5000)
	for{
		n,ReadErr:=ProxyUser.Read(buf)
		if ReadErr!=nil{
			return
		}
		if IsConnected{
			if P=="http"{
				r:=bufio.NewReader(bytes.NewReader(buf[:n]))
				_,ReadRequestErr:=http.ReadRequest(r)
				if ReadRequestErr==nil{
					goto reconnect
				}
			}
			//fmt.Println("浏览器准备发送",len(buf[:n]))
			//(*RemoteConn).Write(buf[:n])
			//client.SendPacketToServer(ProxyClientGloba,buf[:n])
			ProxyClientGloba.ProxyBrideChan<-common.Packet{Command:0,Data:buf[:n]}
			continue
		}
	reconnect:
		r:=bufio.NewReader(bytes.NewReader(buf[:n]))
		rq,ReadRequestErr:=http.ReadRequest(r)
		if ReadRequestErr!=nil{
			return
		}
		address:=rq.Host
		if strings.Index(address,":") ==-1{
			address+=":80"
		}
		fmt.Println("建立连接到",address)
		ProxyClientGlobatmp,newErr:=client.NewTCPProxyConn(address,ProxyUser)
		if newErr!=nil{
			return
		}
		ProxyClientGlobatmp.Locker.Lock()
		ProxyClientGloba=ProxyClientGlobatmp
		//go ProxyRemoteHandle(proxyRecv,c)
		go HTTPSBridge(ProxyClientGloba)
		if rq.Method=="CONNECT"{
			//proxyRecv<-[]byte("HTTP/1.1 200 Connection Established\r\n\r\n")
			SendPacketToProxyUser(ProxyUser,[]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
			P="https"
		}
		if rq.Method!="CONNECT"{
			newbuf:=bytes.Replace(buf[:n],[]byte(" http://"+rq.Host),[]byte(" "),1)
			P="http"
			client.SendPacketToServer(ProxyClientGloba,newbuf)
		}
		IsConnected=true

	}
}
func StartHttpsProxy(){
	l,ListenErr:=net.Listen("tcp",":10800")
	if ListenErr!=nil{
		fmt.Println(ListenErr.Error())
		os.Exit(-1)
	}
	for {
		conn,_:=l.Accept()
		go LocalHttpsProxyHandle(conn)
	}
}
func HTTPSBridge(ProxyClient *common.ProxyClientSturct){
	for packet:=range ProxyClient.ProxyBrideChan{
		//0代表代理发送给远端
		//1代表远端发送给代理
		switch packet.Command{
		case 0:client.SendPacketToServer(ProxyClient,packet.Data)
		case 1://ProxyClient.ProxyRecvChan<-packet
			ProxyClient.ProxyUser.Write(packet.Data)
		}
	}
}
