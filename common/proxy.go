package common

type S_proxy struct{
	Address string
	Port string
	Username string
	Password string

	//加密方式与参数
	Encryption string
	EncryptionParam string

	//混淆方式与参数
	Obscure string
	ObscureParam string

	//本地监听端口
	LocalPort string
	Debug bool
}

var SrvConf *S_proxy

func init(){
	SrvConf=&S_proxy{}
}