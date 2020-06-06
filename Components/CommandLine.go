package Components

import (
	"github.com/webGameLinux/kits/Contracts"
)

// 命令行服务
type CommandLineArgsProvider interface {
	Contracts.Provider
	Argc() int
	Args() []string
	Option(string) string
	Options(string) []string
	GetOptions() map[string]*OptionArg
}

// 命令行工具设置解析器
type Commander interface {
	Init()
	Exec() int
	Argc() int
	Help() string
	Title() string
	Args() []string
	Option(string) string
	Options(string) []string
	GetOptions() map[string]*OptionArg
}

// 参数值
type OptionArg struct {
	Alias  []string
	Values []string
}

// 命令行服务
type CommandLineArgsProviderImpl struct {
	command Commander
	bean    *Contracts.SupportBean
	clazz   Contracts.ClazzInterface
	app     Contracts.ApplicationContainer
}

func CommandLineArgsProviderOf() CommandLineArgsProvider {
	var command = new(CommandLineArgsProviderImpl)
	return command
}

func (this *CommandLineArgsProviderImpl) GetClazz() Contracts.ClazzInterface {
	if this.clazz == nil {
		this.clazz = ClazzOf(this)
	}
	return this.clazz
}

func (this *CommandLineArgsProviderImpl) Init(app Contracts.ApplicationContainer) {
	this.app = app
	this.command = CommanderOf()
}

func (this *CommandLineArgsProviderImpl) GetSupportBean() Contracts.SupportBean {
	if this.bean == nil {
		this.bean = BeanOf()
	}
	return *this.bean
}

func (this *CommandLineArgsProviderImpl) Register() {
	// 解析commander
	this.command.Init()
	this.app.Bind(this.GetClazz().String(), this.command)
}

func (this *CommandLineArgsProviderImpl) Boot() {
	// 解析运行 commander 参数
	commander := this.app.Get(this.String())
	if commander == nil {
		return
	}
	// 运行命令行
	if c, ok := commander.(Commander); ok {
		if c.Exec() < 0 {
			this.app.Stop()
		}
	}
}

func (this *CommandLineArgsProviderImpl) Argc() int {
	return this.command.Argc()
}

func (this *CommandLineArgsProviderImpl) Args() []string {
	return this.command.Args()
}

func (this *CommandLineArgsProviderImpl) Option(key string) string {
	return this.command.Option(key)
}

func (this *CommandLineArgsProviderImpl) Options(key string) []string {
	return this.command.Options(key)
}

func (this *CommandLineArgsProviderImpl) GetOptions() map[string]*OptionArg {
	return this.command.GetOptions()
}

func (this *CommandLineArgsProviderImpl) Factory(app Contracts.ApplicationContainer) interface{} {
	this.Init(app)
	return this
}

func (this *CommandLineArgsProviderImpl) Constructor() interface{} {
	return CommandLineArgsProviderOf()
}

func (this *CommandLineArgsProviderImpl) String() string {
	return this.GetClazz().String()
}
