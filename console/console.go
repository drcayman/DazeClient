package console

import (
	"fmt"
	"bufio"
	"os"
	"DazeClient/util"
	"strings"
	"time"
	"DazeClient/config"
)

func ShowMenu(){
	fmt.Println("**********命令列表**********")
	fmt.Println("help 显示此帮助")
	fmt.Println("sel 选择配置文件（比如sel 1）")
	fmt.Println("debug 临时切换Debug开关")
	fmt.Println("****************************")
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
		fmt.Sscanf(bufstr,"%s",&command)
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
			if num<1||len(config.ConfigArr)<num{
				fmt.Println("不存在此配置文件")
				continue
			}
			config.NowSelect=num-1
			fmt.Printf("成功切换到ID为%d的配置\n",num)
		case "debug":
			config.ConfigArr[config.NowSelect].Debug=!config.ConfigArr[config.NowSelect].Debug
			fmt.Println("DEBUG：",config.ConfigArr[config.NowSelect].Debug)
		default:
			fmt.Println("命令格式错误，请输入help来查看帮助")
		}

	}
}