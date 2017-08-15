package disguise

import "net"

type none struct {
	reserved string
}

func (this *none) Init(arg string,client *interface{})(error){
	return nil
}
func (this *none) Action(conn net.Conn ,client *interface{}) (error){
	return nil
}