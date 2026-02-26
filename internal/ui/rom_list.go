package ui

import (
	"chisknife/internal/lang"
	"chisknife/internal/types"

	"github.com/AllenDang/giu"
)

var selectedRomIndex int32 = -1

func buildRomList(opts *types.BuildOptions) giu.Widget {
	return giu.Column(
		giu.Label(lang.L("ROM List")),
		giu.Separator(),
		giu.ListBox(opts.RomList.Roms).OnChange(func(selectedIndex int) {
		}).ID("###roms"),
	)
}
