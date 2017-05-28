package config

var serverip string
var debug bool=true

func GetServerIP() string{
	return serverip
}
func GetDebug() bool{
	return debug
}
func init(){
	serverip="45.32.34.191:5294"  //test
}
