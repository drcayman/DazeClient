package control

import (
	"net"
	"os"
	"bufio"
	"strings"
	"log"
	"github.com/crabkun/DazeClient/common"
	"github.com/crabkun/DazeClient/proxy"
	"encoding/json"
	"github.com/crabkun/DazeClient/helper"
	"fmt"
	"github.com/crabkun/DazeClient/dns"
	"github.com/crabkun/DazeClient/encryption"
	"github.com/crabkun/DazeClient/obscure"
)
var conn net.Conn
func StartControlServer(address string){
	var err error
	conn, err = net.Dial("tcp", address)
	if err != nil {
		println("控制端口连接失败，原因：", err.Error())
		os.Exit(-2)
	}
	for{
		reader := bufio.NewReader(conn)
		buf, err := reader.ReadString('\n')
		if err != nil {
			//服务端断开了连接
			os.Exit(-5)
		}
		buf=strings.Replace(buf,"\n"," ",-1)
		str:=bufio.NewReader(strings.NewReader(buf))
		command,err:=str.ReadString(' ')
		if err != nil{
			continue
		}
		command=strings.TrimSpace(command)
		switch command {
		case "SPEED":RET(SPEED())
		case "DNS":RET(DNS())
		case "LOG":RET(LOG(str))
		case "DEBUG":RET(DEBUG(str))
		case "GET":RET(GET(str))
		case "SET":RET(SET(str))
		default:
			RET("UNKNOWN")
		}
	}
}
func RET(msg string){
	if msg!=""{
		fmt.Fprintln(conn,msg)
	}
}
func SPEED() string {
	return fmt.Sprintf("%d/%d", proxy.LastUpload, proxy.LastDownload)
}
func DNS() string {
	if !dns.DNSOpenFlag {
		if dns.StartDnsServer() {
			return "OK"
		} else {
			return "FAILED"
		}
	} else {
		return"OK"
	}
}
func LOG(args *bufio.Reader) string{
	arg,_:=args.ReadString(' ')
	arg=strings.TrimSpace(arg)
	switch arg {
	case "ON":
		log.Println("日志已重定向到", conn.RemoteAddr())
		log.SetOutput(conn)
		return ""
	case "OFF":
		log.SetOutput(os.Stdout)
		log.Println("日志已重定向到标准输出")
		return ""
	default:
		return "UNKNOWN"
	}
}
func DEBUG(args *bufio.Reader) string{
	arg,_:=args.ReadString(' ')
	arg=strings.TrimSpace(arg)
	switch arg {
	case "ON":
		common.SrvConf.Debug = true
		log.Println("调试模式已开启")
		return ""
	case "OFF":
		common.SrvConf.Debug = false
		log.Println("调试模式已关闭")
		return ""
	default:
		return "UNKNOWN"
	}
}
func GET(args *bufio.Reader) string{
	arg,_:=args.ReadString(' ')
	arg=strings.TrimSpace(arg)
	switch arg {
	case "ENCRYPTION":
		return strings.Join(encryption.GetEncryptionList(),"|")
	case "OBSCURE":
		return strings.Join(obscure.GetObscureList(),"|")
	default:
		return "UNKNOWN"
	}
}
func SET(args *bufio.Reader) string {
	arg1,_:=args.ReadString(' ')
	arg1=strings.TrimSpace(arg1)
	arg2buf,err:=args.Peek(args.Buffered())
	if err!=nil{
		return "UNKNOWN"
	}
	arg2:=string(arg2buf)
	arg2=strings.TrimSpace(arg2)
	switch arg1 {
	case "PAC":
		helper.PacFile = arg2
		log.Println("PAC文件已设置为",arg2)
		return "OK"
	case "PORT":
		common.SrvConf.LocalPort = arg2
		if proxy.RestartServer() == nil {
			return "OK"
		} else {
			return "FAILED"
		}
	case "SERVER":
		newcfg := new(common.S_proxy)
		if err:=json.Unmarshal(arg2buf, newcfg) ;err!= nil {
			log.Println("设置新配置失败：\n内容：",arg2,"\n错误原因：",err)
			return "FAILED"
		} else {
			newcfg.LocalPort = common.SrvConf.LocalPort
			newcfg.Debug = common.SrvConf.Debug
			common.SrvConf = newcfg
			helper.ShowConfig()
			return "OK"
		}
	default:
		return "UNKNOWN"
	}
}

