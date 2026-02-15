package comm

import (
	"fmt"
	"sync"

	"github.com/emirpasic/gods/sets/hashset"
)

type PluginRegistryT struct {
	loaders               map[string]PluginLoader
	plugins               []Plugin
	pluginsByKind         map[PluginKind]map[string]Plugin
	supportedKinds        hashset.Set
	supportedMajorVersion int
	mutex                 sync.RWMutex
}

type PluginRegistry = *PluginRegistryT

func NewPluginRegistry(supportedMajorVersion int, supportedKinds ...PluginKind) PluginRegistry {
	r := &PluginRegistryT{
		loaders:               map[string]PluginLoader{},
		plugins:               []Plugin{},
		pluginsByKind:         map[PluginKind]map[string]Plugin{},
		supportedKinds:        *Slice2Set(supportedKinds...),
		supportedMajorVersion: supportedMajorVersion,
		mutex:                 sync.RWMutex{},
	}
	return r
}

func (me PluginRegistry) IsSupportedPluginKind(kind PluginKind) bool {
	return me.supportedKinds.Contains(kind)
}

func (me PluginRegistry) SupportedMajorVersion() int {
	return me.supportedMajorVersion
}

func (me PluginRegistry) ValidatePlugin(namespace string, plugin Plugin) error {
	name := plugin.Name()

	major, _ := plugin.Version()
	if major != me.supportedMajorVersion {
		return LocalizeError("error.plugin.version_mismatch", map[string]interface{}{
			"Namespace": namespace,
			"Name":      name,
			"Expected":  me.supportedMajorVersion,
			"Actual":    major,
		})
	}

	kind := plugin.Kind()
	if !me.IsSupportedPluginKind(kind) {
		return LocalizeError("error.plugin.unsupported_kind", map[string]interface{}{
			"Namespace": namespace,
			"Name":      name,
			"Kind":      kind,
		})
	}

	return nil
}

func (me PluginRegistry) ByKind(kind PluginKind) map[string]Plugin {
	me.mutex.RLock()
	defer me.mutex.RUnlock()

	return me.pluginsByKind[kind]
}

func (me PluginRegistry) Init(logger Logger) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	for ns, loader := range me.loaders {
		logCtx := NewLogContext(false)
		logCtx.Str("namespace", ns)
		subLogger := logger.NewSubLogger(logCtx)

		subLogger.Info().Msg(T("log.plugin.loader.starting", nil))
		err := loader.Start(logger)
		if err != nil {
			subLogger.Error(err).Msg(T("log.plugin.loader.start_failed", nil))
		} else {
			subLogger.Info().Msg(T("log.plugin.loader.started", nil))
		}
	}
}

func (me PluginRegistry) Destroy(logger Logger) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	for ns, loader := range me.loaders {
		logCtx := NewLogContext(false)
		logCtx.Str("namespace", ns)
		subLogger := logger.NewSubLogger(logCtx)

		subLogger.Info().Msg(T("log.plugin.loader.stopping", nil))
		err := loader.Stop(logger)
		if err != nil {
			subLogger.Error(err).Msg(T("log.plugin.loader.stop_failed", nil))
		} else {
			subLogger.Info().Msg(T("log.plugin.loader.stopped", nil))
		}
	}

	me.loaders = map[string]PluginLoader{}
	me.plugins = []Plugin{}
	me.pluginsByKind = map[PluginKind]map[string]Plugin{}
}

func (me PluginRegistry) HasNamespace(ns string) bool {
	_, r := me.loaders[ns]
	return r
}

func (me PluginRegistry) Register(loader PluginLoader) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	ns := loader.Namespace()
	if len(ns) == 0 {
		panic(LocalizeError("error.plugin.namespace_not_specified", map[string]interface{}{
			"Loader": fmt.Sprintf("%+v", loader),
		}))
	}

	if existingLoader, alreadyRegistered := me.loaders[ns]; alreadyRegistered {
		panic(LocalizeError("error.plugin.namespace_already_registered", map[string]interface{}{
			"Namespace": ns,
			"Loader":    fmt.Sprintf("%+v", existingLoader),
		}))
	}

	newPlugins := loader.Plugins()

	pluginsByKind := DeepCopyMap(me.pluginsByKind)

	for name, plugin := range newPlugins {
		if err := me.ValidatePlugin(ns, plugin); err != nil {
			panic(err)
		}

		kind := plugin.Kind()
		id := PluginId(ns, name)

		pluginsWithKind, found := pluginsByKind[kind]
		if !found {
			pluginsWithKind = map[string]Plugin{}
			pluginsByKind[kind] = pluginsWithKind
		}

		if existingPlugin, found := pluginsWithKind[name]; found {
			panic(LocalizeError("error.plugin.duplicated_kind", map[string]interface{}{
				"Namespace": id,
				"Kind":      kind,
				"Loader":    fmt.Sprintf("%+v", existingPlugin),
			}))
		}
		pluginsWithKind[name] = plugin
	}

	allPlugins := make([]Plugin, len(me.plugins), len(me.plugins)+len(newPlugins))
	allPlugins = append(allPlugins, me.plugins...)

	for _, plugin := range newPlugins {
		allPlugins = append(allPlugins, plugin)
	}

	me.plugins = allPlugins
	me.pluginsByKind = pluginsByKind
	me.loaders[ns] = loader
}
