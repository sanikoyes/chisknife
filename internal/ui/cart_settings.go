package ui

import (
	"chisknife/asset"
	"chisknife/internal/lang"
	"chisknife/internal/preset"
	"chisknife/internal/types"
	"chisknife/internal/utils"

	"github.com/AllenDang/giu"
)

type cartSettings struct {
	opts *types.BuildOptions
	tex  *giu.Texture
}

func newCartSettings(opts *types.BuildOptions) *cartSettings {
	return &cartSettings{
		opts: opts,
		tex:  utils.LoadTexture(asset.Background),
	}
}

func (ui *cartSettings) build() giu.Widget {
	opts := ui.opts

	cartTypeNames := preset.CartridgeTypes.Names()
	romSizeDescs := preset.RomSizes.Descs()

	return giu.Column(
		giu.Label(lang.L("Menu Settings")),
		giu.Separator(),

		giu.Row(
			giu.Label(lang.L("Cartridge Type")),
			giu.Combo("##cartridge_type", cartTypeNames[opts.Options.CartridgeType], cartTypeNames, &opts.Options.CartridgeType).Size(200),
		),

		giu.Row(
			giu.Label(lang.L("Minimal ROM Size")),
			giu.Combo("##rom_size", romSizeDescs[opts.Options.MinimalRomSize], romSizeDescs, &opts.Options.MinimalRomSize).Size(200),
		),

		giu.Checkbox(lang.L("Have Battery"), &opts.Options.HaveBattery),
		giu.Checkbox(lang.L("Use RTS"), &opts.Options.UseRTS),
		giu.Checkbox(lang.L("Split ROM"), &opts.Options.SplitROM),
		giu.Checkbox(lang.L("SRAM 1M save support"), &opts.Options.Sram1MSaveSupport),

		giu.Separator(),
		giu.Label(lang.L("Menu Background")),
		giu.Align(giu.AlignCenter).To(
			giu.Image(ui.tex).Size(240, 160),
			giu.Button(lang.L("Select the background")).OnClick(func() {
				// TODO: 实现背景选择
			}),
		),
	)
}
