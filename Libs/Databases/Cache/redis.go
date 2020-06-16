package Cache

import (
		"context"
		"crypto/tls"
		"encoding/json"
		"github.com/go-redis/redis/v8"
		"net"
		"os"
		"strconv"
		"time"
)

type RedisProvider interface {
		Conn(...string) *redis.Conn
		Context() context.Context
		Client(...string) *redis.Client
}

type ResultInterface interface {
		Result() (string, error)
}

type RedisInstance struct {
		connector *redis.Client
		info      *ConnOptions
}

type ConnOptions struct {
		redis.Options
}

const (
		EnvRedisDbNum      = "redis_db"
		EnvRedisAddr       = "redis_addr"
		EnvRedisCtxTimeout = "redis_ctx_timeout"
		EnvRedisPassword   = "redis_password"
		EnvRedisUserName   = "redis_username"
		DefaultRedisAddr   = "127.0.0.1:6379"
		StateSuccess       = "OK"
		ContextCancelFunc  = "CancelFunc"
		DefaultTimeOut     = 5 * time.Second
)

func NewConnOptions(args ...interface{}) *ConnOptions {
		var options = new(ConnOptions)
		options.init(args...)
		options.defaults()
		return options
}

func Redis(args ...interface{}) RedisProvider {
		var instance = new(RedisInstance)
		instance.init(args...)
		return instance
}

func (this *ConnOptions) init(args ...interface{}) {
		var (
				argc = len(args)
		)
		if argc == 0 {
				return
		}
		switch args[0].(type) {
		case string:
				this.initByJson(args[0].(string))
		case redis.Options:
				opt := args[0].(redis.Options)
				this.copy(&opt)
		case *redis.Options:
				this.copy(args[0].(*redis.Options))
		case map[string]interface{}:
				this.initByMapper(args[0].(map[string]interface{}))
		}
}

func (this *ConnOptions) initByJson(data string) {
		_ = json.Unmarshal([]byte(data), this)
}

func (this *ConnOptions) initByMapper(data map[string]interface{}) {
		for k, v := range data {
				this.set(k, v)
		}
}

func (this *ConnOptions) set(key string, v interface{}) {
		if key == "" || v == nil {
				return
		}
		switch key {
		case "username":
				fallthrough
		case "Username":
				this.Username = v.(string)
		case "password":
				fallthrough
		case "Password":
				this.Password = v.(string)
		case "addr":
				fallthrough
		case "Addr":
				this.Addr = v.(string)
		case "db":
				fallthrough
		case "Db":
				if db, ok := v.(string); ok {
						num, err := strconv.Atoi(db)
						if err == nil && num >= 0 {
								this.DB = num
						}
				}
				if db, ok := v.(int); ok && db >= 0 {
						this.DB = db
				}
		case "dialer":
				fallthrough
		case "Dialer":
				if fn, ok := v.(func(ctx context.Context, network, addr string) (net.Conn, error)); ok {
						this.Dialer = fn
				}
		case "OnConnect":
				fallthrough
		case "on_connect":
				if fn, ok := v.(func(ctx context.Context, cn *redis.Conn) error); ok {
						this.OnConnect = fn
				}
		case "pool_size":
				fallthrough
		case "PoolSize":
				if n, ok := v.(int); ok && n > 0 {
						this.PoolSize = n
				}
		case "PoolTimeout":
				fallthrough
		case "pool_timeout":
				if n, ok := v.(time.Duration); ok && n > 0 {
						this.PoolTimeout = n
				}
		case "MinIdleConns":
				fallthrough
		case "min_idle_conns":
				if n, ok := v.(int); ok && n > 0 {
						this.MinIdleConns = n
				}
		case "ReadTimeout":
				fallthrough
		case "read_timeout":
				if n, ok := v.(time.Duration); ok && n > 0 {
						this.ReadTimeout = n
				}
		case "WriteTimeout":
				fallthrough
		case "write_timeout":
				if n, ok := v.(time.Duration); ok && n > 0 {
						this.WriteTimeout = n
				}
		case "TLSConfig":
				fallthrough
		case "tls_config":
				if cnf, ok := v.(*tls.Config); ok {
						this.TLSConfig = cnf
				}
		case "Limiter":
				fallthrough
		case "limiter":
				if limiter, ok := v.(redis.Limiter); ok {
						this.Limiter = limiter
				}
		case "MinRetryBackoff":
				fallthrough
		case "min_retry_backoff":
				if d, ok := v.(time.Duration); ok {
						this.MinRetryBackoff = d
				}
		case "IdleTimeout":
				fallthrough
		case "idle_timeout":
				if d, ok := v.(time.Duration); ok {
						this.IdleTimeout = d
				}
		case "IdleCheckFrequency":
				fallthrough
		case "idle_check_frequency":
				if d, ok := v.(time.Duration); ok {
						this.IdleCheckFrequency = d
				}
		}
}

