package main

import (
	"time"
	"github.com/crabkun/DazeClient/proxy"
	"github.com/crabkun/go-args"
	"github.com/crabkun/DazeClient/control"
	"github.com/crabkun/DazeClient/helper"

	"log"
)

func main(){
	var ShowNetSpeed bool=true
	log.Println("DazeClient V3-201709031")
	args:=go_args.ReadArgs()
	//判断是否开启被控模式
	ControlAddress,ControlFlag:=args.GetArg("-control-address")
	if ControlFlag{
		go control.StartControlServer(ControlAddress)
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