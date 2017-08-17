package proxy

import (
	"strings"
	"bufio"
	"net/http"
	"net"
	"DazeClient/common"
	"bytes"
	"DazeClient/client"
	"io/ioutil"
	"DazeClient/config"
	"log"
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
		if rq.URL.Path=="/!daze.pac"{
			b,e:=ioutil.ReadFile("gfwlist.pac")
			if e!=nil{
				return
			}
			SendPacketToProxyUser(ProxyUser,[]byte("HTTP/1.1 200 OK\r\nContent-Type:application/x-ns-proxy-autoconfig\r\n\r\n"))
			SendPacketToProxyUser(ProxyUser,b)
			return
		}else if rq.URL.Path=="/!dazeD.pac"{
			SendPacketToProxyUser(ProxyUser,[]byte("HTTP/1.1 200 OK\r\nContent-Type:application/x-ns-proxy-autoconfig\r\n\r\n"))
			SendPacketToProxyUser(ProxyUser,[]byte("function FindProxyForURL(url, host) {return \"PROXY 127.0.0.1:"+config.GlobalConfig.HTTPProxyPort+"\"}"))
		}
		address:=rq.Host
		if strings.Index(address,":") ==-1{
			address+=":80"
		}
		//log.Println("建立代理连接到",address)
		ProxyClient,newErr:=client.NewProxyConn(address,ProxyUser,true)
		if ProxyClient==nil || newErr!=nil{
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
		log.Println("HTTP(s)代理服务端监听失败，原因：",ListenErr.Error())
		return
	}
	log.Println("HTTP(s)代理服务端成功监听于",address)
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
	buf:=make([]byte,10240)
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
