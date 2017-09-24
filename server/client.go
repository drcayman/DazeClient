package server

import(
	"github.com/crabkun/DazeClient/common"
	"github.com/crabkun/DazeClient/helper"
	"fmt"
	"time"
	"net"
	"encoding/json"
	"github.com/crabkun/DazeClient/util"
	"github.com/crabkun/DazeClient/obscure"
	"github.com/crabkun/DazeClient/encryption"
	"log"
	"bytes"
	"io"
	"strings"
)
type S_Client struct {
	//代理用户的套接字
	ProxyUserConn net.Conn
	//代理目标服务器套接字
	RemoteServerConn net.Conn
	RemoteServeFlag bool

	//是否已连接
	Connected bool
	//代理目标和网络协议
	Network string
	TargetHost string
	TargetHostRealAddr string

	//加密与伪装的接口
	Ob obscure.ObscureAction
	E encryption.EncryptionAction

	UDPAddr *net.UDPAddr
	LocalDisconnect bool
	TryLogin bool
	//配置
	*common.S_proxy

}
func (client *S_Client)Decode(data []byte) []byte{
	buf,err:=client.E.Decrypt(data)
	if err!=nil{
		panic(err.Error())
	}
	return buf
}
func (client *S_Client)Encode(data []byte) []byte{
	buf,err:=client.E.Encrypt(data)
	if err!=nil{
		panic(err.Error())
	}
	return buf
}
func (client *S_Client)Disconnect(){
	if client.Connected{
		client.RemoteServerConn.Close()
		client.Connected=false
	}
	client.ProxyUserConn.Close()

}
func (client *S_Client)Read() []byte {
	//读取头部
	headerEncoded:=client.SafeRead(client.RemoteServerConn,4)
	//解码头部
	header:=client.Decode(headerEncoded)
	if header[0]!=0xF1 && header[3]!=0xF2{
		panic("头部不匹配，可能是伪装或者加密方式和参数不正确")
	}
	//读取负载
	length:=int(header[1])+int(header[2])*256
	if length==0{
		panic("长度错误")
	}
	//解码负载
	bodyEncoded:=client.SafeRead(client.RemoteServerConn,length)
	return client.Decode(bodyEncoded)
}
func (client *S_Client)SafeRead(conn net.Conn,length int) ([]byte) {
	buf:=make([]byte,length)
	_,err:=io.ReadFull(conn,buf)
	if err!=nil {
		client.PanicError()
	}
	return buf
}
func (client *S_Client)Write(data []byte)(n int, err error){
	length:=len(data)
	if data==nil || length==0 || length>65535{
		panic("数据长度不正确(1-65535)")
	}
	header:=[]byte{0xF1,byte(length%0x100),byte(length/0x100),0xF2}
	buffer:=bytes.NewBuffer(client.Encode(header))
	buffer.Write(client.Encode(data))
	client.SafeSend(buffer.Bytes(),client.RemoteServerConn)
	return length,nil
}
func (client *S_Client)SafeSend(data []byte,conn net.Conn){
	length:=len(data)
	for pos:=0;pos<length;{
		n,err:=conn.Write(data[pos:])
		if err!=nil {
			client.PanicError()
		}
		pos+=n
	}
}
func (client *S_Client)PanicError(){
	switch {
	case client.Connected:panic(nil)
	case client.LocalDisconnect:panic(nil)
	case client.TryLogin:panic("服务器连接目标超时")
	default:
		panic("服务器在握手期间断开了连接")
	}
}
func (client *S_Client)Login(){
	var err error
	if client.Network!="tcp" && client.Network!="udp"{
		panic("网络协议有误")
	}
	//开始登录
	authinfo:=common.Json_Auth{
		Username:client.Username,
		Password:util.GetDoubleMd5(client.Password),
		Net:client.Network,
		Host:client.TargetHost,
		Spam:util.GetRandomString(256),
	}
	authinfoBuf,err:=json.Marshal(authinfo)
	if err!=nil{
		panic("生成登录数据失败"+err.Error())
	}
	client.Write(authinfoBuf)
	client.TryLogin=true
	//读取返回结果
	authret:=new(common.Json_Ret)
	authretBuf:=client.Read()
	err=json.Unmarshal(authretBuf,authret)
	if err!=nil{
		panic("解析登录回执失败"+err.Error())
	}
	//解析返回结果
	//-1 服务器无法解析客户端的登录数据
	//-2 网络协议错误
	//-3 IP地址错误
	//-4 代理服务器无法连接指定地址
	//1 登录成功
	switch authret.Code {
	case -1:
		panic("服务器无法解析客户端的登录数据")
	case -2:
		panic("网络协议错误")
	case -3:
		panic("IP地址错误")
	case -4	:
		panic("代理服务器无法连接指定地址")
	case -5:
		panic("登录服务器失败："+authret.Data)
	case 1:
		client.Connected=true
		//验证成功，去除验证超时
		client.RemoteServerConn.SetDeadline(time.Time{})
		client.TargetHostRealAddr=authret.Data
		helper.DebugPrintln(fmt.Sprintf("调试信息：目标([%s]%s)代理建立成功",client.Network,client.TargetHost))
	default:
		panic("服务器返回的数据无法识别")
	}
}
func PackNewUser(l net.Conn,r net.Conn,s *common.S_proxy) *S_Client{
	return &S_Client{
		ProxyUserConn:l,
		RemoteServerConn:r,
		S_proxy:s,
	}
}
func CallProxyServer(ProxyUser net.Conn,cfg *common.S_proxy,host string,network string) *S_Client {
	var client *S_Client
	defer func(){
		if err := recover(); err != nil{
			log.Println(fmt.Sprintf("目标([%s]%s)代理提前结束，原因(%s)",network,host,err))
			if client.RemoteServeFlag{
				client.RemoteServerConn.Close()
			}
		}
	}()
	if strings.Index(host,"127.0.0.1:")!=-1{
		return nil
	}
	helper.DebugPrintln(fmt.Sprintf("调试信息：目标([%s]%s)代理开始",network,host))
	//初始化client结构
	client=PackNewUser(ProxyUser,nil,cfg)
	if cfg.Address==""{
		panic("连接配置不存在！")
	}
	client.TargetHost=host
	client.Network=network
	//加载加密模块
	E,ExistFlag:=encryption.GetEncryption(client.Encryption)
	if !ExistFlag{
		panic("加密方式"+client.Encryption+"不存在")
	}
	client.E=E

	//加载伪装模块
	Ob,ExistFlag:=obscure.GetObscure(client.Obscure)
	if !ExistFlag{
		panic("伪装方式"+client.Obscure+"不存在")
	}
	client.Ob=Ob

	//连接代理服务器
	r,err:=net.Dial("tcp",cfg.Address+":"+cfg.Port)
	if err!=nil{
		panic("代理服务器"+cfg.Address+":"+cfg.Port+"连接失败！")
	}
	client.RemoteServeFlag=true
	client.RemoteServerConn=r

	//设置验证超时时间
	client.RemoteServerConn.SetDeadline(time.Now().Add(time.Second*10))
	//开始伪装
	obErr:=client.Ob.Action(client.RemoteServerConn,client.ObscureParam)
	if obErr!=nil{
		panic("伪装时出现错误："+obErr.Error())
	}
	//为初始化加密方式
	eErr:=client.E.InitUser(client.RemoteServerConn,client.EncryptionParam)
	if eErr!=nil{
		panic("为用户初始化加密方式时出现错误："+eErr.Error())
	}
	client.Login()
	return client
}