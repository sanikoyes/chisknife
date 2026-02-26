package utils

import (
	"bytes"
	"image/png"

	"github.com/AllenDang/giu"
)

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
