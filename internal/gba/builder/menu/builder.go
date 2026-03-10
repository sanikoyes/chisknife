package menu

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"chisknife/internal/gba/builder/emulator"
	"chisknife/internal/gba/builder/rombuilder"
	"chisknife/internal/gba/patcher/batteryless"
	"chisknife/internal/gba/patcher/rts"
	"chisknife/internal/gba/patcher/sram"
	"chisknife/internal/lang"
	"chisknife/internal/preset"

	"github.com/alitto/pond"
)

const (
	romOutDir = "game_patched"
)

// IPS 补丁游戏列表（需要特殊 IPS 补丁的游戏 ID）
var ipsGameList = []string{
	"A2YE", "A3UJ", "ABFJ", "AGHJ", "AGIJ",
	"AK8E", "AK8P", "AK9E", "AK9P", "ALUE",
	"B24J", "B3EJ", "BKME", "BKMJ", "BKMP",
	"BU6J", "BUHJ",
}

// 模拟器游戏列表
var emuGameList = []string{"GMBC", "PNES"}

// 检查游戏 ID 是否在列表中
func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// 构建开始
func BuildStart(opts BuildOptions, roms []GameInput) <-chan BuildInfo {
	ch := make(chan BuildInfo)

	go func() {
		defer close(ch)

		// 创建输出目录
		if _, err := os.Stat(romOutDir); err == nil {
			os.RemoveAll(romOutDir)
		}
		os.MkdirAll(romOutDir, 0755)

		// 准备游戏 JSON 配置
		var gameListMu sync.Mutex
		gameList := []*rombuilder.GameConfig{}

		// 创建 pond 工作池，使用 CPU 核心数作为工作线程数
		pool := pond.New(10, len(roms))

		// 提交所有游戏处理任务
		for _, game := range roms {
			game := game // 捕获循环变量
			pool.Submit(func() {
				fileNameFull := filepath.Base(game.Path)
				fileName := strings.TrimSuffix(fileNameFull, filepath.Ext(fileNameFull))
				fileType := strings.ToLower(filepath.Ext(fileNameFull))
				outFile := filepath.Join(romOutDir, fileName+".gba")

				var err error
				switch fileType {
				case ".gba":
					err = processGBAGame(game, outFile, opts, ch)

				case ".gb", ".gbc":
					err = processGBGame(game, outFile, opts, ch)

				case ".nes":
					err = processNESGame(game, outFile, opts, ch)

				default:
					ch <- BuildInfo{
						Path:    fileNameFull,
						Type:    lang.L("type detect"),
						Message: lang.L("Not a valid type."),
						Success: false,
					}
					return
				}

				if err != nil {
					ch <- BuildInfo{
						Path:    fileNameFull,
						Type:    lang.L("raw copy"),
						Message: fmt.Sprintf(lang.L("process rom failed, err:%v"), err),
						Success: false,
					}
					copyFile(game.Path, outFile)
					err = nil
				}

				// 如果处理成功，添加到游戏列表
				if err == nil {
					gameListMu.Lock()
					gameList = append(gameList, &rombuilder.GameConfig{
						Enabled:   true,
						File:      fileName + ".gba",
						Title:     game.Name,
						TitleFont: 1,
						SaveSlot:  game.SaveSlot,
					})
					gameListMu.Unlock()
				}
			})
		}

		// 等待所有任务完成
		pool.StopAndWait()

		ch <- BuildInfo{
			Path:    "builder.json",
			Type:    lang.L("config generation"),
			Message: lang.L("Configuration file generated successfully."),
			Success: true,
		}

		// 按照原始顺序排序后再生成合卡rom
		slices.SortFunc(gameList, func(a, b *rombuilder.GameConfig) int {
			left := slices.IndexFunc(roms, func(rom GameInput) bool {
				return a.Title == rom.Name
			})

			right := slices.IndexFunc(roms, func(rom GameInput) bool {
				return b.Title == rom.Name
			})

			return cmp.Compare(left, right)
		})

		// 构建配置结构
		config := &rombuilder.Config{
			Cartridge: rombuilder.CartridgeConfig{
				Type:           opts.CartridgeType,
				BatteryPresent: opts.BatteryPresent,
				MinRomSize:     opts.MinRomSize,
			},
			Games: gameList,
		}

		// 调用 ROM builder 构建最终的多游戏菜单 ROM
		ch <- BuildInfo{
			Path:    "",
			Type:    lang.L("rom build"),
			Message: lang.L("Starting ROM build..."),
			Success: true,
		}

		buildOpts := rombuilder.BuildOptions{
			Split:       false,
			NoWait:      true,
			Config:      config, // 直接传入配置结构
			Bg:          opts.BgPath,
			Output:      opts.OutputPath,
			RomBasePath: romOutDir,
			CLIMode:     false,
			LogCallback: func(msg string) {
				// 将 rombuilder 的日志输出到 build progress 窗口
				ch <- BuildInfo{
					Path:    "",
					Type:    lang.L("rom build"),
					Message: msg,
					Success: true,
				}
			},
		}

		result, err := rombuilder.Build(buildOpts)
		if err != nil {
			ch <- BuildInfo{
				Path:    opts.OutputPath,
				Type:    lang.L("rom build"),
				Message: fmt.Sprintf(lang.L("ROM build failed: %v"), err),
				Success: false,
			}
			return
		}

		if !result.Success {
			ch <- BuildInfo{
				Path:    opts.OutputPath,
				Type:    lang.L("rom build"),
				Message: result.Message,
				Success: false,
			}
			return
		}

		// 构建成功
		outputFile := strings.Replace(opts.OutputPath, "<CODE>", result.ROMCode, -1)
		ch <- BuildInfo{
			Path: outputFile,
			Type: lang.L("rom build"),
			Message: fmt.Sprintf(lang.L("ROM build completed. Code: %s, Size: %.2f MB, Games: %d"),
				result.ROMCode, float64(result.ROMSize)/1024/1024, result.GamesAdded),
			Success: true,
		}

		// 报告未能添加的游戏
		if len(result.Data) > 0 {
			for _, game := range result.Data {
				ch <- BuildInfo{
					Path:    game.File,
					Type:    lang.L("rom build"),
					Message: fmt.Sprintf(lang.L("Warning: \"%s\" couldn't be added due to space constraints"), game.Title),
					Success: false,
				}
			}
		}
	}()

	return ch
}

