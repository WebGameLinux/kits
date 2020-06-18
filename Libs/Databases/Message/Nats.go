package Message

import (
		"github.com/nats-io/nats.go"
		"os"
)

type Connector interface {
		Conn() *nats.Conn
}

type NatsConnector struct {
		connector *nats.Conn
		info      *nats.Options
}

const (
		EnvNatsUrl = "nats_url"
)

func NewConnector(args ...interface{}) Connector {
		var conn = new(NatsConnector)
		conn.init(args...)
		return conn
}

func (this *NatsConnector) init(args ...interface{}) {
		for _, arg := range args {
				if opt, ok := arg.(nats.Options); ok && this.info == nil {
						this.info = &opt
				}
				if opt, ok := arg.(*nats.Options); ok && this.info == nil {
						this.info = opt
				}
				if conn, ok := arg.(*nats.Conn); ok && this.connector == nil {
						this.connector = conn
				}
				if conn, ok := arg.(nats.Conn); ok && this.connector == nil {
						this.connector = &conn
				}
		}
		this.defaults()
}

func (this *NatsConnector) Conn() *nats.Conn {
		if this.connector == nil {
				this.open()
		}
		return this.connector
}

func (this *NatsConnector) open() {
		if this.info == nil {
				this.defaults()
		}
		this.connector, _ = nats.Connect(this.info.Url)
}

func (this *NatsConnector) defaults() {
		if this.info == nil {
				this.info = &nats.Options{}
		}
		if this.info.Url == "" {
				this.info.Url = os.Getenv(EnvNatsUrl)
				if this.info.Url == "" {
						this.info.Url = nats.DefaultURL
				}
		}
}

func (this *NatsConnector) Close() {
		if this.connector == nil {
				return
		}
		this.connector.Close()
		this.connector = nil
}
