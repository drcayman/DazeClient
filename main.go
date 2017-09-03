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

func main(){
	var ShowNetSpeed bool=true
	log.Println("DazeClient V3-201709031")
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
		ShowNetSpeed=false
		goto idle
	}
	//判断是否指定了配置文件
	if path,flag:=args.GetArg("-conf");flag{
		helper.ConfFile=path
	}
	if pac,flag:=args.GetArg("-pac");flag{
		helper.PacFile=pac
	}
	helper.LoadConfig()
	proxy.StartProxy()
	idle:
	go proxy.NetSpeedMonitor(ShowNetSpeed)
	for{
		time.Sleep(time.Second*3600)
	}
}