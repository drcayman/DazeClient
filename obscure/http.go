package obscure

import (
	"net"
	"net/http"
	"math/rand"
	"github.com/crabkun/DazeClient/util"
	"bytes"
	"bufio"
	"errors"
)

type Http struct {
	RegArg string
}

func (this *Http) Action(conn net.Conn , param string) (error){
	var err error
	body:=make([]byte,0)
	if this.RegArg=="POST"{
		bodystr:=util.GetRandomString(rand.Intn(10))+"="+
			util.GetRandomString(rand.Intn(512))+"&"+
			util.GetRandomString(rand.Intn(10))+"="+
			util.GetRandomString(rand.Intn(512))
		body=[]byte(bodystr)
	}
	req,err:=http.NewRequest(this.RegArg,"http://"+param+"/"+util.GetRandomString(rand.Intn(10))+".php",bytes.NewReader(body))
	if err!=nil{
		return err
	}
	req.Header=make(http.Header)
	req.Header.Add("Connection","Keep-Alive")
	req.Header.Add("Accept","*/*")
	req.Write(conn)
	reader:=bufio.NewReader(conn)
	rsp,err:=http.ReadResponse(reader,nil)
	if err!=nil{
		return err
	}
	conn.Write([]byte(util.GetRandomString(int(rsp.ContentLength))))
	return nil
}
func SafeRead(conn net.Conn,length int) ([]byte,error) {
	buf:=make([]byte,length)
	for pos:=0;pos<length;{
		n,err:=conn.Read(buf[pos:])
		if err!=nil {
			return nil,errors.New("根据Content-Length读取负载错误")
		}
		pos+=n
	}
	return buf,nil
}
