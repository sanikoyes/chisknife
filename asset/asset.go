// Package asset 提供应用程序的嵌入式资源管理
// 包括背景图片和多语言翻译文件
package asset

import (
	embed "embed"
)

// 嵌入的背景图片数据
// 用于菜单背景显示
//
//go:embed bg.png
var Background []byte

// 嵌入的翻译文件系统
// 包含所有支持语言的 JSON 翻译文件
//
//go:embed translations/*.json
var TranslationsFS embed.FS
