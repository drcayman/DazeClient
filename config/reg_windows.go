// +build windows

package config

import (
	"golang.org/x/sys/windows/registry"
	"fmt"
)
func SetSystemProxy(){
	k,_,err:=registry.CreateKey(registry.CURRENT_USER,"Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",registry.ALL_ACCESS)
	if err!=nil{
		fmt.Println("设置系统代理失败！")
		return
	}
	k.SetDWordValue("ProxyEnable",1)
	k.SetStringValue("ProxyServer","127.0.0.1:"+GlobaConfig.HTTPProxyPort)
	fmt.Println("设置系统HTTP代理成功！请勿非正常关闭此工具，否则会造成无法上网，解决方法是手动关闭IE代理或者正常关闭此工具。")
}
func ClearSystemProxy(){
	k,_,err:=registry.CreateKey(registry.CURRENT_USER,"Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",registry.ALL_ACCESS)
	if err!=nil{
		fmt.Println("恢复系统代理失败！请手动关闭IE代理！")
		return
	}
	k.SetDWordValue("ProxyEnable",0)
	fmt.Println("恢复系统代理成功！")
}