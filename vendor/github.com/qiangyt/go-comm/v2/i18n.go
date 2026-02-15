package comm

import (
	"embed"
	"fmt"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/afero"
	"golang.org/x/text/language"
)

//go:embed locales/*.toml
var localesFS embed.FS

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	mutex     sync.RWMutex
	currLang  string // 当前语言代码
)

// 支持的语言标签映射
var langTags = map[string]language.Tag{
	"zh": language.SimplifiedChinese, // 中文
	"en": language.English,            // 英语
	"ru": language.Russian,            // 俄语
	"fr": language.French,             // 法语
	"es": language.Spanish,            // 西班牙语
	"it": language.Italian,            // 意大利语
	"de": language.German,             // 德语
	"hu": language.Hungarian,          // 匈牙利语
	"ko": language.Korean,             // 韩语
	"ja": language.Japanese,           // 日语
	"vi": language.Vietnamese,         // 越南语
	"th": language.Thai,               // 泰语
	"id": language.Indonesian,         // 印尼语
}

// 默认语言（英语）
const defaultLang = "en"

func init() {
	// Initialize with detected language
	InitI18n(DetectLanguage())
}

// InitI18n initializes the i18n system with the specified language.
// Supported languages: zh, en, ru, fr, es, it, de, hu, ko, ja, vi, th, id
// If an unsupported language is provided, it defaults to "en".
func InitI18n(lang string) {
	mutex.Lock()
	defer mutex.Unlock()

	// 如果语言不支持，默认使用英语
	if _, ok := langTags[lang]; !ok {
		lang = "en"
	}

	currLang = lang

	// 使用英语作为基础语言创建 bundle（这样所有语言的翻译都可以被加载）
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	// Load embedded locale files
	loadEmbeddedLocales()

	// Create localizer with fallback to English
	localizer = i18n.NewLocalizer(bundle, lang, "en")
}

// SetLanguage changes the current language for i18n.
func SetLanguage(lang string) {
	InitI18n(lang)
}

// DetectLanguage detects the user's language from environment variables or default.
// 检测顺序: LANGUAGE 环境变量 > LANG 环境变量 > 默认英语
func DetectLanguage() string {
	lang := os.Getenv("LANGUAGE")
	if lang == "" {
		lang = os.Getenv("LANG")
	}
	if lang == "" {
		return defaultLang
	}

	// Parse language code (e.g., "zh_CN.UTF-8" -> "zh")
	if len(lang) >= 2 {
		code := lang[:2]
		// 检查是否支持该语言
		if _, ok := langTags[code]; ok {
			return code
		}
	}
	return defaultLang
}

// GetLanguageTag returns the language.Tag for the given language code.
// If the language is not supported, returns English tag.
func GetLanguageTag(lang string) language.Tag {
	if tag, ok := langTags[lang]; ok {
		return tag
	}
	return language.English
}

// GetLanguage returns the current language code
func GetLanguage() string {
	mutex.RLock()
	defer mutex.RUnlock()

	if currLang == "" {
		return defaultLang
	}
	return currLang
}

func loadEmbeddedLocales() {
	// 加载所有语言的翻译文件
	langFiles := []string{
		"locales/active.zh.toml",
		"locales/active.en.toml",
		"locales/active.ru.toml",
		"locales/active.fr.toml",
		"locales/active.es.toml",
		"locales/active.it.toml",
		"locales/active.de.toml",
		"locales/active.hu.toml",
		"locales/active.ko.toml",
		"locales/active.ja.toml",
		"locales/active.vi.toml",
		"locales/active.th.toml",
		"locales/active.id.toml",
	}

	for _, file := range langFiles {
		if data, err := localesFS.ReadFile(file); err == nil {
			bundle.MustParseMessageFileBytes(data, file)
		}
	}
}

// T translates a message ID with optional template data.
// Usage:
//   T("error.required", map[string]interface{}{"Hint": "config", "Key": "name"})
func T(messageID string, templateData map[string]interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()

	if localizer == nil {
		return messageID
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		// Fallback to messageID if translation is not found
		return messageID
	}
	return msg
}

// Tf translates a message ID with formatted arguments (printf-style).
// This is a convenience function for simple string formatting.
func Tf(messageID string, args ...interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()

	if localizer == nil {
		return fmt.Sprintf(messageID, args...)
	}

	// Try to get translation first
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		// If translation not found, use messageID as format string
		return fmt.Sprintf(messageID, args...)
	}

	// Apply formatting to translated message
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// LocalizeError creates a localized error message.
func LocalizeError(messageID string, templateData map[string]interface{}) error {
	return fmt.Errorf("%s", T(messageID, templateData))
}

// LocalizeErrorf creates a localized error message with printf-style formatting.
func LocalizeErrorf(messageID string, args ...interface{}) error {
	return fmt.Errorf("%s", Tf(messageID, args...))
}

// localize localizes a message with optional template data.
// This is the unified localization function used by FileOps and other components.
func localize(id string, args ...map[string]any) string {
	mutex.RLock()
	defer mutex.RUnlock()

	cfg := &i18n.LocalizeConfig{
		MessageID:   id,
		PluralCount: 2,
	}
	if len(args) > 0 {
		cfg.TemplateData = args[0]
	}

	r, err := localizer.Localize(cfg)
	if err != nil {
		return id
	}
	return r
}

// LocalizeFunc is the i18n localization function type
type LocalizeFunc func(id string, args ...map[string]any) string

var (
	// DefaultLocalizeFunc is the default localization function (returns id directly)
	DefaultLocalizeFunc LocalizeFunc = func(id string, args ...map[string]any) string {
		return id
	}
)

// NewLocalizedFileOps creates a FileOps with built-in i18n support
func NewLocalizedFileOps(fs afero.Fs) FileOps {
	ops := NewFileOps(fs)
	ops.SetLocalizeFunc(localize)
	return ops
}

// CommLocalize is an alias for localize, for backward compatibility
func CommLocalize(id string, args ...map[string]any) string {
	return localize(id, args...)
}

// SetCommLang is an alias for SetLanguage, for backward compatibility
func SetCommLang(lang string) {
	SetLanguage(lang)
}