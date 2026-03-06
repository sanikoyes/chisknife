// Package rombuilder 提供 GBA 多游戏菜单 ROM 构建功能
// 将多个 GBA ROM 文件合并成一个多游戏菜单 ROM
package rombuilder

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"chisknife/internal/preset"
)

const (
	AppVersion  = "1.2"
	MenuROMName = "lk_multimenu.gba"
)

// 构建选项
type BuildOptions struct {
	Split       bool             // 是否分割输出文件为 32MB 部分
	NoWait      bool             // 完成后不等待用户输入
	Config      *Config          // 配置结构（直接传入）
	Bg          string           // 背景图片路径
	Output      string           // 输出文件名
	RomBasePath string           // ROM 文件基础路径
	CLIMode     bool             // 是否为 CLI 模式
	LogCallback func(msg string) // 日志回调函数
}

// 游戏配置
type GameConfig struct {
	Enabled   bool     `json:"enabled"`            // 是否启用此游戏
	File      string   `json:"file"`               // ROM 文件名
	Title     string   `json:"title"`              // 游戏标题
	TitleFont int      `json:"title_font"`         // 标题字体类型
	SaveSlot  *int     `json:"save_slot"`          // 存档槽位（可选）
	Keys      []string `json:"keys,omitempty"`     // 隐藏 ROM 按键组合
	Map256M   bool     `json:"map_256m,omitempty"` // 是否使用 256M 映射

	// 内部使用字段
	Index        int    `json:"-"` // 游戏在列表中的索引
	Size         int    `json:"-"` // ROM 文件大小
	SectorCount  int    `json:"-"` // 占用的扇区数量
	SectorOffset int    `json:"-"` // 在编译缓冲区中的扇区偏移
	BlockOffset  int    `json:"-"` // 在编译缓冲区中的块偏移
	BlockCount   int    `json:"-"` // 占用的块数量
	SaveType     int    `json:"-"` // 存档类型
	KeysBitmap   uint16 `json:"-"` // 按键组合的位图表示
	Missing      bool   `json:"-"` // 文件是否缺失
}

// 卡带配置
type CartridgeConfig struct {
	Type           int  `json:"type"`
	BatteryPresent bool `json:"battery_present"`
	MinRomSize     int  `json:"min_rom_size"`
}

// 完整配置
type Config struct {
	Cartridge CartridgeConfig `json:"cartridge"`
	Games     []*GameConfig   `json:"games"`
}

// 构建结果
type BuildResult struct {
	Message     string
	Data        []*GameConfig // 未能添加的游戏列表
	Success     bool
	ROMCode     string
	ROMSize     int64
	GamesAdded  int
	SectorsUsed int
	SectorCount int
}

var logBuffer strings.Builder
var currentLogCallback func(string)
var currentLineBuffer strings.Builder

// 返回默认构建选项
func DefaultOptions() BuildOptions {
	return BuildOptions{
		Split:       false,
		NoWait:      false,
		Config:      nil,
		Bg:          "bg.png",
		Output:      "LK_MULTIMENU_<CODE>.gba",
		RomBasePath: "roms",
		CLIMode:     true,
		LogCallback: nil,
	}
}

// 打印并记录日志
func logp(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// 如果有回调函数，按行处理
	if currentLogCallback != nil {
		for _, ch := range msg {
			if ch == '\n' {
				// 遇到换行符，发送当前行（如果有内容）
				line := currentLineBuffer.String()
				if line != "" {
					currentLogCallback(line)
					currentLineBuffer.Reset()
				}
			} else {
				// 累积字符
				currentLineBuffer.WriteRune(ch)
			}
		}
	} else {
		// 否则输出到标准输出
		fmt.Print(msg)
	}

	logBuffer.WriteString(msg)
}

