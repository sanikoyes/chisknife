// Package ui 提供应用程序的图形用户界面
package ui

import (
	"chisknife/internal/lang"
	"chisknife/internal/types"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/giu"
	"github.com/ncruces/zenity"
)

type romList struct {
	list             *types.RomList
	selectedRomIndex int // 记录当前选中的 ROM 项索引
}

func newRomList(list *types.RomList) *romList {
	return &romList{
		list:             list,
		selectedRomIndex: -1,
	}
}

// 构建 ROM 列表界面组件
// 显示所有 ROM 文件，支持选择和拖拽排序
func (ui *romList) build() giu.Widget {
	roms := ui.list.Roms

	return giu.Column(
		giu.Label(lang.L("ROM List")),
		giu.Separator(),
		giu.Child().Size(0, giu.Auto-30).Layout(
			giu.Custom(func() {
				if len(roms) == 0 {
					giu.Label(lang.L("No ROMs")).Build()
					return
				}

				for i := range roms {
					ui.buildDraggableRomItem(&roms, i)
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
				}),
				giu.Button(lang.L("Clear")).OnClick(func() {
					ui.clearRoms()
				}),
			),
		),
	)
}

// 构建单个可拖拽的 ROM 列表项
// 支持点击选择和拖拽重新排序
func (ui *romList) buildDraggableRomItem(roms *[]string, index int) {
	rom := (*roms)[index]
	// 只显示文件名，不显示完整路径
	displayName := filepath.Base(rom)
	payloadType := "ROM_ITEM"

	// 创建可选择和可拖拽的列表项
	giu.Row(
		giu.Selectable(displayName).
			Selected(index == ui.selectedRomIndex).
			OnClick(func() {
				ui.selectedRomIndex = index
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
				if sourceIndex != index && sourceIndex >= 0 && sourceIndex < len(*roms) {
					// 交换
					tmp := (*roms)[sourceIndex]
					(*roms)[sourceIndex] = (*roms)[index]
					(*roms)[index] = tmp

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
			if file != "" {
				// 验证文件扩展名
				ext := strings.ToLower(filepath.Ext(file))
				if ext == ".gba" || ext == ".nes" || ext == ".gb" || ext == ".gbc" {
					ui.list.Roms = append(ui.list.Roms, file)
				}
			}
		}
	}
}

// 移除选中的 ROM 文件
func (ui *romList) removeRom() {
	if ui.selectedRomIndex >= 0 && ui.selectedRomIndex < len(ui.list.Roms) {
		// 删除选中的项
		ui.list.Roms = append(
			ui.list.Roms[:ui.selectedRomIndex],
			ui.list.Roms[ui.selectedRomIndex+1:]...,
		)

		// 调整选中索引
		if ui.selectedRomIndex >= len(ui.list.Roms) {
			ui.selectedRomIndex = len(ui.list.Roms) - 1
		}
	}
}

// 清空所有 ROM 文件
func (ui *romList) clearRoms() {
	ui.list.Roms = []string{}
	ui.selectedRomIndex = -1
}
