package comm

import (
	"fmt"

	"github.com/pkg/errors"
)

type PluginLang = string

const (
	PLUGIN_LANG_GO         = "go"
	PLUGIN_LANG_JAVASCRIPT = "javascript"
	PLUGIN_LANG_SHELL      = "shell"
)

type PluginKind = string

type Plugin interface {
	Name() string
	Kind() PluginKind
	Start(logger Logger)
	Stop(logger Logger)
	Version() (major int, minor int)
}

type PluginLoader interface {
	Namespace() string
	Plugins() map[string]Plugin
	Start(logger Logger) error
	Stop(logger Logger) error
}

func PluginId(namespace string, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func StartPlugin(namespace string, plugin Plugin, logger Logger) (err error) {
	major, minor := plugin.Version()
	ver := fmt.Sprintf("%d/%d", major, minor)
	pluginId := PluginId(namespace, plugin.Name())

	defer func() {
		if p := recover(); p != nil {
			var err2 error
			var isErr bool
			if err2, isErr = p.(error); isErr {
				err = errors.Wrap(err2, T("error.plugin.start_failed", map[string]interface{}{
					"PluginId": pluginId,
					"Version":  ver,
					"Cause":    err2,
				}))
			} else {
				err = LocalizeError("error.plugin.start_failed", map[string]interface{}{
					"PluginId": pluginId,
					"Version":  ver,
					"Cause":    p,
				})
			}
		}
	}()

	logCtx := NewLogContext(false)
	logCtx.Str("pluginId", pluginId).Str("version", ver)
	subLogger := logger.NewSubLogger(logCtx)

	subLogger.Info().Msg("starting")
	plugin.Start(logger)
	subLogger.Info().Msg("started")

	return err
}

func StopPlugin(namespace string, plugin Plugin, logger Logger) (err error) {
	major, minor := plugin.Version()
	ver := fmt.Sprintf("%d/%d", major, minor)
	pluginId := PluginId(namespace, plugin.Name())

	defer func() {
		if p := recover(); p != nil {
			var err2 error
			var isErr bool
			if err2, isErr = p.(error); isErr {
				err = errors.Wrap(err2, T("error.plugin.stop_failed", map[string]interface{}{
					"PluginId": pluginId,
					"Version":  ver,
					"Cause":    err2,
				}))
			} else{
				err = LocalizeError("error.plugin.stop_failed", map[string]interface{}{
					"PluginId": pluginId,
					"Version":  ver,
					"Cause":    p,
				})
			}
		}
	}()

	logCtx := NewLogContext(false)
	logCtx.Str("pluginId", pluginId).Str("version", ver)
	subLogger := logger.NewSubLogger(logCtx)

	subLogger.Info().Msg("stopping")
	plugin.Stop(logger)
	subLogger.Info().Msg("stopped")

	return err
}
