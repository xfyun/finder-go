package zkutil

import (
	"finder-go/common"
	"sync"
)

type OnCfgUpdateEvent func(common.Config)

type ConfigChangedEventPool struct {
	sync.RWMutex
	pool map[string]OnCfgUpdateEvent
}

func NewConfigChangedEventPool() *ConfigChangedEventPool {
	p := new(ConfigChangedEventPool)
	p.pool = make(map[string]OnCfgUpdateEvent)

	return p
}

func (p *ConfigChangedEventPool) Get() map[string]OnCfgUpdateEvent {
	p.RLock()
	defer p.RUnlock()
	return p.pool
}

func (p *ConfigChangedEventPool) Contains(key string) bool {
	p.RLock()
	defer p.RUnlock()
	if _, ok := p.pool[key]; ok {
		return true
	}

	return false
}

func (p *ConfigChangedEventPool) Append(key string, value OnCfgUpdateEvent) {
	p.Lock()
	p.pool[key] = value
	p.Unlock()
}

func (p *ConfigChangedEventPool) Remove(key string) {
	p.Lock()
	delete(p.pool, key)
	p.Unlock()
}
