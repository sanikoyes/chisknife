package ui

import (
	"fmt"
	"path/filepath"

	"chisknife/internal/gba/builder/menu"
	"chisknife/internal/lang"

	"github.com/AllenDang/giu"
)

// 打包进度窗口
type buildProgress struct {
	isOpen        bool        // 窗口是否打开
	currentFile   string      // 当前处理的文件
	currentIdx    int         // 当前文件索引
	totalFiles    int         // 总文件数
	progress      float32     // 进度 (0.0 - 1.0)
	logs          []string    // 日志信息
	isBuilding    bool        // 是否正在构建
	buildComplete bool        // 构建是否完成
	hasError      bool        // 是否有错误
	mainWindow    *mainWindow // 主窗口引用
	outputPath    string      // 输出文件路径
}

// 创建打包进度窗口
func newBuildProgress() *buildProgress {
	return &buildProgress{
		isOpen:        false,
		logs:          []string{},
		isBuilding:    false,
		buildComplete: false,
		hasError:      false,
	}
}

// 打开进度窗口
func (bp *buildProgress) open() {
	bp.isOpen = true
	bp.currentFile = ""
	bp.currentIdx = 0
	bp.totalFiles = 0
	bp.progress = 0.0
	bp.logs = []string{}
	bp.isBuilding = false
	bp.buildComplete = false
	bp.hasError = false
}

// 关闭进度窗口
func (bp *buildProgress) close() {
	bp.isOpen = false
}

// 添加日志
func (bp *buildProgress) addLog(message string) {
	bp.logs = append(bp.logs, message)
	// 限制日志数量
	if len(bp.logs) > 1000 {
		bp.logs = bp.logs[1:]
	}
}

// 更新进度
func (bp *buildProgress) updateProgress(idx, total int, fileName string) {
	bp.currentIdx = idx
	bp.totalFiles = total
	bp.currentFile = fileName
	if total > 0 {
		bp.progress = float32(idx) / float32(total)
	}
}

// 开始构建
func (bp *buildProgress) startBuild(opts menu.BuildOptions, games []menu.GameInput) {
	bp.isBuilding = true
	bp.buildComplete = false
	bp.hasError = false
	bp.totalFiles = len(games)
	bp.currentIdx = 0
	bp.outputPath = opts.OutputPath

	// 在新的 goroutine 中执行构建
	go func() {
		processedGames := make(map[string]bool)
		totalSteps := len(games) + 2 // 游戏数量 + 配置生成 + ROM 构建
		completedSteps := 0

		for info := range menu.BuildStart(opts, games) {
			// 添加日志
			status := "✓"
			if !info.Success {
				status = "✗"
				bp.hasError = true
			}

			// 格式化日志消息
			var logMsg string
			if info.Path != "" {
				logMsg = fmt.Sprintf("%s [%s] %s: %s", status, info.Type, filepath.Base(info.Path), info.Message)
			} else {
				logMsg = fmt.Sprintf("%s [%s] %s", status, info.Type, info.Message)
			}
			bp.addLog(logMsg)

			// 更新进度：只在游戏首次处理完成、配置生成、ROM 构建时增加
			if info.Type == "config generation" {
				completedSteps++
			} else if info.Type == "rom build" {
				completedSteps++
			} else if info.Path != "" {
				// 游戏处理消息，只在首次遇到该游戏时计数
				if !processedGames[info.Path] {
					processedGames[info.Path] = true
					completedSteps++
				}
			}

			bp.updateProgress(completedSteps, totalSteps, filepath.Base(info.Path))
		}

		// 构建完成，确保进度为 100%
		bp.updateProgress(totalSteps, totalSteps, "")
		bp.isBuilding = false
		bp.buildComplete = true
		if bp.hasError {
			bp.addLog("⚠ Build completed with warnings/errors")
		} else {
			bp.addLog("✓ Build completed successfully!")
			// 构建成功，保存输出路径到项目配置
			if bp.mainWindow != nil && bp.outputPath != "" {
				bp.mainWindow.project.LastBuildOutput = bp.outputPath
				// 自动保存项目
				if bp.mainWindow.currentProjectPath != "" {
					bp.mainWindow.doSaveProject(bp.mainWindow.currentProjectPath)
				}
			}
		}
	}()
}

// 构建窗口界面
func (bp *buildProgress) build() giu.Widget {
	if !bp.isOpen {
		return nil
	}

	// 将日志合并为单个字符串
	logText := ""
	for _, log := range bp.logs {
		logText += log + "\n"
	}

	// 进度文本
	progressText := ""
	if bp.isBuilding {
		progressText = fmt.Sprintf("%s (%d/%d)", bp.currentFile, bp.currentIdx, bp.totalFiles)
	} else if bp.buildComplete {
		if bp.hasError {
			progressText = lang.L("Build completed with errors")
		} else {
			progressText = lang.L("Build completed successfully")
		}
	} else {
		progressText = lang.L("Ready to build")
	}

	return giu.Column(
		giu.Label(lang.L("Build Progress")),
		giu.Separator(),

		// 当前文件和进度
		giu.Row(
			giu.Label(progressText),
		),

		// 进度条
		giu.ProgressBar(bp.progress).Size(giu.Auto, 0).
			Overlay(fmt.Sprintf("%.0f%%", bp.progress*100)),

		giu.Separator(),

		// 日志区域（只读但可选择复制）
		giu.Label(lang.L("Build Log:")),
		giu.InputTextMultiline(&logText).
			Size(giu.Auto, giu.Auto-32).
			Flags(giu.InputTextFlagsReadOnly),

		giu.Separator(),

		// 按钮
		giu.Row(
			giu.Button(lang.L("Close")).
				Disabled(bp.isBuilding).
				OnClick(func() {
					bp.close()
				}),
		),
	)
}
