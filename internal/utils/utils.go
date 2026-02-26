// Package utils 提供通用的工具函数
// 包括图像加载、纹理处理等辅助功能
package utils

import (
	"bytes"
	"image/png"

	"github.com/AllenDang/giu"
)

// 从字节数据加载 PNG 图像并创建 GUI 纹理
// 用于在界面中显示图片资源
func LoadTexture(data []byte) *giu.Texture {
	var tex *giu.Texture
	b := bytes.NewReader(data)
	if img, err := png.Decode(b); err == nil {
		giu.NewTextureFromRgba(img, func(t *giu.Texture) {
			tex = t
		})
	}
	return tex
}
