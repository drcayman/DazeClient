package helper

import (
	"io/ioutil"
	"os"
	"encoding/json"
	"github.com/crabkun/DazeClient/common"
	"log"
	"bytes"
)

var ConfFile="server.conf"
var PacFile="auto.pac"
var ProxyAllPac []byte
func GenProxyAllPac(){
	ProxyAllPac=[]byte("HTTP/1.1 200 OK\r\nContent-Type:application/x-ns-proxy-autoconfig\r\n\r\nfunction FindProxyForURL(url, host) {return \"PROXY 127.0.0.1:"+
		common.SrvConf.LocalPort+"\"}")
}
func LoadPAC()([]byte,error){
	buf,err:=ioutil.ReadFile(PacFile)
	if err!=nil{
		log.Printf("代理客户端请求了PAC文件，但加载(%s)失败了，原因：%s\n",PacFile,err)
		return nil,err
	}
	buf=bytes.Replace(buf,[]byte("SOCKS5 127.0.0.1:1080"),[]byte("PROXY 127.0.0.1:"+common.SrvConf.LocalPort),1)
	buffer:=bytes.NewBuffer([]byte("HTTP/1.1 200 OK\r\nContent-Type:application/x-ns-proxy-autoconfig\r\n\r\n"))
	buffer.Write(buf)
	return buffer.Bytes(),err
}
func LoadConfig(){
	var err error
	buf,err:=ioutil.ReadFile(ConfFile)
	if err!=nil{
		log.Printf("配置文件(%s)读取错误：%s",ConfFile,err.Error())
		os.Exit(-3)
	}
	err=json.Unmarshal(buf,common.SrvConf)
	if err!=nil{
		log.Println("配置文件格式错误！请严格按照JSON格式来修改",ConfFile,"(",err.Error(),")")
		os.Exit(-4)
	}
	ShowConfig()
}
func ShowConfig(){
	log.Println("配置文件读取成功：")
	log.Println("    服务器IP：",common.SrvConf.Address)
	log.Println("    服务器端口：",common.SrvConf.Port)
	log.Println("    用户名：",common.SrvConf.Username)
	log.Println("    密码：隐藏")
	log.Println("    加密方式：",common.SrvConf.Encryption)
	log.Println("    加密参数：",common.SrvConf.EncryptionParam)
	log.Println("    伪装方式：",common.SrvConf.Obscure)
	log.Println("    伪装参数：",common.SrvConf.ObscureParam)
	log.Println("    本地HTTP代理监听端口：",common.SrvConf.LocalPort)
	log.Println("    本地SOCKS5代理监听端口：",common.SrvConf.LocalPort)
	log.Println("    调试模式：",common.SrvConf.Debug)
}