package obscure

import (
	"net"
	"net/http"
	"math/rand"
	"github.com/crabkun/DazeClient/util"
	"bytes"
	"bufio"
)

type HttpGet struct {
}

func (this *HttpGet) Action(conn net.Conn , param string) (error){
	var err error
	body:=make([]byte,0)
	req,err:=http.NewRequest("GET","http://"+param+"/"+util.GetRandomString(rand.Intn(10))+".php",bytes.NewReader(body))
	if err!=nil{
		return err
	}
	req.Header=make(http.Header)
	req.Header.Add("Connection","Keep-Alive")
	req.Header.Add("Accept","*/*")
	req.Write(conn)
	reader:=bufio.NewReader(conn)
	_,err=http.ReadResponse(reader,nil)
	if err!=nil{
		return err
	}
	req.Write(conn)
	return nil
}
func init(){
	RegisterObscure("http_get",new(HttpGet))
}