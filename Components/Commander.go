package Components

import "os"

// 命令行结果体
type CommanderImpl struct {
	title      string
	args       []string
	options    map[string]string
	helpMenu   string
	optionArgs map[string]*OptionArg
}

func (this *CommanderImpl) Init() {
	this.args = os.Args
	this.initOptions()
	this.initHelp()
	this.title = ""
}

func (this *CommanderImpl) initOptions() {
	var options = make(map[string]string)
	this.options = options
}

func (this *CommanderImpl) initHelp() {
	this.helpMenu = ""
}

func (this *CommanderImpl) Exec() int {
	panic("implement me")
}

func (this *CommanderImpl) Argc() int {
	return len(this.args)
}

func (this *CommanderImpl) Help() string {
	return this.helpMenu
}

func (this *CommanderImpl) Title() string {
	return this.title
}

func (this *CommanderImpl) Args() []string {
	return this.args
}

func (this *CommanderImpl) Option(key string) string {
	panic("implement me")
}

func (this *CommanderImpl) Options(key string) []string {
	panic("implement me")
}

func (this *CommanderImpl) GetOptions() map[string]*OptionArg {
	panic("implement me")
}

func CommanderOf() Commander {
	var command = new(CommanderImpl)
	return command
}
