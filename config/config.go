package config

import (
	"io/ioutil"
	"strings"
	"encoding/json"
	"os"
	"fmt"
	"strconv"
	"DazeClient/util"
	"golang.org/x/sys/windows/registry"
	"runtime"
)
var NowSelect int
func GetServerIP() string{
	return ConfigArr[NowSelect].ServerIP+":"+GetServerPort()
}
func GetServerPort() string{
	return ConfigArr[NowSelect].ServerPort
}
func GetDebug() bool{
	return ConfigArr[NowSelect].Debug
}
func GetUsername() string{
	return ConfigArr[NowSelect].Username
}
func GetPassword() string{
	return ConfigArr[NowSelect].Password
}
type ConfigStruct struct{
	ServerIP string
	ServerPort string
	Username string
	Password string
	Debug bool
}
type GlobaConfigStruct struct{
	HTTPProxyPort string
	Socks5Port string
	SystemProxy bool
}
var GlobaConfig GlobaConfigStruct
var ConfigArr []ConfigStruct
func Load(){
	globabuf,err:=ioutil.ReadFile("globa.conf")
	if err!=nil{
		fmt.Println("全局配置文件(globa.conf)加载错误！")
		os.Exit(-1)
	}
	err=json.Unmarshal(globabuf,&GlobaConfig)
	if err!=nil{
		fmt.Println("全局配置文件(globa.conf)解析错误！原因：",err.Error())
		os.Exit(-1)
	}
	if GlobaConfig.HTTPProxyPort=="" && GlobaConfig.Socks5Port==""{
		fmt.Println("Http代理和Socks5代理的端口不能同时为空！")
		os.Exit(-1)
	}
	ConfigArr=make([]ConfigStruct,0)
	files,_:=ioutil.ReadDir("conf")
	fmt.Println("*************\n开始加载配置文件：")
	for _,file:=range files{
		if strings.Index(file.Name(),".conf")==-1{
			continue
		}
		buf,_:=ioutil.ReadFile("conf/"+file.Name())
		config:=ConfigStruct{}
		err:=json.Unmarshal(buf,&config)
		if err!=nil{
			fmt.Println("寻找到配置文件(",file.Name(),"),加载失败，原因：",err.Error())
		}else{
			fmt.Println("寻找到配置文件(",file.Name(),"),加载成功")
			ConfigArr=append(ConfigArr,config)
		}
	}
	if len(ConfigArr)==0{
		fmt.Println("没有找到可用配置文件！请确认是否正确的把配置文件放到conf目录。")
		os.Exit(-1)
	}
	fmt.Println("*************\n可用配置文件列表：")
	for i,file:=range ConfigArr {
		fmt.Printf("ID：%d  地址：%s\n",i+1,file.ServerIP+":"+file.ServerPort)
	}
	//加载最后一次的配置文件
	lastbuf,err:=ioutil.ReadFile("conf/lastPos")
	if err==nil{
		i,err:=strconv.Atoi(util.B2s(lastbuf))
		if err==nil{
			fmt.Println("加载上次使用的配置文件：",util.B2s(lastbuf))
			sel(i)
		}
	}
	fmt.Println("使用配置文件ID：",NowSelect+1)
	fmt.Println("*************")
}
func sel(num int){
	if num<1||len(ConfigArr)<num{
		fmt.Println("不存在此配置文件")
		return
	}
	NowSelect=num-1
	fmt.Printf("成功切换到ID为%d的配置\n",num)
	ioutil.WriteFile("conf/lastPos",[]byte(strconv.Itoa(num)),0666)
}
func SetSystemProxy(){
	if runtime.GOOS!="windows"{
		return
	}
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
	if runtime.GOOS!="windows"{
		return
	}
	k,_,err:=registry.CreateKey(registry.CURRENT_USER,"Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings",registry.ALL_ACCESS)
	if err!=nil{
		fmt.Println("恢复系统代理失败！请手动关闭IE代理！")
		return
	}
	k.SetDWordValue("ProxyEnable",0)
	fmt.Println("恢复系统代理成功！")
}