// 处理 GBA 游戏
func processGBAGame(game GameInput, outFile string, opts BuildOptions, ch chan<- BuildInfo) error {
	fileNameFull := filepath.Base(game.Path)

	// 获取游戏 ID
	gameID, err := GetROMID(game.Path)
	if err != nil {
		ch <- BuildInfo{
			Path:    fileNameFull,
			Type:    lang.L("read ROM"),
			Message: fmt.Sprintf(lang.L("Failed to read ROM: %v"), err),
			Success: false,
		}
		return err
	}

	// 检查是否需要 IPS 补丁
	if contains(ipsGameList, gameID) {
		ipsPath := fmt.Sprintf("sram_ips/%s.ips", gameID)
		if err := sram.IPSPatch(game.Path, ipsPath, outFile); err != nil {
			ch <- BuildInfo{
				Path:    fileNameFull,
				Type:    lang.L("IPS patch"),
				Message: lang.L("IPS patch failed."),
				Success: false,
			}
			return err
		}
		ch <- BuildInfo{
			Path:    fileNameFull,
			Type:    lang.L("IPS patch"),
			Message: lang.L("IPS patch succeed."),
			Success: true,
		}
	} else if contains(emuGameList, gameID) {
		// 跳过模拟器
		copyFile(game.Path, outFile)
		ch <- BuildInfo{
			Path:    fileNameFull,
			Type:    lang.L("raw copy"),
			Message: lang.L("Skip emulator rom"),
			Success: false,
		}
		copyFile(game.Path, outFile)

	} else {
		// 检查存档类型
		saveType := CheckSaveType(game.Path)
		if saveType == SaveTypeNone || saveType == SaveTypeSRAM {
			ch <- BuildInfo{
				Path:    fileNameFull,
				Type:    lang.L("raw copy"),
				Message: fmt.Sprintf(lang.L("Skip Save Type:%s"), saveType),
				Success: false,
			}
			copyFile(game.Path, outFile)
		} else {
			// 应用 SRAM 补丁
			if err := sram.SRAMPatchBank(game.Path, outFile, opts.SRAMBankType); err != nil {
				ch <- BuildInfo{
					Path:    fileNameFull,
					Type:    lang.L("SRAM patch"),
					Message: lang.L("SRAM patch failed."),
					Success: false,
				}
				return err
			}
			ch <- BuildInfo{
				Path:    fileNameFull,
				Type:    lang.L("SRAM patch"),
				Message: lang.L("SRAM patch succeed."),
				Success: true,
			}
		}
	}

	// 应用 batteryless 或 RTS 补丁
	if !opts.BatteryPresent && game.SaveSlot != nil && !contains(emuGameList, gameID) {
		patcher := batteryless.NewROMPatcher()
		if err := patcher.Patch(outFile, outFile, opts.BatterylessAutoSave); err != nil {
			ch <- BuildInfo{
				Path:    fileNameFull,
				Type:    lang.L("batteryless patch"),
				Message: lang.L("Batteryless patch failed."),
				Success: false,
			}
			return err
		}
		ch <- BuildInfo{
			Path:    fileNameFull,
			Type:    lang.L("batteryless patch"),
			Message: lang.L("Batteryless patch succeed."),
			Success: true,
		}
	} else if opts.UseRTS {
		if opts.CartridgeType > 0 && opts.CartridgeType <= len(preset.CartridgeTypes) {
			sectorSize := uint32(preset.CartridgeTypes[opts.CartridgeType-1].SectorSize)
			patcher := rts.NewRTSPatcher()
			romData, _ := os.ReadFile(outFile)
			patchedData, err := patcher.PatchBytes(romData, 0, sectorSize, nil)
			if err != nil {
				ch <- BuildInfo{
					Path:    fileNameFull,
					Type:    lang.L("RTS patch"),
					Message: lang.L("RTS patch failed."),
					Success: false,
				}
				return err
			}
			os.WriteFile(outFile, patchedData, 0644)
		}
	}

	return nil
}

