package Components

import "github.com/webGameLinux/kits/Contracts"

type LoggerProvider interface {
		Contracts.Provider
		Logger
}

type Logger interface {
		SetLevel(string)
		Error(args...interface{})
		Debug(args...interface{})
		Info(args...interface{})
		Warn(args...interface{})
}

type LoggerProviderImpl struct {

}


