package client

import (
	"fmt"
	"DazeClient/util"
	"errors"
	"encoding/binary"
	"bytes"
	common "DazeClient/mystruct"
	"net"
	"DazeClient/config"
)
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
	binary.Write(ContentLenBuffer,binary.BigEndian,ContentLen)
	copy(buf[2:],ContentLenBuffer.Bytes())
	copy(buf[4:],content)
	buf[len(buf)-1]=0xF2
	return buf
}
func DePacket(buf []byte) (byte,[]byte,error){
	if len(buf)<6 || buf[0]!=0xF1 || buf[len(buf)-1]!=0xF2{
		return 0,nil,errors.New("error1")
	}
	ContentLen:=int(buf[2])*256+int(buf[3])
	if len(buf)-5!=int(ContentLen){
		return 0,nil,errors.New("error2")
	}
	return buf[1],buf[4:4+ContentLen],nil
}
func ReadFromServer(ProxyClient *common.ProxyClientSturct) ([]byte,error){
	headerbuf:=make([]byte,4)
	n,err:=ProxyClient.Remote.Read(headerbuf)
	if n<4 ||err!=nil{
		return nil,errors.New("read header error")
	}
	AESKey:=ProxyClient.AESKey
	if AESKey==nil{
		AESKey=util.GetAESKeyByDay()
	}
	header:=headerbuf[:4]
	headerDecode,_:=util.DecryptAES(header,AESKey)
	if headerDecode[0]!=0xFB || headerDecode[3]!=0xFC{
		return nil,errors.New("deheader error")
	}
	buflen:=int(headerDecode[1])+int(headerDecode[2])*256
	buf:=make([]byte,buflen)
	pos:=0
	for{
		n,err:=ProxyClient.Remote.Read(buf[pos:])
		if err!=nil{
			return nil,errors.New("read body error")
		}
		buflen-=n
		pos+=n
		if buflen<0{
			return nil,errors.New("body len error")
		}
		if buflen==0{
			break
		}
	}
	buf,_=util.DecryptAES(buf,AESKey)
	return buf,nil
}
func KeyExchange(ProxyClient *common.ProxyClientSturct,PublicKeyDer []byte){
	NewAESKey:=util.GenAESKey(32)
	SendPacketToServer(ProxyClient,util.EncryptRSAWithDer(append(NewAESKey,0xFF),PublicKeyDer))
	ProxyClient.AESKey=NewAESKey
	ProxyClient.IsKeyExchange=true
}
func SendPacketToServer(ProxyClient *common.ProxyClientSturct,data []byte) (int,error){
	//fmt.Println(util.B2s(data))
	AESKey:=ProxyClient.AESKey
	if AESKey==nil{
		AESKey=util.GetAESKeyByDay()
	}
	var bufffer *bytes.Buffer
	var header []byte
	data,_=util.EncryptAES(data,AESKey)
	datelen:=len(data)
	header,_=util.EncryptAES([]byte{0xFB,byte(datelen%0x100),byte(datelen/0x100),0xFC},AESKey)
	bufffer=bytes.NewBuffer(header)
	bufffer.Write(data)
	n,err:=ProxyClient.Remote.Write(bufffer.Bytes())
	//fmt.Println("一共发送了",n)

	return n,err
}
func ReadFromServerThread(ProxyClient *common.ProxyClientSturct,PacketChan chan common.Packet){
	defer func(){
		close(PacketChan)
		ProxyClient.Remote.Close()
		ProxyClient.ProxyUser.Close()
	}()
	for {
		buf, ReadErr := ReadFromServer(ProxyClient)
		if ReadErr != nil {
			//fmt.Println("读取错误,连接断开")
			return
		}
		if ProxyClient.IsKeyExchange == false {
			KeyExchange(ProxyClient,buf)
			continue
		}
		if ProxyClient.IsConnected{
			//ProxyClient.ProxyUser.Write(buf)
			ProxyClient.ProxyBrideChan<-common.Packet{Command:1,Data:buf}
			continue
		}
		command, data, DePacketErr := DePacket(buf)
		if DePacketErr != nil {
			fmt.Println("解码错误",DePacketErr.Error())
			return
		}
		PacketChan <- common.Packet{Command: command, Data: data}
	}
}
func ServeCommand(ProxyClient *common.ProxyClientSturct,PacketChan chan common.Packet,address string,IsUDP bool){
	defer func(){
		ProxyClient.Remote.Close()
	}()
	for packet:=range PacketChan{
		switch packet.Command {
		case 0x04:
			//fmt.Println("key交换成功")
			command:=byte(0xA1)
			if IsUDP{
				command=byte(0xA2)
			}
			SendPacketToServer(ProxyClient,MakePacket(command,[]byte(address)))
		case 0xC1:
			ProxyClient.IsConnected=true
			//fmt.Println("远程连接成功了哦")
			ProxyClient.Locker.Unlock()
		case 0xE1:
			fmt.Println(address,"远程无法解析")
			return
		case 0xE2:
			fmt.Println(address,"连接失败")
			return
		}
	}
}
func NewTCPProxyConn(address string,ProxyUser net.Conn) (*common.ProxyClientSturct,error){
	ServerConn,err:=net.Dial("tcp",config.GetServerIP())
	if err!=nil{
		fmt.Println("代理服务器",config.GetServerIP(),"连接建立失败！！！")
		return nil,err
	}
	PacketChan:=make(chan common.Packet,10)
	ProxyClient:=common.ProxyClientSturct{Remote:ServerConn,ProxyUser:ProxyUser,ProxyBrideChan:make(chan common.Packet,10)}
	ProxyClient.Locker.Lock()
	go ReadFromServerThread(&ProxyClient,PacketChan)
	go ServeCommand(&ProxyClient,PacketChan,address,false)
	return &ProxyClient,nil
}
func NewUDPProxyConn(address string,ProxyUser net.Conn) (*common.ProxyClientSturct,error){
	ServerConn,err:=net.Dial("tcp",config.GetServerIP())
	if err!=nil{
		return nil,err
	}
	PacketChan:=make(chan common.Packet,10)
	ProxyClient:=common.ProxyClientSturct{Remote:ServerConn,ProxyUser:ProxyUser,ProxyBrideChan:make(chan common.Packet,10)}
	ProxyClient.Locker.Lock()
	go ReadFromServerThread(&ProxyClient,PacketChan)
	go ServeCommand(&ProxyClient,PacketChan,address,true)
	return &ProxyClient,nil
}