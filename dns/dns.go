package dns

import (
	"net"
	"log"
	"time"
	"github.com/crabkun/DazeClient/server"
	"github.com/crabkun/DazeClient/common"
	"encoding/json"
)
var DNSOpenFlag bool
func StartDnsServer() bool{
	addr,_:=net.ResolveUDPAddr("udp","127.0.0.1:53")
	dnsConn,err:=net.ListenUDP("udp",addr)
	if err!=nil{
		log.Println("DNS服务端监听失败！",err.Error())
		return false
	}
	log.Println("DNS服务端成功监听在：",addr)
	DNSOpenFlag=true
	go ServerHandle(dnsConn)
	return true
}
func ServerHandle(conn *net.UDPConn){
	for{
		buf:=make([]byte,65507)
		n,addr,err:=conn.ReadFromUDP(buf)
		if err!=nil{
			log.Println("DNS服务端异常关闭！",err.Error())
			return
		}
		go CallProxy(conn,buf[:n],addr)
	}
}
func CallProxy(conn *net.UDPConn,buf []byte,cli *net.UDPAddr){
	defer func(){
		recover()
	}()
	proxyclient:=server.CallProxyServer(conn,common.SrvConf,"","udp")
	if proxyclient==nil{
		return
	}
	var err error
	proxyclient.RemoteServerConn.SetReadDeadline(time.Now().Add(time.Second*5))
	var UDP common.Json_UDP
	UDP.Data=buf
	UDP.Host="8.8.8.8:53"
	jsonBuf,err:=json.Marshal(UDP)
	if err!=nil{
		return
	}
	proxyclient.Write(jsonBuf)
	err =json.Unmarshal(proxyclient.Read(),&UDP)
	if err!=nil{
		return
	}
	conn.WriteToUDP(UDP.Data,cli)
}