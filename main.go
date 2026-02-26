// Package main 是 Chis Army Knife 应用程序的入口点
// 这是一个用于管理和构建游戏 ROM 卡带的图形界面工具
package main

import (
	"chisknife/internal/lang"
	"chisknife/internal/ui"

	"github.com/AllenDang/giu"
)

// 应用程序的主入口函数
// 创建并启动主窗口，初始化 UI 循环
func main() {
	wnd := giu.NewMasterWindow(lang.L("Chis Army Knife"), 640, 480, 0)
	wnd.Run(ui.Loop)
}
