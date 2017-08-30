package obscure

import (
	"strings"
	"net"
)
type Action interface {
	//Action，用户连接后进行的伪装操作
	//conn：用户的连接套接字
	//param：配置文件里面填写的ObscureParam
	Action(conn net.Conn,	param string)(error)
}
type regfunc func()(Action)
var obscureMap map[string]regfunc

func GetObscure(name string) (regfunc,bool){
	name=strings.ToLower(name)
	d,flag:=obscureMap[name]
	return d,flag
}

func init(){
	obscureMap=make(map[string]regfunc)
	obscureMap["none"]=func()(Action){
		return Action(&none{})
	}
	obscureMap["tls_handshake"]=func()(Action){
		return Action(&TlsHandshake{})
	}
	obscureMap["http_get"]=func()(Action){
		return Action(&Http{"GET"})
	}
	obscureMap["http"]=func()(Action){
		return Action(&Http{"GET"})
	}
	obscureMap["http_post"]=func()(Action){
		return Action(&Http{"POST"})
	}
}
