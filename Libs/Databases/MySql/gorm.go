package MySql

import (
		"encoding/json"
		"fmt"
		"github.com/jinzhu/gorm"
		_ "github.com/jinzhu/gorm/dialects/mysql"
		"os"
		"strings"
		"unicode"
)

type OrmMysqlConnector struct {
		connector *gorm.DB
		info      ConnectionInterface
}

type GormConnector interface {
		Close()
		Conn(...string) *gorm.DB
}

const (
		DriverMysql      = "mysql"
		EnvDbPort        = "db_port"
		EnvDbHost        = "db_host"
		DefaultDbPort    = "3306"
		DefaultHost      = "127.0.0.1"
		EnvDbCharset     = "db_charset"
		DefaultCharset   = "utf8"
		EnvDbLocal       = "db_local"
		DefaultLocal     = "Local"
		EnvDbParseTime   = "db_parse_time"
		DefaultParseTime = "True"
		EnvDbUser        = "db_user"
		DefaultDbUser    = "root"
		EnvDbName        = "db_name"
		DefaultDbName    = "test"
		EnvPassword      = "db_password"
)

type ConnectionInterface interface {
		GetUrl() string
		GetHost() string
		GetPort() string
		GetUserName() string
		GetPassword() string
		GetDatabase() string
		Args() map[string]string
}

type Connection struct {
		User      string `json:"user"`
		Password  string `json:"password"`
		Database  string `json:"dbname"`
		Charset   string `json:"charset"`
		ParseTime string `json:"parseTime"`
		Local     string `json:"loc"`
		Host      string `json:"host"`
		Port      string `json:"port"`
}

func NewConnection(args ...interface{}) *Connection {
		var connection = new(Connection)
		connection.init(args...)
		return connection
}

func (this *Connection) init(args ...interface{}) {
		var (
				argc = len(args)
		)
		if argc == 1 {
				if data, ok := args[0].(string); ok {
						this.initByJson(data)
						return
				}
				if data, ok := args[0].(map[string]string); ok {
						this.initByMapper(data)
						return
				}
				if data, ok := args[0].(ConnectionInterface); ok {
						this.initByConn(data)
						return
				}
		}
		this.initDefault()
}

func (this *Connection) initDefault() {
		if this.Port == "" || !this.isNumber(this.Port) {
				this.Port = os.Getenv(EnvDbPort)
				if this.Port == "" {
						this.Port = DefaultDbPort
				}
		}
		if this.Host == "" {
				this.Host = os.Getenv(EnvDbHost)
				if this.Host == "" {
						this.Host = DefaultHost
				}
		}
		if this.Charset == "" {
				this.Charset = os.Getenv(EnvDbCharset)
				if this.Charset == "" {
						this.Charset = DefaultCharset
				}
		}
		if this.Local == "" {
				this.Local = os.Getenv(EnvDbLocal)
				if this.Local == "" {
						this.Local = DefaultLocal
				}
		}
		if this.ParseTime == "" {
				this.ParseTime = os.Getenv(EnvDbParseTime)
				if this.ParseTime == "" {
						this.ParseTime = DefaultParseTime
				}
		}
		if this.User == "" {
				this.User = os.Getenv(EnvDbUser)
				if this.User == "" {
						this.User = DefaultDbUser
				}
		}
		if this.Database == "" {
				this.Database = os.Getenv(EnvDbName)
				if this.Database == "" {
						this.Database = DefaultDbName
				}
		}
		if this.Password == "" {
				this.Password = os.Getenv(EnvPassword)
		}
		if !unicode.IsUpper(rune(this.ParseTime[0])) {
				this.ParseTime = strings.ToUpper(this.ParseTime[0:1]) + this.ParseTime[1:]
		}
}

func (this *Connection) isNumber(v string) bool {
		for _, char := range []rune(v) {
				if !unicode.IsNumber(char) {
						return false
				}
		}
		return true
}

