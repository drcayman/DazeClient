package mylog

import (
	"log"
	"DazeClient/config"
)
func DPrintln(v ...interface{}){
	if config.GetDebug(){
		log.Println(v)
	}
}
func Println(v ...interface{}){
		log.Println(v)
}
