package mylog

import (
	"log"
	"fmt"
	"DazeClient/config"
)
func DPrintln(v ...interface{}){
	if config.GetDebug(){
		str:=""
		for _,s:=range v{
			str+=fmt.Sprint(s)+" "
		}
		log.Println(str)
	}
}
func Println(v ...interface{}){
	str:=""
	for _,s:=range v{
		str+=fmt.Sprint(s)+" "
	}
	log.Println(str)
}