func (this *Connection) initByJson(data string) {
		_ = json.Unmarshal([]byte(data), this)
}

func (this *Connection) initByMapper(data map[string]string) {
		for k, v := range data {
				this.set(k, v)
		}
}

func (this *Connection) initByConn(data ConnectionInterface) {
		this.User = data.GetUserName()
		this.Password = data.GetPassword()
		this.Host = data.GetHost()
		for key, v := range data.Args() {
				this.set(key, v)
		}
}

func (this *Connection) set(key string, v string) {
		if v == "" || key == "" {
				return
		}
		switch key {
		case "user":
				fallthrough
		case "username":
				this.User = v
		case "pass":
				fallthrough
		case "password":
				this.Password = v
		case "ip":
				fallthrough
		case "host":
				this.Host = v
		case "database":
				fallthrough
		case "dbname":
				this.Database = v
		case "charset":
				this.Charset = v
		case "parseTime":
				fallthrough
		case "parse_time":
				if v == "true" || v == "True" || v == "false" || v == "False" {
						this.ParseTime = v
				}
		case "loc":
				fallthrough
		case "local":
				this.Local = v
		case "port":
				this.Port = v
		}
}

func (this *Connection) GetUserName() string {
		return this.User
}

func (this *Connection) GetPassword() string {
		return this.Password
}

func (this *Connection) GetDatabase() string {
		return this.Database
}

func (this *Connection) GetHost() string {
		return this.Host
}

func (this *Connection) Args() map[string]string {
		var args = make(map[string]string)
		args["charset"] = this.Charset
		args["loc"] = this.Local
		args["parseTime"] = this.ParseTime
		return args
}

func (this *Connection) GetUrl() string {
		return this.toString()
}

func (this *Connection) GetPort() string {
		return this.Port
}

func (this *Connection) String() string {
		return this.toString()
}

func (this *Connection) toString() string {
		if this.Charset == "" || this.ParseTime == "" {
				this.initDefault()
		}
		if this.Local == "" || this.Host == "" {
				this.initDefault()
		}
		return fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s",
				this.User, this.Password, this.Host, this.Port, this.Database,
				this.Charset, this.ParseTime, this.Local,
		)
}

func NewMysqlConnector(args ...interface{}) *OrmMysqlConnector {
		var connector = new(OrmMysqlConnector)
		if len(args) != 0 {
				connector.init(args...)
		}
		return connector
}

func (this *OrmMysqlConnector) Init() {
		if this.connector != nil {
				return
		}
		if this.info == nil {
				this.info = NewConnection()
		}
		this.open()
}

func (this *OrmMysqlConnector) open() {
		url := this.info.GetUrl()
		db, err := gorm.Open(DriverMysql, url)
		if err == nil {
				this.connector = db
		}
}

func (this *OrmMysqlConnector) Close() {
		if this.connector == nil {
				return
		}
		_ = this.connector.Close()
		this.connector = nil
}

func (this *OrmMysqlConnector) init(args ...interface{}) {
		var (
				argc = len(args)
		)
		if argc == 0 {
				return
		}
		for _, arg := range args {
				if db, ok := arg.(*gorm.DB); ok && this.connector == nil {
						this.connector = db
				}
				if info, ok := arg.(ConnectionInterface); ok && this.info == nil {
						this.info = info
				}
				if info, ok := arg.(Connection); ok && this.info == nil {
						this.info = &info
				}
				if data, ok := arg.(string); ok && this.info == nil {
						this.info = NewConnection(data)
				}
		}
}

func (this *OrmMysqlConnector) Conn(db ...string) *gorm.DB {
		if this.info == nil {
				this.info = NewConnection()
		}
		if len(db) != 0 && this.info.GetDatabase() != db[0] {
				connection := NewConnection(this.info)
				connection.Database = db[0]
				conn, _ := gorm.Open(connection.GetUrl())
				return conn
		}
		if this.connector == nil {
				this.open()
		}
		return this.connector
}
