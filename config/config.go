package config

var serverip string

func GetServerIP() string{
	return serverip
}
func init(){
	serverip="127.0.0.1:5294"  //test
}
