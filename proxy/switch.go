package proxy
import(
	"net"
	"github.com/crabkun/DazeClient/common"
	"log"
	"github.com/crabkun/DazeClient/helper"
	"sync/atomic"
	"time"
)
var ServerListener *net.Listener
func StartProxy() (error){
	listener, err := net.Listen("tcp", "127.0.0.1:"+common.SrvConf.LocalPort)
	if err != nil {
		log.Println("本地HTTP/SOCKS5代理监听失败！原因： ", err)
		return err
	}
	log.Println("本地HTTP/SOCKS5代理成功监听于",listener.Addr())
	helper.GenProxyAllPac()
	go func(){
		for {
			conn, err := listener.Accept()
			if err != nil {
				if err,ok:=err.(net.Error);ok&&!err.Temporary(){
					return
				}
				continue
			}
			go handleConnection(conn)
		}
	}()
	ServerListener=&listener
	return nil
}
func handleConnection(conn net.Conn){
	defer func(){
		if err := recover(); err != nil{
			conn.Close()
		}
	}()
	testchar:=make([]byte,1)
	_,err:=conn.Read(testchar)
	if err!=nil{
		return
	}
	if testchar[0]==5{
		Socks5handleConnection(&SwitchConn{conn,testchar})
	}else{
		HTTPProxyHandle(&SwitchConn{conn,testchar})
	}
}

type SwitchConn struct {
	net.Conn
	Testchar []byte
}

var Download uint64
var Upload uint64
var LastDownload uint64
var LastUpload uint64
func NetSpeedMonitor(show bool){
	for{
		time.Sleep(time.Second)
		LastDownload=atomic.LoadUint64(&Download)
		LastUpload=atomic.LoadUint64(&Upload)
		if show{
			print("当前网速[上传:",LastUpload/1024,"KB/s，下载：",LastDownload/1024,"KB/S]                                  \r")
		}
		atomic.StoreUint64(&Download,0)
		atomic.StoreUint64(&Upload,0)
	}
}
func (this *SwitchConn) Read(b []byte) (n int, err error){
	if this.Testchar!=nil{
		b[0]=this.Testchar[0]
		this.Testchar=nil
		b=b[1:]
		n,err:=this.Conn.Read(b)
		return n+1,err
	}
	n,err=this.Conn.Read(b)
	atomic.AddUint64(&Upload,uint64(n))
	 return
}
func (this *SwitchConn)Write(b []byte) (n int, err error){
	n,err=this.Conn.Write(b)
	atomic.AddUint64(&Download,uint64(n))
	return
}
func RestartServer()(error){
	if ServerListener!=nil{
		(*ServerListener).Close()
	}
	return StartProxy()
}