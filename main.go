package main

import (
	"time"
	"github.com/crabkun/DazeClient/proxy"
	"github.com/crabkun/go-args"
	"os"
	"github.com/crabkun/DazeClient/control"
	"github.com/crabkun/DazeClient/helper"

	"log"
)

//TODO
//加密方式
//完善控制部分
func main(){
	log.Println("DazeClient V3-201708301")
	args:=go_args.ReadArgs()
	//判断是否开启被控模式
	ControlPort,ControlFlag:=args.GetArg("-control-port")
	if ControlFlag{
		ControlPass,flag:=args.GetArg("-control-password")
		if !flag || ControlPass==""{
			println("指定了控制端口但未指定控制密码！")
			os.Exit(-1)
		}
		go control.StartControlServer(ControlPort,ControlPass)
		goto idle
	}
	//判断是否指定了配置文件
	if path,flag:=args.GetArg("-conf");flag{
		helper.ConfFile=path
	}
	helper.LoadConfig()
	go proxy.StartProxy()
	idle:
	for{
		time.Sleep(time.Second*10)
	}
}