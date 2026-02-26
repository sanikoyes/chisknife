package lang

import (
	"chisknife/asset"
	"encoding/json"
)

var translations map[string]string

func L(key string) string {
	if v, ok := translations[key]; ok {
		return v
	}
	return key
}

func loadTranslations() {
	data, err := asset.TranslationsFS.ReadFile("translations/zh-CN.json")
	if err != nil {
		data, _ = asset.TranslationsFS.ReadFile("translations/en.json")
	}
	json.Unmarshal(data, &translations)
}

func init() {
	loadTranslations()
}
