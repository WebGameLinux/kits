package Components

import (
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Libs"
		"sync"
)

type LoggerProvider interface {
		Contracts.Provider
		Logger
}

type Logger interface {
		SetLevel(string)
		Error(args ...interface{})
		Debug(args ...interface{})
		Info(args ...interface{})
		Warn(args ...interface{})
}

type LoggerProviderImpl struct {
		Name     string
		app      Contracts.ApplicationContainer
		clazz    Contracts.ClazzInterface
		bean     Contracts.SupportInterface
		instance Logger
}

const (
		LoggerAlias         = "logger"
		LoggerProviderClass = "LoggerProvider"
)

var (
		loggerInstanceLock sync.Once
		loggerInstance     *LoggerProviderImpl
)

func LoggerProviderOf() LoggerProvider {
		if loggerInstance == nil {
				loggerInstanceLock.Do(loggerProviderNew)
		}
		return loggerInstance
}

func loggerProviderNew() {
		loggerInstance = new(LoggerProviderImpl)
		loggerInstance.init()
}

func (this *LoggerProviderImpl) GetClazz() Contracts.ClazzInterface {
		if this.clazz == nil {
				this.clazz = ClazzOf(this)
		}
		return this.clazz
}

func (this *LoggerProviderImpl) Init(app Contracts.ApplicationContainer) {
		if this.app == nil {
				this.app = app
		}
		this.initBase()
}

func (this *LoggerProviderImpl) initLoggerBird() {
		obj := this.app.Get(ConfigureProviderClass)
		if configure, ok := obj.(ConfigureProvider); ok {
				cnf := configure.Any(LoggerAlias)
				if args, ok := cnf.([]interface{}); ok {
						this.instance = Libs.NewLoggerBird(args...)
				} else {
						this.instance = Libs.NewLoggerBird(cnf)
				}
		}
}

func (this *LoggerProviderImpl) initBase() {
		this.GetClazz()
		this.GetSupportBean()
}

func (this *LoggerProviderImpl) GetSupportBean() Contracts.SupportInterface {
		if this.bean == nil {
				this.bean = BeanOf()
		}
		return this.bean
}

func (this *LoggerProviderImpl) Register() {
		this.app.Bind(this.String(), this)
		this.app.Singleton(LoggerAlias, this.getLoggerInstance)
}

func (this *LoggerProviderImpl) getLoggerInstance(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		this.initLoggerBird()
		return this.instance
}

func (this *LoggerProviderImpl) Boot() {
		// read all
		// start watch
}

func (this *LoggerProviderImpl) String() string {
		return this.Name
}

func (this *LoggerProviderImpl) SetLevel(level string) {
		this.logger().SetLevel(level)
}

func (this *LoggerProviderImpl) Error(args ...interface{}) {
		this.logger().Error(args...)
}

func (this *LoggerProviderImpl) Debug(args ...interface{}) {
		this.logger().Debug(args...)
}

func (this *LoggerProviderImpl) Info(args ...interface{}) {
		this.logger().Info(args...)
}

func (this *LoggerProviderImpl) Warn(args ...interface{}) {
		this.logger().Warn(args...)
}

func (this *LoggerProviderImpl) Factory(app Contracts.ApplicationContainer) interface{} {
		logger := this.Constructor()
		if loggerProvider, ok := logger.(*LoggerProviderImpl); ok {
				loggerProvider.Init(app)
				return loggerProvider.instance
		}
		return this.instance
}

func (this *LoggerProviderImpl) Constructor() interface{} {
		return LoggerProviderOf()
}

func (this *LoggerProviderImpl) init() {
		this.Name = LoggerProviderClass
}

func (this *LoggerProviderImpl) logger() Logger {
		if logger, ok := this.app.Get(LoggerAlias).(Logger); ok {
				return logger
		}
		return this.instance
}
