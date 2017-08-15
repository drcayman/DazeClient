package disguise

import (
	"net"
	"net/http"
	"bytes"
	mrand "math/rand"
	"time"
	"DazeClient/util"
	"bufio"
	"github.com/pkg/errors"
)
type HTTP struct {
	reserved string
}

func (this *HTTP) Init(arg string,client *interface{})(error){
	*client=arg
	return nil
}
func (this *HTTP) Action(conn net.Conn ,client *interface{}) (error){
	host,flag:=(*client).(string)
	if !flag{
		return errors.New("unknown error")
	}
	body:=make([]byte,0)
	if this.reserved=="POST"{
		bodystr:=this.GetRandomString(mrand.Intn(10))+"="+
			this.GetRandomString(mrand.Intn(512))+"&"+
			this.GetRandomString(mrand.Intn(10))+"="+
			this.GetRandomString(mrand.Intn(512))
		body=util.S2b(&bodystr)
	}
	req,reqErr:=http.NewRequest(this.reserved,"http://"+host+"/"+this.GetRandomString(mrand.Intn(10))+".php",bytes.NewReader(body))
	if reqErr!=nil{
		return reqErr
	}
	req.Write(conn)
	_,err:=http.ReadResponse(bufio.NewReader(conn),nil)
	conn.Write([]byte{byte(mrand.Intn(128))})
	return err
}
func (this *HTTP)GetRandomString(strlen int) string{
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	byts := []byte(str)
	result := []byte{}
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	for i := 0; i < strlen; i++ {
		result = append(result, byts[r.Intn(len(byts))])
	}
	return string(result)
}
