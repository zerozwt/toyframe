package tests

//go:generate msgp
type ChatMsg_Login struct {
	Name string `msg:"name"`
}

type ChatMsg_Send struct {
	Reciever string `msg:"recv"`
	Content  string `msg:"msg"`
}
