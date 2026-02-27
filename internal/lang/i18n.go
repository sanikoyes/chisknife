// Package lang 提供国际化和本地化支持
// 负责加载和管理多语言翻译
package lang

import (
	"chisknife/asset"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// 存储当前语言的所有翻译键值对
var translations map[string]string

// 根据键获取对应的翻译文本
// 如果找不到翻译，则返回原始键值
func L(key string) string {
	if v, ok := translations[key]; ok {
		return v
	}
	return key
}

// 获取系统locale
func getLocale() (string, string) {
	osHost := runtime.GOOS
	defaultLang := "en"
	defaultLoc := "US"
	switch osHost {
	case "windows":
		// Exec powershell Get-Culture on Windows.
		cmd := exec.Command("powershell", "Get-Culture | select -exp Name")
		output, err := cmd.Output()
		if err == nil {
			langLocRaw := strings.TrimSpace(string(output))
			langLoc := strings.Split(langLocRaw, "-")
			lang := langLoc[0]
			loc := langLoc[1]
			return lang, loc
		}
	case "darwin":
		// Exec shell Get-Culture on MacOS.
		cmd := exec.Command("sh", "osascript -e 'user locale of (get system info)'")
		output, err := cmd.Output()
		if err == nil {
			langLocRaw := strings.TrimSpace(string(output))
			langLoc := strings.Split(langLocRaw, "_")
			lang := langLoc[0]
			loc := langLoc[1]
			return lang, loc
		}
	case "linux":
		envlang, ok := os.LookupEnv("LANG")
		if ok {
			langLocRaw := strings.TrimSpace(envlang)
			langLocRaw = strings.Split(envlang, ".")[0]
			langLoc := strings.Split(langLocRaw, "_")
			lang := langLoc[0]
			loc := langLoc[1]
			return lang, loc
		}
	}
	return defaultLang, defaultLoc
}

// 从嵌入的文件系统加载翻译数据
// 优先加载中文翻译，失败时回退到英文
func loadTranslations() {
	var data []byte

	if lang, loc := getLocale(); lang != "en" {
		data, _ = asset.TranslationsFS.ReadFile(fmt.Sprintf("translations/%s-%s.json", lang, loc))
	}

	if len(data) == 0 {
		data, _ = asset.TranslationsFS.ReadFile("translations/en.json")
	}

	json.Unmarshal(data, &translations)
}

// 在包初始化时自动加载翻译数据
func init() {
	loadTranslations()
}
