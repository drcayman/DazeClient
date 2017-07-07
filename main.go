package main

import (
	"DazeClient/proxy"
	"fmt"
	"DazeClient/config"
	"DazeClient/console"
	"os"
	"os/signal"
	"syscall"
)
func catch() {
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		config.ClearSystemProxy()
		os.Exit(0)
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
}
func main(){
	catch()
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
