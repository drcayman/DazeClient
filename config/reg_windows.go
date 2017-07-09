// +build windows

package config

import (
	"golang.org/x/sys/windows/registry"
	"fmt"
	"syscall"
	"net/http"
	"io/ioutil"
	"bytes"
)
func SetSystemProxy(){
	k,_,err:=registry.CreateKey(registry.CURRENT_USER,"Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",registry.ALL_ACCESS)
	if err!=nil{
		fmt.Println("设置系统代理失败！")
		return
	}
	k.SetDWordValue("ProxyEnable",1)
	k.SetStringValue("ProxyServer","127.0.0.1:"+GlobaConfig.HTTPProxyPort)
	k.DeleteValue("AutoConfigURL")
	NotifySystem()
	fmt.Println("设置系统HTTP代理成功！请勿忘记恢复系统代理，否则会造成无法上网，解决方法是手动关闭IE代理或者输入off。")
}
func ClearSystemProxy(){
	k,_,err:=registry.CreateKey(registry.CURRENT_USER,"Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",registry.ALL_ACCESS)
	if err!=nil{
		fmt.Println("恢复系统代理失败！请手动关闭IE代理！")
		return
	}
	k.SetDWordValue("ProxyEnable",0)
	k.DeleteValue("AutoConfigURL")
	NotifySystem()
	fmt.Println("恢复系统代理成功！")
}
func UpdatePac() bool{
	rsp,err:=http.Get(GlobaConfig.PAC)
	if err!=nil{
		return false
	}
	buf,_:=ioutil.ReadAll(rsp.Body)
	buf=bytes.Replace(buf,[]byte("SOCKS5 127.0.0.1:1080"),[]byte("PROXY 127.0.0.1:"+GlobaConfig.HTTPProxyPort),1)
	ioutil.WriteFile("gfwlist.pac",buf,0666)
	if err!=nil{
		return false
	}
	fmt.Println("更新gfwlist成功！")
	return true
}
func SetPacProxyGFW(){
	//if _,err:=os.Stat("gfwlist.pac");err!=nil{
		if !UpdatePac(){
			fmt.Println("更新gfwlist失败！请检查globa.conf里面的PAC项是否可用！")
			return
		}
	//}
	ClearSystemProxy()
	k,_,err:=registry.CreateKey(registry.CURRENT_USER,"Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",registry.ALL_ACCESS)
	if err!=nil{
		fmt.Println("设置系统代理失败！")
		return
	}
	k.SetStringValue("AutoConfigURL","http://127.0.0.1:"+GlobaConfig.HTTPProxyPort+"/!daze.pac")
	NotifySystem()
	fmt.Println("设置系统HTTP代理PAC模式成功！")
}
func SetPacProxyDirect(){
	ClearSystemProxy()
	k,_,err:=registry.CreateKey(registry.CURRENT_USER,"Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",registry.ALL_ACCESS)
	if err!=nil{
		fmt.Println("设置系统代理失败！")
		return
	}
	k.SetStringValue("AutoConfigURL","http://127.0.0.1:"+GlobaConfig.HTTPProxyPort+"/!dazeD.pac")
	NotifySystem()
	fmt.Println("设置系统HTTP代理成功！")
}
func init(){
	NotifySystem()
}
func NotifySystem(){
	syscall.MustLoadDLL("Wininet.dll").MustFindProc("InternetSetOptionA").Call(0,39,0,0)
	syscall.MustLoadDLL("Wininet.dll").MustFindProc("InternetSetOptionA").Call(0,37,0,0)
	syscall.MustLoadDLL("Wininet.dll").MustFindProc("InternetSetOptionA").Call(0,95,0,0)
}