package zkutil

import (
	"finder-go/common"
	"sync"
)

type OnServiceUpdateEvent func(common.Service)

type ServiceChangedEventPool struct {
	sync.RWMutex
	pool map[string]OnServiceUpdateEvent
}

func NewServiceChangedEventPool() *ServiceChangedEventPool {
	p := new(ServiceChangedEventPool)
	p.pool = make(map[string]OnServiceUpdateEvent)

	return p
}

func (p *ServiceChangedEventPool) Get() map[string]OnServiceUpdateEvent {
	p.RLock()
	defer p.RUnlock()
	return p.pool
}

func (p *ServiceChangedEventPool) Contains(key string) bool {
	p.RLock()
	defer p.RUnlock()
	if _, ok := p.pool[key]; ok {
		return true
	}

	return false
}

func (p *ServiceChangedEventPool) Append(key string, value OnServiceUpdateEvent) {
	p.Lock()
	p.pool[key] = value
	p.Unlock()
}

func (p *ServiceChangedEventPool) Remove(key string) {
	p.Lock()
	delete(p.pool, key)
	p.Unlock()
}
