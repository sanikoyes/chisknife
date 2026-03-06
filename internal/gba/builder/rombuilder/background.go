package rombuilder

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// updateBackground 更新背景图片
func updateBackground(menuROM []byte, bgPath string) error {
	// 确定背景图片路径
	if bgPath == "" {
		bgPath = "bg.png"
	}

	// 检查文件是否存在
	if !fileExists(bgPath) {
		return fmt.Errorf("background image not found: %s", bgPath)
	}

	// 读取 PNG 图片
	f, err := os.Open(bgPath)
	if err != nil {
		return fmt.Errorf("failed to open background image: %w", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode PNG: %w", err)
	}

	// 转换为调色板图像
	paletted, palette := convertToPaletted(img)

	// 转换调色板为 RGB555
	paletteRGB555 := convertPaletteToRGB555(palette)

	// 获取原始位图数据
	rawBitmap := paletted

	// 生成调色板数据
	rawPalette := make([]byte, 0x200)
	for i, color := range paletteRGB555 {
		if i >= 256 {
			break
		}
		binary.LittleEndian.PutUint16(rawPalette[i*2:], color)
	}

	// 查找背景偏移
	marker := []byte("RTFN\xff\xfe")
	bgOffset := bytes.Index(menuROM, marker)
	if bgOffset == -1 {
		return fmt.Errorf("background marker not found in menu ROM")
	}
	bgOffset -= 0x9800

	// 更新背景数据
	if bgOffset+0x9800 > len(menuROM) {
		return fmt.Errorf("background offset out of range")
	}

	copy(menuROM[bgOffset:bgOffset+0x9600], rawBitmap)
	copy(menuROM[bgOffset+0x9600:bgOffset+0x9800], rawPalette)

	return nil
}

// convertToPaletted 转换图像为调色板格式
func convertToPaletted(img image.Image) ([]byte, []color.Color) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 收集所有颜色
	colorMap := make(map[color.Color]int)
	palette := make([]color.Color, 0, 256)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			if _, exists := colorMap[c]; !exists {
				if len(palette) < 256 {
					colorMap[c] = len(palette)
					palette = append(palette, c)
				}
			}
		}
	}

	// 创建索引位图
	bitmap := make([]byte, width*height)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			if colorIdx, exists := colorMap[c]; exists {
				bitmap[idx] = byte(colorIdx)
			}
			idx++
		}
	}

	return bitmap, palette
}

// convertPaletteToRGB555 转换调色板为 RGB555 格式
func convertPaletteToRGB555(palette []color.Color) []uint16 {
	rgb555 := make([]uint16, len(palette))
	for i, c := range palette {
		r, g, b, _ := c.RGBA()
		// 转换为 8 位
		r8 := uint16(r >> 8)
		g8 := uint16(g >> 8)
		b8 := uint16(b >> 8)
		// 转换为 5 位并组合为 RGB555
		r5 := r8 >> 3
		g5 := g8 >> 3
		b5 := b8 >> 3
		rgb555[i] = (b5 << 10) | (g5 << 5) | r5
	}
	return rgb555
}
