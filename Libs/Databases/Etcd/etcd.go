package Etcd

import (
		"context"
		"crypto/tls"
		"encoding/json"
		"github.com/coreos/etcd/clientv3"
		"go.uber.org/zap"
		"google.golang.org/grpc"
		"os"
		"strconv"
		"strings"
		"time"
)

type Connector interface {
		Close()
		Conn() *clientv3.Client
		Context() context.Context
		GetConfig() clientv3.Config
		SetConfig(clientv3.Config) Connector
}

type ConnectorImpl struct {
		Config *clientv3.Config
		client *clientv3.Client
}

const (
		EnvEtcdUserName   = "etcd_username"
		EnvEtcdPassword   = "etcd_password"
		EnvEtcdEndpoints  = "etcd_endpoints"
		EnvEtcdCtxTimeout = "etcd_ctx_timeout"
		DefaultEndPoints  = "localhost:2379"
		ContextCancelFunc = "CancelFunc"
		DefaultTimeOut    = 5 * time.Second
)

func NewConnector(args ...interface{}) Connector {
		var conn = new(ConnectorImpl)
		conn.init(args...)
		return conn
}

func (this *ConnectorImpl) open() {
		if this.client == nil {
				this.client, _ = clientv3.New(this.GetConfig())
		}
}

func (this *ConnectorImpl) init(args ...interface{}) {
		var (
				argc = len(args)
		)
		if argc == 0 {
				this.defaults()
				return
		}
		for _, arg := range args {
				if data, ok := arg.(map[string]interface{}); ok {
						this.initByMapper(data)
				}
				if data, ok := arg.(string); ok && this.Config == nil {
						this.initCnfByJson([]byte(data))
				}
				if data, ok := arg.(clientv3.Config); ok && this.Config == nil {
						this.Config = &data
				}
				if data, ok := arg.(*clientv3.Config); ok && this.Config == nil {
						this.Config = data
				}
				if client, ok := arg.(*clientv3.Client); ok && this.client == nil {
						this.client = client
				}
				if client, ok := arg.(clientv3.Client); ok && this.client == nil {
						this.client = &client
				}
		}
}

func (this *ConnectorImpl) initCnfByJson(data []byte) {
		this.Config = new(clientv3.Config)
		_ = json.Unmarshal(data, this.Config)
}

func (this *ConnectorImpl) Conn() *clientv3.Client {
		if this.client == nil {
				this.open()
		}
		return this.client
}

func (this *ConnectorImpl) Context() context.Context {
		ctx, fn := context.WithTimeout(context.Background(), this.getContextTimeout())
		return context.WithValue(ctx, ContextCancelFunc, fn)
}

func (this *ConnectorImpl) getContextTimeout() time.Duration {
		var t = os.Getenv(EnvEtcdCtxTimeout)
		if t == "" {
				return DefaultTimeOut
		}
		if d, err := time.ParseDuration(t); err == nil {
				return d
		}
		return DefaultTimeOut
}

func (this *ConnectorImpl) GetConfig() clientv3.Config {
		if this.Config == nil {
				this.defaults()
		}
		return *this.Config
}

func (this *ConnectorImpl) initByMapper(mapper map[string]interface{}) {
		if len(mapper) == 0 {
				return
		}
		if this.Config == nil {
				this.Config = new(clientv3.Config)
		}
		for key, it := range mapper {
				SetConfigure(this.Config, key, it)
		}
}

func (this *ConnectorImpl) SetConfig(config clientv3.Config) Connector {
		this.Config = &config
		return this
}

func (this *ConnectorImpl) defaults() {
		this.Config = &clientv3.Config{
				Endpoints:   this.getEndpoints(),
				DialTimeout: 2 * time.Second,
				Username:    os.Getenv(EnvEtcdUserName),
				Password:    os.Getenv(EnvEtcdPassword),
		}
}

func (this *ConnectorImpl) getEndpoints() []string {
		var endPoints = os.Getenv(EnvEtcdEndpoints)
		if endPoints == "" {
				return []string{DefaultEndPoints}
		}
		return strings.SplitN(endPoints, ",", -1)
}

func (this *ConnectorImpl) Close() {
		if this.client == nil {
				return
		}
		_ = this.client.Close()
		this.client = nil
}

