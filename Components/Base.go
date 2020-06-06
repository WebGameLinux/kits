package Components

import "github.com/webGameLinux/kits/Contracts"

func BeanOf() *Contracts.SupportBean {
	var bean = new(Contracts.SupportBean)
	bean.Boot = true
	bean.Register = true
	return bean
}
