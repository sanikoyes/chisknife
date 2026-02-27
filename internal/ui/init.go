package ui

import "github.com/AllenDang/giu"

// 全局主窗口实例
var w *mainWindow

// 主界面循环函数
func MainWindow(wnd *giu.MasterWindow) *mainWindow {
	if w == nil {
		w = newMainWindow()
	}

	wnd.SetDropCallback(func(s []string) {
		w.romList.handleExternalFileDrop(s)
	})

	return w
}
