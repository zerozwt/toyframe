package tests

//go:generate msgp
type TestMsg struct {
	ID   int    `msg:"id"`
	Name string `msg:"name"`
	Data []byte `msg:"data"`
}