// 格式化文件大小
func formatFileSize(size int64) string {
	if size == 1 {
		return "1 Byte"
	} else if size < 1024 {
		return fmt.Sprintf("%d Bytes", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
}

// 更新扇区映射
func updateSectorMap(sectorMap []rune, start, length int, c rune) {
	if start+length > len(sectorMap) {
		return
	}
	sectorMap[start] = []rune(strings.ToUpper(string(c)))[0]
	for i := 1; i < length; i++ {
		sectorMap[start+i] = c
	}
}

// 解析按键组合
func parseKeys(keys []string) uint16 {
	var bitmap uint16
	for _, key := range keys {
		switch strings.ToUpper(key) {
		case "A":
			bitmap |= 1 << 0
		case "B":
			bitmap |= 1 << 1
		case "SELECT":
			bitmap |= 1 << 2
		case "START":
			bitmap |= 1 << 3
		case "RIGHT":
			bitmap |= 1 << 4
		case "LEFT":
			bitmap |= 1 << 5
		case "UP":
			bitmap |= 1 << 6
		case "DOWN":
			bitmap |= 1 << 7
		case "R":
			bitmap |= 1 << 8
		case "L":
			bitmap |= 1 << 9
		}
	}
	return bitmap
}

// 执行 ROM 构建
func Build(opts BuildOptions) (*BuildResult, error) {
	logBuffer.Reset()

	// 设置日志回调
	currentLogCallback = opts.LogCallback

	logp("GBA Multi Game Menu ROM Builder v%s\nby Lesserkuma\n\n", AppVersion)

	// 检查输出文件名
	if opts.Output == MenuROMName {
		return &BuildResult{
			Message: "Error: The file must not be named \"lk_multimenu.gba\"",
			Success: false,
		}, fmt.Errorf("invalid output name")
	}

	// 检查菜单 ROM 是否存在
	if _, err := os.Stat(MenuROMName); os.IsNotExist(err) {
		return &BuildResult{
			Message: "Error: The Menu ROM is missing. Expected file name: \"lk_multimenu.gba\"",
			Success: false,
		}, fmt.Errorf("menu ROM not found")
	}

	// 读取或生成配置文件
	config, err := loadOrGenerateConfig(opts)
	if err != nil {
		return &BuildResult{
			Message: fmt.Sprintf("Error: %v", err),
			Success: false,
		}, err
	}

	// 如果是新生成的配置，返回
	if config == nil {
		return &BuildResult{
			Message: "A new configuration was created.",
			Success: true,
		}, nil
	}

	// 执行构建
	result, err := buildROM(opts, config)

	return result, err
}

// 加载或生成配置文件
func loadOrGenerateConfig(opts BuildOptions) (*Config, error) {
	// 直接使用传入的配置
	if opts.Config != nil {
		return opts.Config, nil
	}

	return nil, fmt.Errorf("no configuration provided")
}

// 构建 ROM
func buildROM(opts BuildOptions, config *Config) (*BuildResult, error) {
	// 获取卡带类型
	cartridgeType := config.Cartridge.Type - 1
	if cartridgeType < 0 || cartridgeType >= len(preset.CartridgeTypes) {
		return &BuildResult{
			Message: "Error: Invalid cartridge type",
			Success: false,
		}, fmt.Errorf("invalid cartridge type")
	}

	cart := preset.CartridgeTypes[cartridgeType]
	flashSize := cart.FlashSize
	sectorSize := cart.SectorSize
	blockSize := cart.BlockSize
	sectorCount := flashSize / sectorSize
	_ = 0x80000 / sectorSize // sectorsPerBlock (未使用但保留计算)

	// 初始化编译缓冲区
	compilation := make([]byte, flashSize)
	for i := range compilation {
		compilation[i] = 0xFF
	}

	// 初始化扇区映射
	sectorMap := make([]rune, sectorCount)
	for i := range sectorMap {
		sectorMap[i] = '.'
	}

	// 读取菜单 ROM
	menuROM, err := os.ReadFile(MenuROMName)
	if err != nil {
		return &BuildResult{
			Message: "Error: Failed to read menu ROM",
			Success: false,
		}, err
	}

	// 对齐菜单 ROM
	menuROM = alignMenuROM(menuROM)

	// 添加构建时间戳
	buildTimestampOffset := len(menuROM) - 0x20
	timestamp := time.Now().Format(time.RFC3339)
	copy(menuROM[buildTimestampOffset:], []byte(timestamp))

	// 处理背景图片
	if opts.Bg != "" || fileExists("bg.png") {
		if err := updateBackground(menuROM, opts.Bg); err != nil {
			logp("Warning: Failed to update background image: %v\n", err)
		}
	}

	// 复制菜单 ROM 到编译缓冲区
	copy(compilation, menuROM)
	_ = findMenuROMSize(menuROM) // menuROMSize (未使用但保留计算)
	updateSectorMap(sectorMap, 0, int(math.Ceil(float64(len(menuROM))/float64(sectorSize))), 'm')

	// 计算偏移量
	itemListOffset := len(menuROM)
	itemListOffset = ((itemListOffset + 0x40000 - 1) / 0x40000) * 0x40000
	itemListOffset = int(math.Ceil(float64(itemListOffset) / float64(sectorSize)))
	updateSectorMap(sectorMap, itemListOffset, 1, 'l')

	statusOffset := itemListOffset + 1
	updateSectorMap(sectorMap, statusOffset, 1, 'c')

	// 创建状态区域
	status := createStatus(config.Cartridge.BatteryPresent)
	copy(compilation[statusOffset*sectorSize:], status)

	saveDataSectorOffset := statusOffset + 1

	// 检查启动 logo
	bootLogoFound := checkBootLogo(compilation[0x04:0xA0])

	// 处理游戏列表
	result := processGames(opts, config, compilation, sectorMap, sectorSize, blockSize,
		saveDataSectorOffset, config.Cartridge.BatteryPresent, &bootLogoFound)

	// 生成游戏列表
	itemList := generateItemList(config.Games, blockSize, saveDataSectorOffset)
	copy(compilation[itemListOffset*sectorSize:], itemList)

	// 生成 ROM 代码
	romCode := generateROMCode(status, itemList)

	// 打印信息
	printBuildInfo(config, sectorMap, sectorSize, sectorCount, menuROM, itemList,
		itemListOffset, statusOffset, cartridgeType, result.SectorsUsed, result.GamesAdded)

	// 计算 ROM 大小
	romSize := calculateROMSize(sectorMap, sectorSize)

	// 更新 ROM 头
	updateROMHeader(compilation, romCode)

	// 写入输出文件
	outputFile := strings.Replace(opts.Output, "<CODE>", romCode, -1)
	if err := writeOutput(opts, compilation, romSize, flashSize, outputFile); err != nil {
		return &BuildResult{
			Message: fmt.Sprintf("Error: Failed to write output: %v", err),
			Success: false,
		}, err
	}

	result.ROMCode = romCode
	result.ROMSize = int64(romSize)

	if len(result.Data) > 0 {
		result.Message = fmt.Sprintf("Target rom generated. Some games failed to be included.")
	} else {
		result.Message = "Target rom generated."
	}

	return result, nil
}

// 对齐菜单 ROM
func alignMenuROM(menuROM []byte) []byte {
	// 对齐到 16 字节
	padding := (16 - (len(menuROM) % 16)) % 16
	if padding > 0 {
		menuROM = append(menuROM, make([]byte, padding)...)
		for i := len(menuROM) - padding; i < len(menuROM); i++ {
			menuROM[i] = 0xFF
		}
	}

	// 添加 32 字节用于时间戳
	timestamp := make([]byte, 0x20)
	for i := range timestamp {
		timestamp[i] = 0xFF
	}
	menuROM = append(menuROM, timestamp...)

	return menuROM
}

// 查找菜单 ROM 大小
func findMenuROMSize(menuROM []byte) int {
	marker := []byte("dkARM\x00\x00\x00")
	for i := 0; i <= len(menuROM)-len(marker); i++ {
		if string(menuROM[i:i+len(marker)]) == string(marker) {
			return i + len(marker)
		}
	}
	return len(menuROM)
}

// 创建状态区域
func createStatus(batteryPresent bool) []byte {
	status := make([]byte, 16)
	copy(status[0:4], []byte("KUMA"))
	if batteryPresent {
		status[5] = 0x01
	}
	return status
}

// 检查启动 logo
func checkBootLogo(logo []byte) bool {
	expected := []byte{
		0x17, 0xDA, 0xA0, 0xFE, 0xC0, 0x2F, 0xC3, 0x3C,
		0x0F, 0x6A, 0xBB, 0x54, 0x9A, 0x8B, 0x80, 0xB6,
		0x61, 0x3B, 0x48, 0xEE,
	}
	hash := sha1.Sum(logo)
	return string(hash[:]) == string(expected)
}

// 处理游戏列表
// processGames 处理游戏列表
func processGames(opts BuildOptions, config *Config, compilation []byte, sectorMap []rune,
	sectorSize, blockSize, saveDataSectorOffset int, batteryPresent bool, bootLogoFound *bool) *BuildResult {

	result := &BuildResult{
		Success: true,
		Data:    make([]*GameConfig, 0),
	}

	savesRead := make(map[int]bool)

	// 过滤启用的游戏
	enabledGames := make([]*GameConfig, 0)
	for i := range config.Games {
		if config.Games[i].Enabled {
			enabledGames = append(enabledGames, config.Games[i])
		}
	}

	// 处理每个游戏
	index := 0
	for i := range enabledGames {
		game := enabledGames[i]

		// 检查文件是否存在
		gamePath := filepath.Join(opts.RomBasePath, game.File)
		if _, err := os.Stat(gamePath); os.IsNotExist(err) {
			game.Missing = true
			continue
		}

		// 获取文件大小
		fileInfo, err := os.Stat(gamePath)
		if err != nil {
			game.Missing = true
			continue
		}

		size := int(fileInfo.Size())

		// 调整大小为 2 的幂
		if (size & (size - 1)) != 0 {
			x := 0x80000
			for x < size {
				x *= 2
			}
			size = x
		}

		// 检查最小 ROM 大小
		if size < 0x400000 {
			data, err := os.ReadFile(gamePath)
			if err == nil && strings.Contains(string(data), "Batteryless mod by Lesserkuma") {
				size = max(0x400000, config.Cartridge.MinRomSize)
			} else {
				size = max(size, config.Cartridge.MinRomSize)
			}
		}

		game.Index = index
		game.Size = size
		if game.TitleFont > 0 {
			game.TitleFont--
		}
		game.SectorCount = size / sectorSize

		// 处理隐藏 ROM 按键
		game.KeysBitmap = parseKeys(game.Keys)

		// 处理存档
		if batteryPresent && game.SaveSlot != nil {
			game.SaveType = 2
			saveSlot := *game.SaveSlot - 1
			game.SaveSlot = &saveSlot

			offset := saveDataSectorOffset + saveSlot
			updateSectorMap(sectorMap, offset, 1, 's')

			// 读取存档数据
			if !savesRead[saveSlot] {
				saveDataFile := strings.TrimSuffix(gamePath, filepath.Ext(gamePath)) + ".sav"
				saveData := make([]byte, sectorSize)

				if fileExists(saveDataFile) {
					data, err := os.ReadFile(saveDataFile)
					if err == nil {
						if len(data) < sectorSize {
							copy(saveData, data)
						} else {
							copy(saveData, data[:sectorSize])
						}
					}
				}

				copy(compilation[offset*sectorSize:], saveData)
				savesRead[saveSlot] = true
			}
		} else {
			game.SaveType = 0
			zero := 0
			game.SaveSlot = &zero
		}

		index++
	}

	// 过滤掉缺失的游戏
	validGames := make([]*GameConfig, 0)
	for _, game := range enabledGames {
		if !game.Missing {
			validGames = append(validGames, game)
		}
	}

	if len(validGames) == 0 {
		logp("No ROMs found. Delete the \"%+v\" file to reset your configuration.\n", opts.Config.Games)
		result.Success = false
		return result
	}

	// 重新分配索引
	for i := range validGames {
		validGames[i].Index = i
	}

	// 计算保存结束偏移
	saveEndOffset := saveDataSectorOffset
	for i := len(sectorMap) - 1; i >= 0; i-- {
		if sectorMap[i] == 'S' || sectorMap[i] == 's' {
			saveEndOffset = i + 1
			break
		}
	}

	// 按大小排序（大的在前）
	sort.Slice(validGames, func(i, j int) bool {
		return validGames[i].Size > validGames[j].Size
	})

	// 分配 ROM 空间
	gamesNotFound := make([]*GameConfig, 0)
	for i := range validGames {
		game := validGames[i]
		found := false

		sectorCountMap := game.SectorCount
		if game.Map256M {
			sectorCountMap = (32 * 1024 * 1024) / sectorSize
		}

		for j := saveEndOffset; j < len(sectorMap); j++ {
			if j%sectorCountMap != 0 {
				continue
			}

			// 检查是否有足够的连续空间
			hasSpace := true
			for k := 0; k < game.SectorCount; k++ {
				if j+k >= len(sectorMap) || sectorMap[j+k] != '.' {
					hasSpace = false
					break
				}
			}

			if hasSpace {
				updateSectorMap(sectorMap, j, game.SectorCount, 'r')

				// 读取 ROM 数据
				romData, err := os.ReadFile(filepath.Join(opts.RomBasePath, game.File))
				if err == nil {
					copy(compilation[j*sectorSize:], romData)

					game.SectorOffset = j
					game.BlockOffset = (game.SectorOffset * sectorSize) / blockSize
					game.BlockCount = (sectorCountMap * sectorSize) / blockSize
					found = true

					// 检查启动 logo
					if !*bootLogoFound && len(romData) >= 0xA0 {
						if checkBootLogo(romData[0x04:0xA0]) {
							copy(compilation[0x04:0xA0], romData[0x04:0xA0])
							*bootLogoFound = true
						}
					}
				}
				break
			}
		}

		if !found {
			gamesNotFound = append(gamesNotFound, game)
			logp("\"%s\" couldn't be added because it exceeds the available cartridge space.\n", game.Title)
		}
	}

	if !*bootLogoFound {
		logp("Warning: Valid boot logo is missing!\n")
	}

	// 移除未找到的游戏
	finalGames := make([]*GameConfig, 0)
	for _, game := range validGames {
		found := false
		for _, notFound := range gamesNotFound {
			if game.File == notFound.File {
				found = true
				break
			}
		}
		if !found {
			finalGames = append(finalGames, game)
		}
	}

	// 按索引排序
	sort.Slice(finalGames, func(i, j int) bool {
		return finalGames[i].Index < finalGames[j].Index
	})

	config.Games = finalGames
	result.Data = gamesNotFound
	result.GamesAdded = len(finalGames)

	// 计算使用的扇区数
	sectorsUsed := 0
	for _, c := range sectorMap {
		if c != '.' && c != ' ' {
			sectorsUsed++
		}
	}
	result.SectorsUsed = sectorsUsed

	return result
}

// generateItemList 生成游戏列表
func generateItemList(games []*GameConfig, blockSize, saveDataSectorOffset int) []byte {
	itemList := make([]byte, 0)

	// 收集所有按键组合
	keysSet := make(map[uint16]bool)
	keysSet[0] = true
	for _, game := range games {
		if game.KeysBitmap > 0 {
			keysSet[game.KeysBitmap] = true
		}
	}

	// 转换为切片并排序
	keys := make([]uint16, 0, len(keysSet))
	for k := range keysSet {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	// 为每个按键组合生成条目
	for _, key := range keys {
		for _, game := range games {
			if game.KeysBitmap != key {
				continue
			}

			// 截断标题
			title := game.Title
			if len(title) > 0x30 {
				title = title[:0x2F] + "…"
			}

			// 填充标题到 48 字符
			titleBytes := make([]byte, 0x30*2) // UTF-16LE
			titleRunes := []rune(title)
			for i := 0; i < 0x30; i++ {
				var r rune
				if i < len(titleRunes) {
					r = titleRunes[i]
				} else {
					r = 0
				}
				binary.LittleEndian.PutUint16(titleBytes[i*2:], uint16(r))
			}

			// 构建条目
			entry := make([]byte, 16+0x30*2)
			entry[0] = byte(game.TitleFont)
			entry[1] = byte(len([]rune(game.Title)))
			binary.LittleEndian.PutUint16(entry[2:4], uint16(game.BlockOffset))
			binary.LittleEndian.PutUint16(entry[4:6], uint16(game.BlockCount))
			entry[6] = byte(game.SaveType)
			if game.SaveSlot != nil {
				entry[7] = byte(*game.SaveSlot)
			}
			binary.LittleEndian.PutUint16(entry[8:10], game.KeysBitmap)
			copy(entry[16:], titleBytes)

			itemList = append(itemList, entry...)
		}
	}

	return itemList
}

// generateROMCode 生成 ROM 代码
func generateROMCode(status, itemList []byte) string {
	combined := append(status, itemList...)
	hash := sha1.Sum(combined)
	return fmt.Sprintf("L%s", strings.ToUpper(fmt.Sprintf("%x", hash[:2])[:3]))
}

// printBuildInfo 打印构建信息
func printBuildInfo(config *Config, sectorMap []rune, sectorSize, sectorCount int,
	menuROM, itemList []byte, itemListOffset, statusOffset, cartridgeType, sectorsUsed, gamesAdded int) {

	// 打印扇区映射
	logp("Sector map (1 block = %d KiB):\n", sectorSize/1024)
	for i, c := range sectorMap {
		logp("%c", c)
		if i%64 == 63 {
			logp("\n")
		}
	}
	if len(sectorMap)%64 != 0 {
		logp("\n")
	}

	logp("%.2f%% (%d of %d sectors) used\n\n",
		float64(sectorsUsed)/float64(sectorCount)*100, sectorsUsed, sectorCount)
	logp("Added %d ROM(s) to the compilation\n\n", gamesAdded)

	// 打印游戏列表表头
	if config.Cartridge.BatteryPresent {
		logp("    | Offset     | Map Size  | Save Slot      | Title\n")
	} else {
		logp("    | Offset     | Map Size  | Title\n")
	}

	// 收集按键组合
	keysSet := make(map[uint16]bool)
	keysSet[0] = true
	for _, game := range config.Games {
		if game.KeysBitmap > 0 {
			keysSet[game.KeysBitmap] = true
		}
	}

	keys := make([]uint16, 0, len(keysSet))
	for k := range keysSet {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	// 打印游戏列表
	for _, key := range keys {
		c := 0
		for _, game := range config.Games {
			if game.KeysBitmap != key {
				continue
			}

			title := game.Title
			if len(title) > 0x30 {
				title = title[:0x2F] + "…"
			}

			if c%8 == 0 {
				if game.KeysBitmap != 0 {
					if config.Cartridge.BatteryPresent {
						logp("----+------------+-----------+----------------+--------[Hidden]-\n")
					} else {
						logp("----+------------+-----------+-----------------[Hidden]---------\n")
					}
				} else {
					if config.Cartridge.BatteryPresent {
						logp("----+------------+-----------+----------------+--------------------------------\n")
					} else {
						logp("----+------------+-----------+-------------------------------------------------\n")
					}
				}
			}

			blockSize := 0x80000
			line := fmt.Sprintf("%3d | 0x%08X | 0x%08X | ",
				game.Index+1,
				game.BlockOffset*blockSize,
				game.BlockCount*blockSize)

			if config.Cartridge.BatteryPresent {
				if game.SaveType > 0 && game.SaveSlot != nil {
					line += fmt.Sprintf("%2d (0x%07X) | ", *game.SaveSlot+1, (statusOffset+1+*game.SaveSlot)*sectorSize)
				} else {
					line += "               | "
				}
			}

			line += title
			logp("%s\n", line)
			c++
		}
	}

	logp("\n")
	logp("Menu ROM:        0x%08X–0x%08X\n", 0, len(menuROM))
	logp("Game List:       0x%08X–0x%08X\n",
		itemListOffset*sectorSize, itemListOffset*sectorSize+len(itemList))
	logp("Status Area:     0x%08X–0x%08X\n",
		statusOffset*sectorSize, statusOffset*sectorSize+0x1000)
	logp("\n")

	batteryStr := "without battery"
	if config.Cartridge.BatteryPresent {
		batteryStr = "with battery"
	}
	logp("Cartridge Type:  %d (%s) %s\n",
		cartridgeType+1, preset.CartridgeTypes[cartridgeType].Name, batteryStr)
}

// calculateROMSize 计算 ROM 大小
func calculateROMSize(sectorMap []rune, sectorSize int) int {
	lastUsed := 0
	for i := len(sectorMap) - 1; i >= 0; i-- {
		if sectorMap[i] != '.' {
			lastUsed = i + 1
			break
		}
	}
	return lastUsed * sectorSize
}

// updateROMHeader 更新 ROM 头
func updateROMHeader(compilation []byte, romCode string) {
	// 更新 ROM 代码
	copy(compilation[0xAC:0xB0], []byte(romCode))

	// 计算校验和
	checksum := 0
	for i := 0xA0; i < 0xBD; i++ {
		checksum = checksum - int(compilation[i])
	}
	checksum = (checksum - 0x19) & 0xFF
	compilation[0xBD] = byte(checksum)
}

// writeOutput 写入输出文件
func writeOutput(opts BuildOptions, compilation []byte, romSize, flashSize int, outputFile string) error {
	logp("Output ROM Size: %.2f MiB\n", float64(romSize)/1024/1024)
	logp("Output ROM Code: %s\n", string(compilation[0xAC:0xB0]))

	if opts.Split {
		// 分割输出
		for i := 0; i < int(math.Ceil(float64(flashSize)/0x2000000)); i++ {
			pos := i * 0x2000000
			size := 0x2000000

			if pos >= romSize {
				break
			}
			if pos+size > romSize {
				size = romSize - pos
			}

			ext := filepath.Ext(outputFile)
			base := strings.TrimSuffix(outputFile, ext)
			partFile := fmt.Sprintf("%s_part%d%s", base, i, ext)

			if err := os.WriteFile(partFile, compilation[pos:pos+size], 0644); err != nil {
				return fmt.Errorf("failed to write part %d: %w", i, err)
			}
		}
	} else {
		// 单文件输出
		if err := os.WriteFile(outputFile, compilation[:romSize], 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
