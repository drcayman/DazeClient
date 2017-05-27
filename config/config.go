package config

var serverip string
var debug bool

func GetServerIP() string{
	return serverip
}
func GetDebug() bool{
	return debug
}
func init(){
	serverip="127.0.0.1:5294"  //test
}
