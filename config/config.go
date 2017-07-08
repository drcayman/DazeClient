package config

import (
	"io/ioutil"
	"strings"
	"encoding/json"
	"os"
	"fmt"
	"strconv"
	"DazeClient/util"
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
