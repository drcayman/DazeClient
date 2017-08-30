package common

type Json_Auth struct{
	Username string
	Password string
	Net string
	Host string
	Port string
	Spam string
}
type Json_Ret struct{
	Code int
	Data string
}
type Json_UDP struct {
	Host string
	Data []byte
}