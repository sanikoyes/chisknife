// Package ui 提供应用程序的图形用户界面
// 包括主窗口、ROM 列表和卡带设置等组件
package ui

import (
	"chisknife/internal/types"

	"github.com/AllenDang/giu"
)

// 应用程序的主窗口结构
type mainWindow struct {
	opts         *types.Options // 构建选项配置
	list         *types.RomList // ROM 文件列表
	sashPos      float32        // 分割面板的位置
	cartSettings *cartSettings  // 卡带设置组件
	romList      *romList       // rom列表组件
}

// 创建并初始化主窗口实例
func newMainWindow() *mainWindow {
	// 设置默认的构建选项
	opts := &types.Options{
		CartridgeType:     0,
		MinimalRomSize:    0,
		HaveBattery:       true,
		UseRTS:            false,
		SplitROM:          false,
		Sram1MSaveSupport: false,
	}

	list := &types.RomList{
		Roms: []string{},
	}

	return &mainWindow{
		sashPos:      320,
		opts:         opts,
		list:         list,
		cartSettings: newCartSettings(opts),
		romList:      newRomList(list),
	}
}

// 构建主窗口的界面布局
func (ui *mainWindow) build() {
	giu.SingleWindow().RegisterKeyboardShortcuts().Layout(
		giu.SplitLayout(
			giu.DirectionVertical,
			&ui.sashPos,
			ui.romList.build(),
			ui.cartSettings.build(),
		),
	)
}

// 全局主窗口实例
var w *mainWindow

// 主界面循环函数
func Loop() {
	if w == nil {
		w = newMainWindow()
	}
	// 在每一帧被调用以更新和渲染界面
	w.build()
}
