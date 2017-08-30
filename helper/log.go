package helper

import (
	"log"
	"github.com/crabkun/DazeClient/common"
)
func DebugPrintln(msg string){
	if common.SrvConf.Debug{
		log.Println(msg)
	}
}
