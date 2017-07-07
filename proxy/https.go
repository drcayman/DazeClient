package proxy

import (
	"strings"
	"bufio"
	"net/http"
	"net"
	common "DazeClient/mystruct"
	"bytes"
	"DazeClient/client"
	"DazeClient/mylog"
	"DazeClient/config"
)

func SendPacketToProxyUser(ProxyUser net.Conn,data []byte){
	ProxyUser.Write(data)
}
func LocalHttpsProxyHandle(ProxyUser net.Conn,preBuf []byte){
		flag:=0
		n:=0
		var buf []byte
		defer func() {
			if flag==0{
				ProxyUser.Close()
			}
		}()
		if preBuf!=nil{
			buf=preBuf
			n=len(preBuf)
		}else{
			newbuf:=make([]byte,10240)
			newlen,ReadErr:=ProxyUser.Read(newbuf)
			if ReadErr!=nil{
				return
			}
			buf=newbuf
			n=newlen
		}
		r:=bufio.NewReader(bytes.NewReader(buf[:n]))
		rq,ReadRequestErr:=http.ReadRequest(r)
		if ReadRequestErr!=nil{
			return
		}
		address:=rq.Host
		if strings.Index(address,":") ==-1{
			address+=":80"
		}
		mylog.DPrintln("建立代理连接到",address)
		ProxyClient,newErr:=client.NewTCPProxyConn(address,ProxyUser)
		if newErr!=nil{
			return
		}
		if rq.Method=="CONNECT"{
			//proxyRecv<-[]byte("HTTP/1.1 200 Connection Established\r\n\r\n")
			SendPacketToProxyUser(ProxyUser,[]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
			ProxyClient.Protocol="https"
		}
		if rq.Method!="CONNECT"{
			newbuf:=bytes.Replace(buf[:n],[]byte(" http://"+rq.Host),[]byte(" "),1)
			ProxyClient.Protocol="http"
			client.SendPacketToServer(ProxyClient,newbuf)
		}
		go HTTPSBridgeProxyToRemote(ProxyClient)
		go HTTPSBridgeRemoteToProxy(ProxyClient)
		flag=1
}
func StartHttpsProxy(address string){
	l,ListenErr:=net.Listen("tcp",address)
	if ListenErr!=nil{
		mylog.Println("HTTP(s)代理服务端监听失败，原因：",ListenErr.Error())
		return
	}
	mylog.Println("HTTP(s)代理服务端成功监听于",address)
	config.SetSystemProxy()
	for {
		conn,_:=l.Accept()
		go LocalHttpsProxyHandle(conn,nil)
	}
}
func HTTPSBridgeRemoteToProxy(ProxyClient *common.ProxyClientSturct){
	for {
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

func HTTPSBridgeProxyToRemote(ProxyClient *common.ProxyClientSturct){
	buf:=make([]byte,65536)
	for {
		n,err:=ProxyClient.ProxyUser.Read(buf)
		if err!=nil{
			goto quit
		}
		if ProxyClient.Protocol=="http" {
			r:=bufio.NewReader(bytes.NewReader(buf[:n]))
			_,ReadRequestErr:=http.ReadRequest(r)
			if ReadRequestErr==nil{
				go LocalHttpsProxyHandle(ProxyClient.ProxyUser,buf[:n])
				return
			}
		}
		client.SendPacketToServer(ProxyClient,buf[:n])
	}
	quit:
		ProxyClient.ProxyUser.Close()
		ProxyClient.Remote.Close()

}
