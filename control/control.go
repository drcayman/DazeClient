package control

import (
	"net"
	"os"
	"bufio"
	"strings"
	"log"
	"github.com/crabkun/DazeClient/common"
)
//单线程，只接受1个客户端

func StartControlServer(port string,password string){
	l,err:=net.Listen("tcp","127.0.0.1:"+port)
	if err!=nil{
		println("控制端口监听失败，原因：",err)
		os.Exit(-2)
	}
	for{
		conn,err:=l.Accept()
		if err!=nil{
			continue
		}
		auth:=false
		ret:=""
		for{
			reader:=bufio.NewReader(conn)
			str,err:=reader.ReadString('\n')
			if err!=nil{
				log.SetOutput(os.Stdout)
				log.Println("日志已重定向到标准输出")
				conn.Close()
				break
			}
			str=str[:len(str)-2]
			arr:=strings.Split(str," ")
			switch arr[0] {
				case "AUTH":
					if len(arr)!=2{
						goto UNKNOWN
					}
					if password==arr[1]{
						auth=true
						ret="AUTHOK"
						goto RET
					}else{
						auth=false
						ret="AUTHERR"
						goto RET
					}
				case "LOG":
					if !auth{
						goto UNAUTH
					}
					if len(arr)!=2{
						goto UNKNOWN
					}
					if arr[1]=="ON"{
						log.Println("日志已重定向到",conn.RemoteAddr())
						log.SetOutput(conn)
					}else{
						log.Println("日志已重定向到标准输出")
						log.SetOutput(os.Stdout)
					}
					ret="OK"
					goto RET
				case "DEBUG":
					if !auth{
						goto UNAUTH
					}
					if len(arr)!=2{
						goto UNKNOWN
					}
					if arr[1]=="ON"{
						common.SrvConf.Debug=true
					}else{
						common.SrvConf.Debug=false
					}
					ret="OK"
					goto RET
				case "SET":
					if !auth{
						goto UNAUTH
					}
					if len(arr)!=3{
						goto UNKNOWN
					}
					switch arr[1] {
					case "HTTP":
					case "SOCKS5":
					case "SERVER":
					}
			default:
				goto UNKNOWN
			}
			continue
		UNKNOWN:
			conn.Write([]byte("UNKNOWN\n"))
			continue
		RET:
			conn.Write([]byte(ret+"\n"))
			continue
		UNAUTH:
			conn.Write([]byte("UNAUTH\n"))
			continue
		}

	}
}
