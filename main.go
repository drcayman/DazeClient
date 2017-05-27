package main

import (
	"DazeClient/proxy"
)

func main(){
	go proxy.StartHttpsProxy(":8080")
	proxy.StartSocks5(":1080")
}
