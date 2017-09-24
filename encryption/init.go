package encryption

import (
	"net"
	"reflect"
	"errors"
)
type EncryptionAction interface {
	//InitUser，连接代理服务器后进行的初始化操作
	//conn：服务器的连接套接字
	//param：配置文件里面填写的EncryptionParam
	InitUser(conn net.Conn,param string)(error)

	//Encrypt，加密
	//data：源数据
	//输出 加密后的数据 与 一个error(若发生了错误)
	Encrypt(data []byte) ([]byte,error)

	//Decrypt，解密
	//data：加密数据
	//输出 解密后的数据 与 一个error(若发生了错误)
	Decrypt(data []byte) ([]byte,error)
}
var encryptionMap map[string]reflect.Type

func GetEncryption(name string) (EncryptionAction,bool){
	if encryptionMap==nil{
		goto FAILED
	}
	if v,ok:=encryptionMap[name];ok{
		return reflect.New(v).Interface().(EncryptionAction),true
	}
FAILED:
	return nil,false
}
func GetEncryptionList()[]string{
	list:=make([]string,0)
	for k,_:=range encryptionMap{
		list=append(list, k)
	}
	return list
}
func RegisterEncryption(name string,action EncryptionAction)(error){
	if encryptionMap==nil{
		encryptionMap=make(map[string]reflect.Type)
	}
	if _,ok:=encryptionMap[name];ok{
		return errors.New("exist")
	}
	Ptype:=reflect.ValueOf(action)
	STtype:=reflect.Indirect(Ptype).Type()
	encryptionMap[name]=STtype
	return nil
}
