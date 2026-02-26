package main

import (
	"chisknife/internal/lang"
	"chisknife/internal/ui"

	"github.com/AllenDang/giu"
)

func main() {
	wnd := giu.NewMasterWindow(lang.L("Chis Army Knife"), 640, 480, 0)
	wnd.Run(ui.Loop)
}
