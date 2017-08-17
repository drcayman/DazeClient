package main

import (
	"DazeClient/proxy"
	"fmt"
	"DazeClient/config"
	"DazeClient/console"
	"github.com/crabkun/go-args"
	"time"
)

func main(){
	m:=go_args.ReadArgs()
	fmt.Println("DazeClient V2.0-2017081701 Author:螃蟹")
	conf,flag:=m.GetArg("-conf")
	if flag{
		config.LoadConfFile(conf)
	}else{
		config.Load()
	}
	_,noConsoleFlag:=m.GetArg("-noconsole")
	if !noConsoleFlag{
		go console.Start()
	}

	if config.GlobalConfig.HTTPProxyPort!=""{
		go proxy.StartHttpsProxy(":"+config.GlobalConfig.HTTPProxyPort)
	}
	if config.GlobalConfig.Socks5Port!=""{
		go proxy.StartSocks5(":"+config.GlobalConfig.Socks5Port)
	}
	for{
		time.Sleep(time.Second*3600)
	}

}
