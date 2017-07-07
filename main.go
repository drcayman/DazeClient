package main

import (
	"DazeClient/proxy"
	"fmt"
	"DazeClient/config"
	"DazeClient/console"
)

func main(){
	fmt.Println("DazeClient V1.0 Author:螃蟹")
	config.Load()
	if config.GlobaConfig.HTTPProxyPort!=""{
		go proxy.StartHttpsProxy(":"+config.GlobaConfig.HTTPProxyPort)
	}
	if config.GlobaConfig.Socks5Port!=""{
		go proxy.StartSocks5(":"+config.GlobaConfig.Socks5Port)
	}
	console.Start()
}
