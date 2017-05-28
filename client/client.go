package client

import (
	"DazeClient/mylog"
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
	return n,err
}
func CallProxyServer(ProxyClient *common.ProxyClientSturct) (error) {
	defer func(){
		if !ProxyClient.IsConnected{
			ProxyClient.Remote.Close()
		}
	}()
	for {
		buf, ReadErr := ReadFromServer(ProxyClient)
		if ReadErr != nil {
			//fmt.Println("读取错误,连接断开")
			return errors.New("conn proxy server error")
		}
		if ProxyClient.IsKeyExchange == false {
			KeyExchange(ProxyClient,buf)
			continue
		}
		command, data, DePacketErr := DePacket(buf)
		if DePacketErr != nil {
			mylog.DPrintln("解码错误",DePacketErr.Error())
			return errors.New("decode error")
		}
		switch command {
		case 0x04:
			//fmt.Println("key交换成功")
			command:=byte(0xA1)
			if ProxyClient.IsUDP{
				command=byte(0xA2)
			}
			SendPacketToServer(ProxyClient,MakePacket(command,[]byte(ProxyClient.Address)))
		case 0xC1:
			ProxyClient.IsConnected=true
			ProxyClient.RemoteRealAddr=util.B2s(data)
			return nil
		case 0xE1:
			mylog.DPrintln(ProxyClient.Address,"远程无法解析")
			return errors.New("ip error")
		case 0xE2:
			mylog.DPrintln(ProxyClient.Address,"连接失败")
			return errors.New("conn remote error")
		}
	}
}

func NewTCPProxyConn(address string,ProxyUser net.Conn) (*common.ProxyClientSturct,error){
	ServerConn,err:=net.Dial("tcp",config.GetServerIP())
	if err!=nil{
		mylog.DPrintln("代理服务器",config.GetServerIP(),"连接建立失败！！！")
		return nil,err
	}
	ProxyClient:=common.ProxyClientSturct{Remote:ServerConn,ProxyUser:ProxyUser,Address:address}
	CallProxyServerErr:=CallProxyServer(&ProxyClient)
	return &ProxyClient,CallProxyServerErr
}
func NewUDPProxyConn(address string,ProxyUser net.Conn) (*common.ProxyClientSturct,error){
	ServerConn,err:=net.Dial("tcp",config.GetServerIP())
	if err!=nil{
		mylog.DPrintln("代理服务器",config.GetServerIP(),"连接建立失败！！！")
		return nil,err
	}
	ProxyClient:=common.ProxyClientSturct{Remote:ServerConn,ProxyUser:ProxyUser,Address:address,IsUDP:true}
	CallProxyServerErr:=CallProxyServer(&ProxyClient)
	return &ProxyClient,CallProxyServerErr
}