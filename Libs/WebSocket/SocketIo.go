package WebSocket

import (
		"github.com/google/martian/log"
		engineio "github.com/googollee/go-engine.io"
		socketio "github.com/googollee/go-socket.io"
		"sync"
)

func SocketIoServerOf(args ...interface{}) *socketio.Server {
		return initServer(args...)
}

func initServer(args ...interface{}) *socketio.Server {
		var (
				argc = len(args)
		)
		if argc == 0 {
				socket, err := socketio.NewServer(nil)
				if err != nil {
						log.Errorf(err.Error())
				}
				return socket
		}
		for _, arg := range args {
				if opt, ok := arg.(engineio.Options); ok {
						socket, err := socketio.NewServer(&opt)
						if err != nil {
								log.Errorf(err.Error())
						}
						return socket
				}
				if opt, ok := arg.(*engineio.Options); ok {
						socket, err := socketio.NewServer(opt)
						if err != nil {
								log.Errorf(err.Error())
						}
						return socket
				}
		}
		socket, err := socketio.NewServer(nil)
		if err != nil {
				log.Errorf(err.Error())
		}
		return socket
}

type SocketConnPool struct {
		mutex sync.Mutex
		size  int
		conns map[string]socketio.Conn
}

func NewConnPool() *SocketConnPool {
		var pool = new(SocketConnPool)
		pool.size = 0
		pool.conns = make(map[string]socketio.Conn)
		return pool
}

func (this *SocketConnPool) Add(id string, conn socketio.Conn) {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		this.conns[id] = conn
		this.size = len(this.conns)
}

func (this *SocketConnPool) Get(id string) socketio.Conn {
		return this.conns[id]
}

func (this *SocketConnPool) Remove(id string) {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		delete(this.conns, id)
}

func (this *SocketConnPool) Count() int {
		return this.size
}

func (this *SocketConnPool) Len() int {
		return this.size
}

func (this *SocketConnPool) Foreach(each func(id string, conn socketio.Conn) bool) {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		for id, conn := range this.conns {
				if !each(id, conn) {
						break
				}
		}
}

func (this *SocketConnPool) Filter(filter func(id string, conn socketio.Conn) bool) *SocketConnPool {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		var (
				pool = NewConnPool()
		)
		for id, conn := range this.conns {
				if filter(id, conn) {
						pool.Add(id, conn)
				}
		}
		return pool
}