func (this *ConnOptions) copy(obj *redis.Options) {
		this.Addr = obj.Addr
		this.Username = obj.Username
		this.Password = obj.Password
		this.DB = obj.DB
		this.Dialer = obj.Dialer
		this.OnConnect = obj.OnConnect
		this.MaxRetries = obj.MaxRetries
		this.ReadTimeout = obj.ReadTimeout
		this.TLSConfig = obj.TLSConfig
		this.Limiter = obj.Limiter
		this.PoolSize = obj.PoolSize
		this.MinRetryBackoff = obj.MinRetryBackoff
		this.MaxConnAge = obj.MaxConnAge
		this.WriteTimeout = obj.WriteTimeout
		this.MinIdleConns = obj.MinIdleConns
		this.PoolTimeout = obj.PoolTimeout
		this.IdleCheckFrequency = obj.IdleCheckFrequency
		this.IdleTimeout = obj.IdleTimeout
}

func (this *ConnOptions) defaults() {
		if this.DB <= 0 {
				this.DB, _ = strconv.Atoi(os.Getenv(EnvRedisDbNum))
		}
		if this.Password == "" {
				this.Password = os.Getenv(EnvRedisPassword)
		}
		if this.Addr == "" {
				this.Addr = os.Getenv(EnvRedisAddr)
				if this.Addr == "" {
						this.Addr = DefaultRedisAddr
				}
		}
		if this.Username == "" {
				this.Username = os.Getenv(EnvRedisUserName)
		}
}

func (this *RedisInstance) init(args ...interface{}) {
		var (
				argc = len(args)
		)
		if argc == 0 {
				return
		}
		for _, arg := range args {
				if m, ok := arg.(map[string]interface{}); ok && this.info == nil {
						info := NewConnOptions(m)
						this.info = info
				}
				if client, ok := arg.(*redis.Client); ok && this.connector == nil {
						this.connector = client
				}
				if options, ok := arg.(*redis.Options); ok && this.info == nil {
						this.info = NewConnOptions(options)
				}
				if options, ok := arg.(*ConnOptions); ok && this.info == nil {
						this.info = options
				}
		}
}

func (this *RedisInstance) Client(dbs ...string) *redis.Client {
		var (
				argc    = len(dbs)
				options = this.getOptions()
				dbNum   = options.DB
		)
		if argc > 0 {
				if db, err := strconv.Atoi(dbs[0]); err == nil {
						if db != dbNum && db >= 0 {
								options.DB = dbNum
						}
				}
		}
		if this.connector == nil {
				this.connector = redis.NewClient(options)
		}
		return this.connector
}

func (this *RedisInstance) Conn(dbs ...string) *redis.Conn {
		return this.Client(dbs...).Conn(this.getContext())
}

func (this *RedisInstance) getContext() context.Context {
		if this.connector == nil {
				return nil
		}
		return this.connector.Context()
}

func (this *RedisInstance) getOptions() *redis.Options {
		if this.info != nil {
				return &this.info.Options
		}
		this.info = NewConnOptions()
		return &this.info.Options
}

func (this *RedisInstance) Close() {
		if this.connector == nil {
				return
		}
		_ = this.connector.Close()
		this.connector = nil
}

func (this *RedisInstance) Context() context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), this.getContextTimeout())
		return context.WithValue(ctx, ContextCancelFunc, cancel)
}

func (this *RedisInstance) getContextTimeout() time.Duration {
		var t = os.Getenv(EnvRedisCtxTimeout)
		if t == "" {
				return DefaultTimeOut
		}
		if d, err := time.ParseDuration(t); err == nil {
				return d
		}
		return DefaultTimeOut
}

func IsSuccess(result ResultInterface) bool {
		if res, err := result.Result(); err == nil {
				if _, ok := result.(*redis.StatusCmd); ok && res == StateSuccess {
						return true
				}
				if res == "" {
						return false
				}
		}
		return false
}

func GetResult(result ResultInterface) string {
		res, err := result.Result()
		if err == nil {
				return res
		}
		return ""
}
