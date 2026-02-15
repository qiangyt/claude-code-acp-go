package comm

import (
	"sync"
)

type BasePluginT struct {
	name    string
	kind    PluginKind
	started bool

	mutex sync.RWMutex
}

type BasePlugin = *BasePluginT

func NewBasePlugin(name string, kind PluginKind) BasePluginT {
	return BasePluginT{
		name:    name,
		kind:    kind,
		started: false,
	}
}

func (me BasePlugin) Name() string {
	return me.name
}

func (me BasePlugin) Kind() PluginKind {
	return me.kind
}

func (me BasePlugin) Start(logger Logger) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	if me.started {
		return
	}
	me.started = true
}

func (me BasePlugin) IsStarted() bool {
	me.mutex.RLock()
	defer me.mutex.RUnlock()

	return me.started
}

func (me BasePlugin) Stop(logger Logger) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	if !me.started {
		return
	}
	me.started = false
}

func (me BasePlugin) Version() (major int, minor int) {
	return 1, 0
}
