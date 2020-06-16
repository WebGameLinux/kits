package Mongodb

import "gopkg.in/mgo.v2"

type Connector interface {
		Conn() *mgo.Session
		Close()
}

type Mongodb struct {
		connector *mgo.Session
		url       string
}

const (
		DefaultMongodbUrl = "127.0.0.1:27017"
)

func NewMog(args ...interface{}) Connector {
		var m = new(Mongodb)
		m.init(args...)
		m.defaults()
		return m
}

func (this *Mongodb) defaults() {
		if this.url == "" {
				this.url = DefaultMongodbUrl
		}
}

func (this *Mongodb) init(args ...interface{}) {
		if len(args) == 0 {
				return
		}
		if url, ok := args[0].(string); ok {
				this.url = url
		}
		return
}

func (this *Mongodb) Conn() *mgo.Session {
		if this.connector == nil {
				this.open()
		}
		return this.connector
}

func (this *Mongodb) open() {
		if this.connector != nil {
				return
		}
		this.connector, _ = mgo.Dial(this.url)
}

func (this *Mongodb) Close() {
		if this.connector == nil {
				return
		}
		this.connector.Close()
		this.connector = nil
}
