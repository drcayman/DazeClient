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
func GetDisguise() (string,string){
	return ConfigArr[NowSelect].Disguise,ConfigArr[NowSelect].DisguiseParam
}
func GetEncryption() (string,string){
	return ConfigArr[NowSelect].Encryption,ConfigArr[NowSelect].EncryptionParam
}
type ConfigStruct struct{
	ServerIP string
	ServerPort string
	Username string
	Password string
	Disguise string
	DisguiseParam string
	Encryption string
	EncryptionParam string
	Debug bool
}
type GlobaConfigStruct struct{
	HTTPProxyPort string
	Socks5Port string
	PAC string
}
var GlobaConfig GlobaConfigStruct
var ConfigArr []ConfigStruct
func loadConfFile(filepath string) bool {
	buf,ReadFileErr:=ioutil.ReadFile(filepath)
	if ReadFileErr!=nil{
		fmt.Println("无法访问配置文件：",filepath,"原因：",ReadFileErr.Error())
		return false
	}
	config:=ConfigStruct{}
	err:=json.Unmarshal(buf,&config)
	if err!=nil{
		fmt.Println("寻找到配置文件(",filepath,"),加载失败，原因：",err.Error())
	}else{
		fmt.Println("寻找到配置文件(",filepath,"),加载成功")
		ConfigArr=append(ConfigArr,config)
	}
	return true
}
func LoadConfFile(filepath string){
	if !loadConfFile(filepath){
		os.Exit(-1)
	}else{
		fmt.Printf("配置文件%s加载成功，地址：%s 加密：%s 加密参数：%s 伪装:%s 伪装参数：%s 调试：%v\n",filepath,ConfigArr[0].ServerIP+":"+
			ConfigArr[0].ServerPort,
			ConfigArr[0].Encryption,
			ConfigArr[0].EncryptionParam,
			ConfigArr[0].Disguise,
			ConfigArr[0].DisguiseParam,
			ConfigArr[0].Debug,
		)
		NowSelect=0
	}

}
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
		loadConfFile("conf/"+file.Name())
	}
	if len(ConfigArr)==0{
		fmt.Println("没有找到可用配置文件！请确认是否正确的把配置文件放到conf目录。")
		os.Exit(-1)
	}
	fmt.Println("*************\n可用配置文件列表：")
	for i,file:=range ConfigArr {
		fmt.Printf("ID：%d  地址：%s 加密：%s 加密参数：%s 伪装:%s 伪装参数：%s 调试：%v\n",i+1,file.ServerIP+":"+file.ServerPort,
					file.Encryption,file.EncryptionParam,file.Disguise,file.DisguiseParam,file.Debug)
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
