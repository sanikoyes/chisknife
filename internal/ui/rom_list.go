// Package ui 提供应用程序的图形用户界面
package ui

import (
	"chisknife/internal/gba/builder/menu"
	"chisknife/internal/lang"
	"chisknife/internal/types"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/giu"
	"github.com/ncruces/zenity"
)

type romList struct {
	project          *types.Project // 打包工程数据
	mainWindow       *mainWindow    // 主窗口引用
	selectedRomIndex int            // 记录当前选中的 ROM 项索引
	editingIndex     int            // 正在编辑的 ROM 项索引
	editingName      string         // 编辑中的名称
}

func newRomList(project *types.Project) *romList {
	return &romList{
		project:          project,
		mainWindow:       nil, // 稍后设置
		selectedRomIndex: -1,
		editingIndex:     -1,
	}
}

// 刷新界面状态
func (ui *romList) refresh() {
	// 重置选中状态
	ui.selectedRomIndex = -1
	ui.editingIndex = -1
	ui.editingName = ""
}

// 构建 ROM 列表界面组件
// 显示所有 ROM 文件，支持选择和拖拽排序
func (ui *romList) build() giu.Widget {
	roms := ui.project.Roms

	return giu.Column(
		giu.Label(lang.L("ROM List")),
		giu.Separator(),
		giu.Child().Size(0, giu.Auto-56).Layout(
			giu.Custom(func() {
				if len(roms) == 0 {
					giu.Label(lang.L("No ROMs")).Build()
					return
				}

				var edited bool

				for i := range roms {
					if ui.buildDraggableRomItem(i) {
						edited = true
					}
				}

				// 检测选中项按回车键进入编辑模式
				if !edited && ui.selectedRomIndex >= 0 && ui.selectedRomIndex < len(roms) && ui.editingIndex == -1 {
					if imgui.IsKeyPressedBool(imgui.KeyEnter) || imgui.IsKeyPressedBool(imgui.KeyKeypadEnter) {
						ui.editingIndex = ui.selectedRomIndex
						ui.editingName = roms[ui.selectedRomIndex].Name
					}
				}
			}),
		),
		giu.Separator(),
		giu.Align(giu.AlignCenter).To(
			giu.Row(
				giu.Button(lang.L("Add")).OnClick(func() {
					ui.addRom()
				}),
				giu.Button(lang.L("Remove")).OnClick(func() {
					ui.removeRom()
				}).Disabled(ui.selectedRomIndex == -1),
				giu.Button(lang.L("Clear")).OnClick(func() {
					ui.clearRoms()
				}).Disabled(len(roms) == 0),
				giu.Button(lang.L("Sort")).OnClick(func() {
					ui.sortRoms()
				}).Disabled(len(roms) < 2),
			),
		),
		giu.Align(giu.AlignCenter).To(
			giu.Row(
				giu.Button(lang.L("Build ROM")).OnClick(func() {
					ui.buildRom()
				}).Disabled(len(roms) == 0),
			),
		),
	)
}

