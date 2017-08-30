package encryption

import (
	"strings"
	"net"
)
type Action interface {
	//InitUser，连接代理服务器后进行的初始化操作
	//conn：服务器的连接套接字
	//param：配置文件里面填写的EncryptionParam
	//client：此用户对象中为加密模块预留的空间
	InitUser(conn net.Conn,param string,client *interface{})(error)

	//Encrypt，加密
	//client同上
	//data：源数据
	//输出 加密后的数据 与 一个error(若发生了错误)
	Encrypt(client *interface{},	data []byte) ([]byte,error)

	//Decrypt，解密
	//client同上
	//data：加密数据
	//输出 解密后的数据 与 一个error(若发生了错误)
	Decrypt(client *interface{},	data []byte) ([]byte,error)
}
type regfunc func()(Action)
var encryptionMap map[string]regfunc

func GetEncryption(name string) (regfunc,bool){
	name=strings.ToLower(name)
	d,flag:=encryptionMap[name]
	return d,flag
}

func init(){
	encryptionMap=make(map[string]regfunc)

	//自己开发的加密模块必需在此注册
	encryptionMap["none"]=func()(Action){
		return Action(&none{})
	}
	encryptionMap["keypair-aes"]=func()(Action){
		return Action(&KeypairAes{})
	}
	encryptionMap["psk-aes-128-cfb"]=func()(Action){
		return Action(&PskAesCfb{})
	}
	encryptionMap["psk-aes-256-cfb"]=func()(Action){
		return Action(&PskAes256Cfb{})
	}
	encryptionMap["psk-rc4-md5"]=func()(Action){
		return Action(&PskRc4Md5{})
	}
}

