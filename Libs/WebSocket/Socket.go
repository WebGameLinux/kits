package WebSocket

import (
		"github.com/gorilla/websocket"
		"github.com/prometheus/common/log"
		"net/http"
)

func NewUpgrader() websocket.Upgrader {
		var socket = websocket.Upgrader{}
		return socket
}

var (
		SocketInstance *websocket.Conn
)

func SocketOf(args ...interface{}) *websocket.Conn {
		if SocketInstance != nil && len(args) == 0 {
				return SocketInstance
		}
		return newWebSocket(args...)
}

func newWebSocket(args ...interface{}) *websocket.Conn {
		var (
				r      *http.Request
				w      http.ResponseWriter
				header http.Header
				argc   = len(args)
		)
		if argc < 2 {
				return nil
		}
		for _, arg := range args {
				if req, ok := arg.(*http.Request); ok && r == nil {
						r = req
				}
				if writer, ok := arg.(http.ResponseWriter); ok && w == nil {
						w = writer
				}
				if h, ok := arg.(http.Header); ok && header == nil {
						header = h
				}
				if h, ok := arg.(*http.Header); ok && header == nil {
						header = *h
				}
		}
		if r == nil || w == nil {
				return nil
		}
		upgrader := NewUpgrader()
		conn, err := upgrader.Upgrade(w, r, header)
		if err == nil {
				return conn
		}
		log.Error(err)
		return conn
}

func SocketHandler() http.Handler {
		return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request){
				SocketOf(w,r,r.Header)
		})
}

