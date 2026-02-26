// Package lang 提供国际化和本地化支持
// 负责加载和管理多语言翻译
package lang

import (
	"chisknife/asset"
	"encoding/json"
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

// 从嵌入的文件系统加载翻译数据
// 优先加载中文翻译，失败时回退到英文
func loadTranslations() {
	data, err := asset.TranslationsFS.ReadFile("translations/zh-CN.json")
	if err != nil {
		data, _ = asset.TranslationsFS.ReadFile("translations/en.json")
	}
	json.Unmarshal(data, &translations)
}

// 在包初始化时自动加载翻译数据
func init() {
	loadTranslations()
}
