// Package ui 提供应用程序的图形用户界面
// 包括主窗口、ROM 列表和卡带设置等组件
package ui

import (
	"chisknife/internal/config"
	"chisknife/internal/lang"
	"chisknife/internal/types"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/AllenDang/giu"
	"github.com/ncruces/zenity"
)

// 应用程序的主窗口结构
type mainWindow struct {
	project            *types.Project // 打包工程数据
	sashPos            float32        // 分割面板的位置
	cartSettings       *cartSettings  // 卡带设置组件
	romList            *romList       // rom列表组件
	recentProjects     []string       // 最近打开的项目列表
	currentProjectPath string         // 当前项目文件路径
}

// 创建并初始化主窗口实例
func newMainWindow() *mainWindow {
	// 设置默认的构建选项
	opts := types.Options{
		CartridgeType:     0,
		MinimalRomSize:    0,
		HaveBattery:       true,
		UseRTS:            false,
		SplitROM:          false,
		Sram1MSaveSupport: false,
	}

	list := types.RomList{}

	project := &types.Project{
		Options: opts,
		Roms:    list,
	}

	mw := &mainWindow{
		sashPos:            320,
		project:            project,
		cartSettings:       newCartSettings(project),
		romList:            newRomList(project),
		recentProjects:     []string{},
		currentProjectPath: "",
	}

	// 自动加载最后一个项目
	lastProject := config.GetLastProject()
	if lastProject != "" {
		mw.doLoadProject(lastProject)
	}

	return mw
}

// 保存项目到 JSON 文件
func (ui *mainWindow) saveProject() {
	// 如果已有路径，直接保存
	if ui.currentProjectPath != "" {
		ui.doSaveProject(ui.currentProjectPath)
	} else {
		// 否则另存为
		ui.saveProjectAs()
	}
}

// 另存为项目
func (ui *mainWindow) saveProjectAs() {
	// 打开文件保存对话框
	filePath, err := zenity.SelectFileSave(
		zenity.Title(lang.L("Save Project As")),
		zenity.FileFilters{
			{
				Name:     "JSON files",
				Patterns: []string{"*.json"},
				CaseFold: true,
			},
		},
	)
	if err == nil && filePath != "" {
		ui.doSaveProject(filePath)
	}
}

// 加载项目从 JSON 文件
func (ui *mainWindow) loadProject() {
	// 打开文件选择对话框
	filePath, err := zenity.SelectFile(
		zenity.Title(lang.L("Load Project")),
		zenity.FileFilters{
			{
				Name:     "JSON files",
				Patterns: []string{"*.json"},
				CaseFold: true,
			},
		},
	)
	if err == nil && filePath != "" {
		ui.doLoadProject(filePath)
	}
}

// 执行保存操作
func (ui *mainWindow) doSaveProject(filePath string) error {
	data, err := json.MarshalIndent(ui.project, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	// 更新当前项目路径
	ui.currentProjectPath = filePath

	// 添加到最近项目列表
	config.AddRecentProject(filePath)
	ui.updateRecentProjects()

	return nil
}

// 执行加载操作
func (ui *mainWindow) doLoadProject(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var project types.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return err
	}

	ui.project.Options = project.Options
	ui.project.Roms = project.Roms
	ui.cartSettings.refresh()
	ui.romList.refresh()

	// 更新当前项目路径
	ui.currentProjectPath = filePath

	// 添加到最近项目列表
	config.AddRecentProject(filePath)
	ui.updateRecentProjects()

	return nil
}

// 更新最近项目列表
func (ui *mainWindow) updateRecentProjects() {
	ui.recentProjects = config.GetRecentProjects()
}

// 构建最近项目菜单
func (ui *mainWindow) buildRecentMenu() giu.Widget {
	if len(ui.recentProjects) == 0 {
		return giu.MenuItem(lang.L("No recent projects")).Enabled(false)
	}

	items := make([]giu.Widget, 0, len(ui.recentProjects))
	for _, path := range ui.recentProjects {
		// 创建局部变量避免闭包问题
		projectPath := path
		fileName := filepath.Base(projectPath)
		items = append(items, giu.MenuItem(fileName).OnClick(func() {
			ui.doLoadProject(projectPath)
		}))
	}

	return giu.Layout(items)
}

// 主窗口运行逻辑
func (ui *mainWindow) Loop() {
	// 更新最近项目列表
	ui.updateRecentProjects()

	giu.SingleWindow().
		Flags(giu.WindowFlagsNoTitleBar+giu.WindowFlagsMenuBar).
		RegisterKeyboardShortcuts().
		Layout(
			giu.MenuBar().Layout(
				giu.Menu(lang.L("File")).Layout(
					giu.MenuItem(lang.L("Save Project")).OnClick(ui.saveProject),
					giu.MenuItem(lang.L("Save Project As")).OnClick(ui.saveProjectAs),
					giu.MenuItem(lang.L("Load Project")).OnClick(ui.loadProject),
					giu.Separator(),
					giu.Menu(lang.L("Recent")).Layout(
						ui.buildRecentMenu(),
					),
				),
				giu.Menu(lang.L("ROM")).Layout(
					giu.MenuItem(lang.L("Add")).OnClick(func() {
						ui.romList.addRom()
					}),
					giu.MenuItem(lang.L("Remove")).OnClick(func() {
						ui.romList.removeRom()
					}).Enabled(ui.romList.selectedRomIndex != -1),
					giu.MenuItem(lang.L("Clear")).OnClick(func() {
						ui.romList.clearRoms()
					}).Enabled(len(ui.project.Roms) > 0),
					giu.MenuItem(lang.L("Sort")).OnClick(func() {
						ui.romList.sortRoms()
					}).Enabled(len(ui.project.Roms) > 1),
				),
			),
			giu.SplitLayout(
				giu.DirectionVertical,
				&ui.sashPos,
				ui.romList.build(),
				ui.cartSettings.build(),
			),
		)
}
