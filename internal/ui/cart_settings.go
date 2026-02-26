// Package ui 提供应用程序的图形用户界面
package ui

import (
	"chisknife/asset"
	"chisknife/internal/lang"
	"chisknife/internal/preset"
	"chisknife/internal/types"
	"chisknife/internal/utils"

	"github.com/AllenDang/giu"
)

// 卡带设置界面组件
type cartSettings struct {
	opts *types.Options // 构建选项引用
	tex  *giu.Texture   // 背景图片纹理
}

// 创建卡带设置组件实例
// 加载背景图片纹理并初始化配置
func newCartSettings(opts *types.Options) *cartSettings {
	return &cartSettings{
		opts: opts,
		tex:  utils.LoadTexture(asset.Background),
	}
}

// 构建卡带设置界面
// 包括卡带类型、ROM 大小、各种选项和背景图片设置
func (ui *cartSettings) build() giu.Widget {
	opts := ui.opts

	cartTypeNames := preset.CartridgeTypes.Names()
	romSizeDescs := preset.RomSizes.Descs()

	return giu.Column(
		giu.Label(lang.L("Menu Settings")),
		giu.Separator(),

		giu.Row(
			giu.Label(lang.L("Cartridge Type")),
			giu.Combo("##cartridge_type", cartTypeNames[opts.CartridgeType], cartTypeNames, &opts.CartridgeType).Size(200),
		),

		giu.Row(
			giu.Label(lang.L("Minimal ROM Size")),
			giu.Combo("##rom_size", romSizeDescs[opts.MinimalRomSize], romSizeDescs, &opts.MinimalRomSize).Size(200),
		),

		giu.Checkbox(lang.L("Have Battery"), &opts.HaveBattery),
		giu.Checkbox(lang.L("Use RTS"), &opts.UseRTS),
		giu.Checkbox(lang.L("Split ROM"), &opts.SplitROM),
		giu.Checkbox(lang.L("SRAM 1M save support"), &opts.Sram1MSaveSupport),

		giu.Separator(),
		giu.Label(lang.L("Menu Background")),
		giu.Align(giu.AlignCenter).To(
			giu.Image(ui.tex).Size(240, 160),
			giu.Button(lang.L("Select the background")).OnClick(func() {
				// TODO: 实现背景图片选择功能
			}),
		),
	)
}
