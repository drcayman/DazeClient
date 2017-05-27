package main

import (
	"DazeClient/proxy"
	"fmt"
	"DazeClient/config"
)

func main(){
	fmt.Println("DazeClient V1.0 Author:螃蟹")
	fmt.Println("DazeProxyServer:",config.GetServerIP())
	go proxy.StartHttpsProxy(":8080")
	proxy.StartSocks5(":1080")
}