// 构建单个可拖拽的 ROM 列表项
// 支持点击选择和拖拽重新排序
func (ui *romList) buildDraggableRomItem(index int) bool {
	roms := ui.project.Roms
	defer func() {
		ui.project.Roms = roms
	}()

	// 只显示文件名，不显示完整路径
	rom := roms[index]
	displayName := rom.Name
	payloadType := "ROM_ITEM"

	// 如果正在编辑此项，显示输入框
	if ui.editingIndex == index {
		giu.SetKeyboardFocusHere()

		giu.InputText(&ui.editingName).
			Flags(giu.InputTextFlagsEnterReturnsTrue).
			Size(giu.Auto).
			Build()

		var shouldExited bool
		var shouldUpdate bool

		// 检测回车键确认（InputTextFlagsEnterReturnsTrue 会让输入框在按回车时失去焦点）
		if imgui.IsItemDeactivatedAfterEdit() {
			shouldExited = true
			shouldUpdate = true
		}

		// 检测点击其他地方退出编辑
		if imgui.IsMouseClickedBool(imgui.MouseButtonLeft) && !imgui.IsItemHovered() {
			shouldExited = true
			shouldUpdate = true
		}

		// 在 Custom 外部检测 ESC 键
		if imgui.IsKeyPressedBool(imgui.KeyEscape) {
			shouldExited = true
			shouldUpdate = false
		}

		// 检查是否退出编辑
		if shouldExited {
			// 需要更新编辑内容？
			if shouldUpdate && ui.editingName != "" {
				roms[index].Name = ui.editingName
			}
			ui.editingIndex = -1
		}

		return shouldExited || shouldUpdate
	}

	// 创建可选择和可拖拽的列表项
	giu.Row(
		giu.Selectable(displayName).
			Selected(index == ui.selectedRomIndex).
			OnClick(func() {
				ui.selectedRomIndex = index
			}).
			OnDClick(func() {
				// 双击进入编辑模式
				ui.editingIndex = index
				ui.editingName = rom.Name
			}).
			Flags(giu.SelectableFlagsAllowDoubleClick),
	).Build()

	// 设置为拖拽源，允许拖动此项
	if imgui.BeginDragDropSource() {
		// 设置拖拽数据，传递当前项的索引
		indexCopy := index // 创建副本避免闭包问题
		indexPtr := uintptr(unsafe.Pointer(&indexCopy))
		imgui.SetDragDropPayload(payloadType, indexPtr, uint64(unsafe.Sizeof(indexCopy)))
		imgui.Text(displayName)
		imgui.EndDragDropSource()
	}

	// 设置为拖拽目标，允许接收拖放
	if imgui.BeginDragDropTarget() {
		payload := imgui.AcceptDragDropPayload(payloadType)
		if payload != nil && payload.CData != nil && payload.Data() != 0 && uint64(payload.DataSize()) >= uint64(unsafe.Sizeof(int(0))) {
			// 安全地读取拖拽数据
			dataPtr := payload.Data()
			sourceIndexPtr := (*int)(unsafe.Pointer(dataPtr))
			if sourceIndexPtr != nil {
				sourceIndex := *sourceIndexPtr
				if sourceIndex != index && sourceIndex >= 0 && sourceIndex < len(roms) {
					// 交换
					roms[sourceIndex], roms[index] = roms[index], roms[sourceIndex]

					// 更新选中索引以跟随移动的项
					switch ui.selectedRomIndex {
					case sourceIndex:
						ui.selectedRomIndex = index
					case index:
						ui.selectedRomIndex = sourceIndex
					}
				}
			}
		}
		imgui.EndDragDropTarget()
	}

	return false
}

// 添加rom文件路径
func (ui *romList) appendRom(path string) {
	roms := ui.project.Roms

	// 不重复添加rom
	for _, rom := range roms {
		if rom.Path == path {
			return
		}
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".gba", ".nes", ".gb", ".gbc":
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		roms = append(roms, types.Rom{
			Name: name,
			Path: path,
		})
		ui.project.Roms = roms
	}
}

// 添加 ROM 文件
// 打开文件选择对话框，支持 gba/nes/gb/gbc 格式
func (ui *romList) addRom() {
	// 使用 zenity 打开文件选择对话框
	files, err := zenity.SelectFileMultiple(
		zenity.Title(lang.L("Select ROM files")),
		zenity.FileFilters{
			{
				Name:     "GBA ROM files",
				Patterns: []string{"*.gba"},
				CaseFold: true,
			},
			{
				Name:     "NES ROM files",
				Patterns: []string{"*.nes"},
				CaseFold: true,
			},
			{
				Name:     "GB/GBC ROM files",
				Patterns: []string{"*.gb", "*.gbc"},
				CaseFold: true,
			},
			{
				Name:     "All files",
				Patterns: []string{"*"},
			},
		},
	)

	if err == nil && len(files) > 0 {
		for _, file := range files {
			ui.appendRom(file)
		}
	}
}

// 移除选中的 ROM 文件
func (ui *romList) removeRom() {
	roms := ui.project.Roms

	if ui.selectedRomIndex >= 0 && ui.selectedRomIndex < len(roms) {
		// 删除选中的项
		roms = slices.Delete(roms, ui.selectedRomIndex, ui.selectedRomIndex+1)
		ui.project.Roms = roms

		// 调整选中索引
		ui.selectedRomIndex = -1
	}
}

