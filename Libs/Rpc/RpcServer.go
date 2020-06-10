package Rpc

type Rpc interface {
		Server() Rpc
		Start() error
		Send(interface{})
}

type ServerRpcImpl struct {
}
