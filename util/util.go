package util

import (
	"encoding/pem"
	"crypto/x509"
	"os"
	"crypto/rsa"
	"crypto/rand"
	"crypto/cipher"
	"fmt"
	"crypto/aes"
	"strconv"
	"time"
	"strings"
)
func GenAESKey(bytes int) []byte{
	buf:=make([]byte,bytes)
	rand.Read(buf)
	return buf
}
func EncryptRSA(data []byte,KeyFileBuf []byte) ([]byte){
	block,_:=pem.Decode(KeyFileBuf)
	PublicKey,PublicKeyParseErr:=x509.ParsePKIXPublicKey(block.Bytes)
	if PublicKeyParseErr!=nil{
		fmt.Println("[×E]公钥文件解析错误！！系统强制退出",PublicKeyParseErr.Error())
		os.Exit(-1)
	}
	EncryptBuf,EncryptErr:=rsa.EncryptPKCS1v15(rand.Reader,PublicKey.(*rsa.PublicKey),data)
	if EncryptErr!=nil{
		fmt.Println("[×E1]公钥文件解析错误！！系统强制退出",EncryptErr.Error())
		os.Exit(-1)
	}
	return EncryptBuf
}
func EncryptRSAWithDer(data []byte,der []byte) ([]byte){
	PublicKey,PublicKeyParseErr:=x509.ParsePKIXPublicKey(der)
	if PublicKeyParseErr!=nil{
		fmt.Println("[×E]公钥文件解析错误！！系统强制退出",PublicKeyParseErr.Error())
		os.Exit(-1)
	}
	EncryptBuf,EncryptErr:=rsa.EncryptPKCS1v15(rand.Reader,PublicKey.(*rsa.PublicKey),data)
	if EncryptErr!=nil{
		fmt.Println("[×E1]公钥文件解析错误！！系统强制退出",EncryptErr.Error())
		os.Exit(-1)
	}
	return EncryptBuf
}
func DecryptAES(data []byte,key []byte) ([]byte,error){
	block,CipherErr:=aes.NewCipher(key)
	Decrypter:=cipher.NewCFBDecrypter(block,key[:block.BlockSize()])
	decoded:=make([]byte,len(data))
	Decrypter.XORKeyStream(decoded,data)
	return decoded,CipherErr
}
func EncryptAES(data []byte,key []byte) ([]byte,error){
	block,CipherErr:=aes.NewCipher(key)
	Decrypter:=cipher.NewCFBEncrypter(block,key[:block.BlockSize()])
	encoded:=make([]byte,len(data))
	Decrypter.XORKeyStream(encoded,data)
	return encoded,CipherErr
}
func GetAESKeyByDay() []byte{
	daystr:=strconv.FormatInt(int64(time.Now().UTC().Day()),10)
	if len(daystr)==1{
		return []byte(strings.Repeat(daystr,16))
	}else{
		return []byte(strings.Repeat(daystr,8))
	}
}