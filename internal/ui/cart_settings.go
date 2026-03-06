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
	project   *types.Project // 打包工程数据
	tex       *giu.Texture   // 背景图片纹理
	texLoaded bool           // 纹理是否已加载
}

// 创建卡带设置组件实例
// 加载背景图片纹理并初始化配置
func newCartSettings(project *types.Project) *cartSettings {
	return &cartSettings{
		project:   project,
		texLoaded: false,
	}
}

// 刷新界面状态
func (ui *cartSettings) refresh() {
	// 卡带设置界面会自动从 project 读取最新数据
	// 无需额外操作
}

// 构建卡带设置界面
// 包括卡带类型、ROM 大小、各种选项和背景图片设置
func (ui *cartSettings) build() giu.Widget {
	opts := &ui.project.Options

	// 延迟加载纹理，确保在渲染上下文中加载
	if !ui.texLoaded {
		ui.tex = utils.LoadTexture(asset.Background)
		ui.texLoaded = true
	}

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