// 清空所有 ROM 文件
func (ui *romList) clearRoms() {
	ui.project.Roms = types.RomList{}
	ui.selectedRomIndex = -1
}

// 按文件名排序 ROM 列表
// 如果当前已经是正序，则切换为倒序
func (ui *romList) sortRoms() {
	roms := ui.project.Roms

	if len(roms) <= 1 {
		return
	}

	// 创建一个副本用于比较
	original := make(types.RomList, len(roms))
	copy(original, roms)

	// 按文件名排序
	sort.Slice(roms, func(i, j int) bool {
		nameI := strings.ToLower((roms)[i].Name)
		nameJ := strings.ToLower((roms)[j].Name)
		return nameI < nameJ
	})

	// 检查排序后是否发生变化
	isSame := true
	for i := range roms {
		if roms[i].Path != original[i].Path {
			isSame = false
			break
		}
	}

	// 如果排序后没有变化，或者当前已经是正序状态，则进行倒序
	if isSame {
		// 倒序
		for i, j := 0, len(roms)-1; i < j; i, j = i+1, j-1 {
			roms[i], roms[j] = roms[j], roms[i]
		}
	}

	ui.project.Roms = roms

	// 重置选中索引
	ui.selectedRomIndex = -1
}

// 生成rom文件
func (ui *romList) buildRom() {
	if ui.mainWindow == nil || ui.mainWindow.buildProgress == nil {
		return
	}

	// 使用上次构建的路径作为默认值
	executableDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		executableDir = "."
	}
	defaultPath := executableDir
	if ui.project.LastBuildOutput != "" {
		defaultPath = ui.project.LastBuildOutput
	}

	// 弹出文件保存对话框
	outputPath, err := zenity.SelectFileSave(
		zenity.Title(lang.L("Save ROM As")),
		zenity.Filename(defaultPath),
		zenity.FileFilters{
			{
				Name:     "GBA ROM files",
				Patterns: []string{"*.gba"},
				CaseFold: true,
			},
		},
	)

	// 用户取消或出错
	if err != nil || outputPath == "" {
		return
	}

	// 确保文件扩展名为 .gba
	if !strings.HasSuffix(strings.ToLower(outputPath), ".gba") {
		outputPath += ".gba"
	}

	// 打开进度窗口
	ui.mainWindow.buildProgress.open()

	// 准备构建选项
	opts := ui.prepareBuildOptions()
	opts.OutputPath = outputPath // 使用用户选择的路径

	// 准备游戏列表
	games := ui.prepareGameList()

	// 开始构建
	ui.mainWindow.buildProgress.startBuild(opts, games)
}

// 准备构建选项
func (ui *romList) prepareBuildOptions() menu.BuildOptions {
	project := ui.project
	cartType := int(project.Options.CartridgeType) + 1 // 转换为 1-based

	return menu.BuildOptions{
		CartridgeType:       cartType,
		BatteryPresent:      project.Options.HaveBattery,
		MinRomSize:          int(project.Options.MinimalRomSize),
		SRAMBankType:        0, // 默认值
		BatterylessAutoSave: false,
		UseRTS:              project.Options.UseRTS,
		ConfigPath:          "builder.json",
		RomBasePath:         "game_patched",
		OutputPath:          "multimenu.gba",
		BgPath:              "",
		Split:               project.Options.SplitROM,
	}
}

// 准备游戏列表
func (ui *romList) prepareGameList() []menu.GameInput {
	games := make([]menu.GameInput, 0, len(ui.project.Roms))

	for i, rom := range ui.project.Roms {
		saveSlot := i + 1 // 1-based 索引
		game := menu.GameInput{
			Path:     rom.Path,
			Name:     rom.Name,
			SaveSlot: &saveSlot,
		}
		games = append(games, game)
	}

	return games
}

// 处理从外部拖拽文件到列表区域
// 支持从文件管理器拖拽 gba/nes/gb/gbc 文件
func (ui *romList) handleExternalFileDrop(filePaths []string) {
	for _, file := range filePaths {
		file = strings.TrimSpace(file)
		ui.appendRom(file)
	}
}
