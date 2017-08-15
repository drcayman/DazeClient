package client

import (
	"DazeClient/util"
	"errors"
	"encoding/binary"
	"bytes"
	"DazeClient/common"
	"net"
	"DazeClient/config"
	"log"
	"DazeClient/encryption"
	"DazeClient/disguise"
	"encoding/json"
)
//生成控制数据包
//[头部：F1][命令][内容长度][内容][尾部：F2]
//头部尾部均为1字节
//命令的长度为1字节
//内容长度的长度为2字节
//内容的长度不限
func MakePacket(command byte,content []byte) []byte{
	if content==nil{
		content=[]byte{0x0}
	}
	ContentLen:=uint16(len(content))
	if ContentLen>0xFFFF {
		return nil
	}
	ContentLenBuffer:=bytes.NewBuffer([]byte{})
	buf:=make([]byte,5+len(content))
	buf[0]=0xF1
	buf[1]=command
	binary.Write(ContentLenBuffer,binary.LittleEndian,ContentLen)
	copy(buf[2:],ContentLenBuffer.Bytes())
	copy(buf[4:],content)
	buf[len(buf)-1]=0xF2
	return buf
}

//解析控制数据包，解析看上面
func DePacket(buf []byte) (byte,[]byte,error){
	if len(buf)<6 || buf[0]!=0xF1 || buf[len(buf)-1]!=0xF2{
		return 0,nil,errors.New("error1")
	}
	ContentLen:=int(buf[2])+int(buf[3])*256
	if len(buf)-5!=int(ContentLen){
		return 0,nil,errors.New("error2")
	}
	return buf[1],buf[4:4+ContentLen],nil
}


//打包并发送数据包给服务端
//[头部][内容]
//头部和内容分开加密
//头部为4字节,[FB][内容长度][FC]
//内容无限长
func SendPacketToServer(ProxyClient *common.ProxyClientSturct,data []byte){
	dataEncoded,dataEncodedErr:=ProxyClient.Encryption.Encrypt(&ProxyClient.EncReserved,data)
	if dataEncodedErr!=nil{
		return
	}
	for _,pkt:=range dataEncoded{
		datalen:=len(pkt)
		header:=[]byte{0xFB,byte(datalen%0x100),byte(datalen/0x100),0xFC}
		headerEncoded,headerEncodedErr:=ProxyClient.Encryption.Encrypt(&ProxyClient.EncReserved,header)
		if headerEncodedErr!=nil{
			return
		}
		bufffer:=bytes.NewBuffer(headerEncoded[0])
		bufffer.Write(pkt)
		_,err:=ProxyClient.Remote.Write(bufffer.Bytes())
		if err!=nil{
			return
		}
	}
}

//解析服务端发过来的数据包，解析看上面
func ReadFromServer(ProxyClient *common.ProxyClientSturct) ([]byte,error){
	HeaderBuf:=make([]byte,4)
	n,err:=ProxyClient.Remote.Read(HeaderBuf)
	if n<4 ||err!=nil{
		return nil,errors.New("read header error ")
	}
	headerDecode,err:=ProxyClient.Encryption.Decrypt(&ProxyClient.EncReserved,HeaderBuf)
	if err!=nil || headerDecode[0]!=0xFB || headerDecode[3]!=0xFC{
		return nil,errors.New("decode header error")
	}
	PacketLen:=int(headerDecode[1])+int(headerDecode[2])*256
	buf:=make([]byte,PacketLen)
	pos:=0
	for{
		n,err:=ProxyClient.Remote.Read(buf[pos:])
		if err!=nil{
			return nil,errors.New("read body error")
		}
		PacketLen-=n
		pos+=n
		if PacketLen<0{
			return nil,errors.New("body len error")
		}
		if PacketLen==0{
			break
		}
	}
	//copy(buf,HeaderBuf)

	DecodeBuf,err:=ProxyClient.Encryption.Decrypt(&ProxyClient.EncReserved,buf)
	if err!=nil{
		return nil,errors.New("decode body error")
	}
	return DecodeBuf,nil
}