// 处理 GB/GBC 游戏
func processGBGame(game GameInput, outFile string, opts BuildOptions, ch chan<- BuildInfo) error {
	fileNameFull := filepath.Base(game.Path)

	goombaPath := "emulator/jagoombacolor.gba"
	if !opts.BatteryPresent && game.SaveSlot != nil {
		goombaPath = "emulator/jagoombacolor_batteryless.gba"
	}

	if err := emulator.BuildGoomba([]string{game.Path}, outFile, goombaPath); err != nil {
		ch <- BuildInfo{
			Path:    fileNameFull,
			Type:    lang.L("goomba build"),
			Message: fmt.Sprintf(lang.L("Goomba build failed: %v"), err),
			Success: false,
		}
		return err
	}

	ch <- BuildInfo{
		Path:    fileNameFull,
		Type:    lang.L("goomba build"),
		Message: lang.L("Goomba build succeed."),
		Success: true,
	}
	return nil
}

// 处理 NES 游戏
func processNESGame(game GameInput, outFile string, opts BuildOptions, ch chan<- BuildInfo) error {
	fileNameFull := filepath.Base(game.Path)

	pocketnesPath := "emulator/pocketnes.gba"
	if !opts.BatteryPresent && game.SaveSlot != nil {
		pocketnesPath = "emulator/pocketnes_batteryless.gba"
	}

	if err := emulator.BuildPocketNES([]string{game.Path}, outFile, pocketnesPath, "emulator/pnesmmw.mdb"); err != nil {
		ch <- BuildInfo{
			Path:    fileNameFull,
			Type:    lang.L("pocketnes build"),
			Message: fmt.Sprintf(lang.L("PocketNES build failed: %v"), err),
			Success: false,
		}
		return err
	}

	ch <- BuildInfo{
		Path:    fileNameFull,
		Type:    lang.L("pocketnes build"),
		Message: lang.L("PocketNES build succeed."),
		Success: true,
	}
	return nil
}

// 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
