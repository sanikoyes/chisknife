package ui

import (
	"chisknife/internal/types"

	"github.com/AllenDang/giu"
)

type mainWindow struct {
	opts         *types.BuildOptions
	sashPos      float32
	cartSettings *cartSettings
}

func newMainWindow() *mainWindow {
	opts := &types.BuildOptions{
		Options: types.Options{
			CartridgeType:     0,
			MinimalRomSize:    0,
			HaveBattery:       true,
			UseRTS:            false,
			SplitROM:          false,
			Sram1MSaveSupport: false,
		},
		RomList: types.RomList{
			Roms: []string{"1", "2", "3"},
		},
	}

	return &mainWindow{
		sashPos:      320,
		opts:         opts,
		cartSettings: newCartSettings(opts),
	}
}

func (ui *mainWindow) build() {
	opts := ui.opts
	giu.SingleWindow().RegisterKeyboardShortcuts().Layout(
		giu.SplitLayout(
			giu.DirectionVertical,
			&ui.sashPos,
			buildRomList(opts),
			ui.cartSettings.build(),
		),
	)
}

var w *mainWindow

func Loop() {
	if w == nil {
		w = newMainWindow()
	}
	w.build()
}