// 设置配置
func SetConfigure(cnf *clientv3.Config, key string, v interface{}) {
		if cnf == nil {
				return
		}
		switch key {
		case "Context":
				fallthrough
		case "context":
				if ctx, ok := v.(context.Context); ok {
						cnf.Context = ctx
				}
		case "Username":
				fallthrough
		case "username":
				if username, ok := v.(string); ok {
						cnf.Username = username
				}
		case "Password":
				fallthrough
		case "password":
				if password, ok := v.(string); ok {
						cnf.Password = password
				}
		case "PermitWithoutStream":
				fallthrough
		case "permit_without_stream":
				fallthrough
		case "permit-without-stream":
				if b, ok := v.(bool); ok {
						cnf.PermitWithoutStream = b
				}
				if b, ok := v.(string); ok {
						if b == "1" || b == "true" || b == "True" {
								cnf.PermitWithoutStream = true
						}
						if b == "0" || b == "false" || b == "False" {
								cnf.PermitWithoutStream = false
						}
				}
		case "endpoints":
				fallthrough
		case "Endpoints":
				if endpoints, ok := v.(string); ok && endpoints != "" {
						cnf.Endpoints = strings.SplitN(endpoints, ",", -1)
				}
				if endpoints, ok := v.([]string); ok && len(endpoints) > 0 {
						cnf.Endpoints = endpoints
				}
		case "AutoSyncInterval":
				fallthrough
		case "auto_sync_interval":
				fallthrough
		case "auto-sync-interval":
				if t, ok := v.(time.Duration); ok {
						cnf.AutoSyncInterval = t
				}
				if t, ok := v.(string); ok {
						if d, err := time.ParseDuration(t); err == nil {
								cnf.AutoSyncInterval = d
						}
				}
		case "DialTimeout":
				fallthrough
		case "dial_timeout":
				fallthrough
		case "dial-timeout":
				if t, ok := v.(time.Duration); ok {
						cnf.DialTimeout = t
				}
				if t, ok := v.(string); ok {
						if d, err := time.ParseDuration(t); err == nil {
								cnf.DialTimeout = d
						}
				}
		case "LogConfig":
				fallthrough
		case "log_config":
				fallthrough
		case "log-config":
				if c, ok := v.(*zap.Config); ok {
						cnf.LogConfig = c
				}
				if c, ok := v.(zap.Config); ok {
						cnf.LogConfig = &c
				}
		case "DialKeepAliveTime":
				fallthrough
		case "dial_keep_alive_time":
				fallthrough
		case "dial-keep-alive-time":
				if t, ok := v.(time.Duration); ok {
						cnf.DialKeepAliveTime = t
				}
				if t, ok := v.(string); ok {
						if d, err := time.ParseDuration(t); err == nil {
								cnf.DialKeepAliveTime = d
						}
				}
		case "DialKeepAliveTimeout":
				fallthrough
		case "dial-keep-alive-timeout":
				fallthrough
		case "dial_keep_alive_timeout":
				if t, ok := v.(time.Duration); ok {
						cnf.DialKeepAliveTimeout = t
				}
				if t, ok := v.(string); ok {
						if d, err := time.ParseDuration(t); err == nil {
								cnf.DialKeepAliveTimeout = d
						}
				}
		case "MaxCallSendMsgSize":
				fallthrough
		case "max-call-send-msg-size":
				fallthrough
		case "max_call_send_msg_size":
				if n, ok := v.(int); ok {
						cnf.MaxCallSendMsgSize = n
				}
				if str, ok := v.(string); ok {
						if n, err := strconv.Atoi(str); err == nil && n > 0 {
								cnf.MaxCallSendMsgSize = n
						}
				}
		case "MaxCallRecvMsgSize":
				fallthrough
		case "max-call-recv-msg-size":
				fallthrough
		case "max_call_recv_msg_size":
				if n, ok := v.(int); ok {
						cnf.MaxCallRecvMsgSize = n
				}
				if str, ok := v.(string); ok {
						if n, err := strconv.Atoi(str); err == nil && n > 0 {
								cnf.MaxCallRecvMsgSize = n
						}
				}
		case "TLS":
				fallthrough
		case "tls":
				if t, ok := v.(tls.Config); ok {
						cnf.TLS = &t
				}
				if t, ok := v.(*tls.Config); ok {
						cnf.TLS = t
				}
		case "RejectOldCluster":
				fallthrough
		case "reject-old-cluster":
				fallthrough
		case "reject_old_cluster":
				if b, ok := v.(bool); ok {
						cnf.RejectOldCluster = b
				}
				if b, ok := v.(string); ok {
						if b == "1" || b == "true" || b == "True" {
								cnf.RejectOldCluster = true
						}
						if b == "0" || b == "false" || b == "False" {
								cnf.RejectOldCluster = false
						}
				}
		case "DialOptions":
				fallthrough
		case "dial_options":
				fallthrough
		case "dial-options":
				if opts, ok := v.([]grpc.DialOption); ok {
						cnf.DialOptions = opts
				}
		}
}