//呼叫代理服务器并开始代理
func CallProxyServer(ProxyClient *common.ProxyClientSturct) (error) {
	host,port,err:=net.SplitHostPort(ProxyClient.Address)
	if err!=nil{
		panic("目标代理IP解析失败")
	}
	netType:="tcp"
	if ProxyClient.IsUDP{
		netType="udp"
	}
	authinfo:=common.JsonAuth{
		Username:config.GetUsername(),
		Password:config.GetPassword(),
		Host:host,
		Port:port,
		Net:netType,
	}
	authinfoBuf,_:=json.Marshal(authinfo)
	SendPacketToServer(ProxyClient,MakePacket(0x02,authinfoBuf))
	for {
		buf, ReadErr := ReadFromServer(ProxyClient)
		if ReadErr != nil {
			panic("数据传输错误！请检查加密方式和参数是否跟服务器一致！")
		}
		command, data, DePacketErr := DePacket(buf)
		if DePacketErr != nil {
			panic("数据解析错误！请检查加密方式和参数是否跟服务器一致！")

		}
		switch command {
		case 0xC1:
			ProxyClient.IsConnected=true
			ProxyClient.RemoteRealAddr=util.B2s(data)
			log.Println(ProxyClient.Address,"代理连接建立成功")
			return nil
		case 0xE1:
			panic("远程无法解析目标IP")
		case 0xE2:
			panic("远程无法连接目标IP")
		case 0xE3:
			panic("用户名或者密码错误")
		}
	}
}

//断开代理用户的连接和代理服务器的连接
func Disconnect(ProxyClient *common.ProxyClientSturct){
	if ProxyClient.Remote!=nil{
		ProxyClient.Remote.Close()
	}
	if ProxyClient.ProxyUser!=nil{
		ProxyClient.ProxyUser.Close()
	}
}

//新代理客户端链接
func NewProxyConn(address string,ProxyUser net.Conn,IsTCP bool) (*common.ProxyClientSturct,error){
	ServerConn,err:=net.Dial("tcp",config.GetServerIP())
	if err!=nil{
		log.Println("代理服务器",config.GetServerIP(),"连接失败！！！")
		return nil,err
	}
	ProxyClient:=common.ProxyClientSturct{
		Remote:ServerConn,
		ProxyUser:ProxyUser,
		Address:address,
		IsUDP:!IsTCP,
	}
	defer func(){
		if err := recover(); err != nil{
				log.Printf("代理服务器%s连接失败！原因：%s",config.GetServerIP(),err)
				Disconnect(&ProxyClient)
			}

	}()
	//加载加密模块和初始化
	EncryptionName,EncryptionParam:=config.GetEncryption()
	enc,encflag:=encryption.GetEncryption(EncryptionName)
	if !encflag{
		panic("加密方式"+EncryptionName+"不存在")
	}
	ProxyClient.Encryption=enc()
	encInitErr:=ProxyClient.Encryption.Init(EncryptionParam,&ProxyClient.EncReserved)
	if encInitErr!=nil{
		panic("加密方式"+EncryptionName+"加载错误！原因："+encInitErr.Error())
	}
	//加载伪装模块和初始化
	DisguiseName,DisguiseParam:=config.GetDisguise()
	dsg,dsgflag:=disguise.GetDisguise(DisguiseName)
	if !dsgflag{
		panic("伪装方式"+DisguiseName+"不存在")
	}
	ProxyClient.Disguise=dsg()
	dsgInitErr:=ProxyClient.Disguise.Init(DisguiseParam,&ProxyClient.DsgReserved)
	if dsgInitErr!=nil{
		panic("伪装方式"+DisguiseName+"加载错误！原因："+dsgInitErr.Error())
	}
	//开始伪装
	dsgErr:=ProxyClient.Disguise.Action(ServerConn,&ProxyClient.DsgReserved)
	if dsgErr!=nil{
		panic("伪装时出现错误："+dsgErr.Error())
	}
	encErr:=ProxyClient.Encryption.InitUser(ProxyClient.Remote,&ProxyClient.EncReserved)
	if encErr!=nil{
		panic("为用户初始化加密方式时出现错误："+encErr.Error())
	}
	CallProxyServerErr:=CallProxyServer(&ProxyClient)
	return &ProxyClient,CallProxyServerErr
}
