package Components

import (
		"github.com/webGameLinux/kits/Contracts"
		"sync"
)

type EventBusProvider interface {
		Contracts.Provider
}

type EventBusProviderImpl struct {
		AppServiceProvider
}

var (
		eventInstanceLock sync.Once
		eventBus          *EventBusProviderImpl
)

const (
		EventBusProviderClass = "EventBusProvider"
)

func newEventBus() {
		eventBus = new(EventBusProviderImpl)
		eventBus.Name = EventBusProviderClass
}

func EventBusProviderOf() EventBusProvider {
		if eventBus == nil {
				eventInstanceLock.Do(newEventBus)
		}
		return eventBus
}

func (this *EventBusProviderImpl) Constructor() interface{} {
		return EventBusProviderOf()
}

func (this *EventBusProviderImpl) Factory(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		return this
}

func (this *EventBusProviderImpl) Register() {

}

func (this *EventBusProviderImpl) Boot() {

}
