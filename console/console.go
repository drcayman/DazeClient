package console

import (
	"fmt"
	"bufio"
	"os"
	"DazeClient/util"
	"strings"
	"time"
	"DazeClient/config"
	"io/ioutil"
	"strconv"
)

func ShowMenu(){
	fmt.Println("**********命令列表**********")
	fmt.Println("help 显示此帮助")
	fmt.Println("sel 选择配置文件（比如sel 1）")
	fmt.Println("debug 临时切换Debug开关")
	fmt.Println("on 设置系统代理并代理仅被GFW河蟹的IP")
	fmt.Println("ond 设置系统代理并代理所有IP")
	fmt.Println("off 恢复系统代理")
	fmt.Println("upd 更新gfwlist规则文件")
	fmt.Println("****************************")
}
func sel(num int){
	if num<1||len(config.ConfigArr)<num{
		fmt.Println("不存在此配置文件")
		return
	}
	config.NowSelect=num-1
	fmt.Printf("成功切换到ID为%d的配置\n",num)
	ioutil.WriteFile("conf/lastPos",[]byte(strconv.Itoa(num)),0666)
}
func Start(){
	time.Sleep(time.Second*1)
	ShowMenu()
	r:=bufio.NewReader(os.Stdin)
	command:=""
	for{
		fmt.Print(">>>>>>")
		buf,_,_:=r.ReadLine()
		bufstr:=util.B2s(buf)
		n,_:=fmt.Sscanf(bufstr,"%s",&command)
		if n==0{
			continue
		}
		switch strings.ToLower(command) {
		case "help":
			ShowMenu()
		case "sel":
			var num int
			n,_:=fmt.Sscanf(bufstr,"%s%d",&command,&num)
			if n!=2{
				fmt.Println("命令格式错误")
				continue
			}
			sel(num)
		case "debug":
			config.ConfigArr[config.NowSelect].Debug=!config.ConfigArr[config.NowSelect].Debug
			fmt.Println("DEBUG：",config.ConfigArr[config.NowSelect].Debug)
		case "on":
			config.SetPacProxyGFW()
		case "upd":
			config.UpdatePac()
		case "ond":
			config.SetPacProxyDirect()
		case "off":
			config.ClearSystemProxy()
		default:
			fmt.Println("命令格式错误，请输入help来查看帮助")
		}

	}
}