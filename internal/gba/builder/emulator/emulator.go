// Package emulator 提供 GBA 模拟器 ROM 构建功能
// 支持 Goomba (GB/GBC) 和 PocketNES (NES) 模拟器
package emulator

import (
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
)

const (
	emuHeaderSize = 0x30
	nesHeaderSize = 0x10
)

// 构建 Goomba 模拟器 ROM (GB/GBC)
func BuildGoomba(romPaths []string, outPath string, goombaPath string) error {
	// 读取 Goomba 模拟器基础 ROM
	goombaData, err := os.ReadFile(goombaPath)
	if err != nil {
		return err
	}

	buildRom := make([]byte, len(goombaData))
	copy(buildRom, goombaData)

	// 追加所有 ROM 数据
	for _, romPath := range romPaths {
		romData, err := os.ReadFile(romPath)
		if err != nil {
			return err
		}
		buildRom = append(buildRom, romData...)
	}

	// 写入最终文件
	return os.WriteFile(outPath, buildRom, 0644)
}

// PocketNES ROM 数据库记录
type PocketNESDBRecord struct {
	CRC       string
	Title     string
	Flag      int
	DefFollow int
}

// 从数据库文件加载 PocketNES ROM 信息
func loadPocketNESDB(dbPath string) (map[string]PocketNESDBRecord, error) {
	data, err := os.ReadFile(dbPath)
	if err != nil {
		return nil, err
	}

	db := make(map[string]PocketNESDBRecord)
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if !strings.Contains(line, "|") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}

		record := PocketNESDBRecord{
			CRC:   parts[0],
			Title: parts[1],
		}

		if len(parts) > 2 && strings.TrimSpace(parts[2]) != "" {
			record.Flag = parseInt(strings.Fields(parts[2])[0])
		}

		if len(parts) > 3 && strings.TrimSpace(parts[3]) != "" {
			record.DefFollow = parseInt(strings.Fields(parts[3])[0])
		}

		db[record.CRC] = record
	}

	return db, nil
}

// 构建 PocketNES 模拟器 ROM (NES)
func BuildPocketNES(romPaths []string, outPath string, pocketnesPath string, romdataDB string) error {
	// 读取 PocketNES 模拟器基础 ROM
	pocketnesData, err := os.ReadFile(pocketnesPath)
	if err != nil {
		return err
	}

	buildRom := make([]byte, len(pocketnesData))
	copy(buildRom, pocketnesData)

	// 对齐到 256 字节边界
	padding := (256 - ((len(buildRom) + emuHeaderSize + nesHeaderSize) % 256)) % 256
	buildRom = append(buildRom, make([]byte, padding)...)

	// 加载数据库
	db, err := loadPocketNESDB(romdataDB)
	if err != nil {
		// 如果数据库加载失败，继续但不使用数据库
		db = make(map[string]PocketNESDBRecord)
	}

	// 处理每个 ROM
	for _, romPath := range romPaths {
		romData, err := os.ReadFile(romPath)
		if err != nil {
			return err
		}

		// 检查并移除 NES 头部
		romDataNoHeader := romData
		if len(romData) >= 4 && string(romData[0:4]) == "NES\x1a" {
			romDataNoHeader = romData[nesHeaderSize:]
		}

		// 计算 CRC32
		crc := crc32.ChecksumIEEE(romDataNoHeader)
		crcStr := strings.ToLower(strings.TrimPrefix(strings.ToLower(strings.TrimPrefix(formatHex(crc), "0x")), "0x"))

		// 从数据库查找信息
		title := []byte{}
		flag := 0
		defFollow := 0

		if record, found := db[crcStr]; found {
			title = []byte(record.Title)
			flag = record.Flag
			defFollow = record.DefFollow
		} else {
			// 从文件名生成标题
			titleText := strings.TrimSuffix(filepath.Base(romPath), filepath.Ext(romPath))
			title = []byte(titleText)

			// 检查区域标记
			if strings.Contains(titleText, "(E)") ||
				strings.Contains(titleText, "(Europe)") ||
				strings.Contains(titleText, "(EUR)") {
				flag = flag | (1 << 2)
			}
		}

		// 限制标题长度为 31 字节
		if len(title) > 31 {
			title = title[:31]
		}

		// 对齐 ROM 数据
		finRom := append(romData, make([]byte, (256-((len(romData)+emuHeaderSize)%256))%256)...)

		// 构建头部
		header := make([]byte, emuHeaderSize)
		copy(header[0:31], title)
		// header[31] = 0 (已经是零值)
		binary.LittleEndian.PutUint32(header[32:36], uint32(len(finRom)))
		binary.LittleEndian.PutUint32(header[36:40], uint32(flag))
		binary.LittleEndian.PutUint32(header[40:44], uint32(defFollow))
		binary.LittleEndian.PutUint32(header[44:48], 0)

		// 追加到构建 ROM
		buildRom = append(buildRom, header...)
		buildRom = append(buildRom, finRom...)
	}

	// 写入最终文件
	return os.WriteFile(outPath, buildRom, 0644)
}

// 辅助函数：解析整数
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			break
		}
	}
	return result
}

// 辅助函数：格式化十六进制
func formatHex(n uint32) string {
	const hexDigits = "0123456789abcdef"
	result := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		result[i] = hexDigits[n&0xf]
		n >>= 4
	}
	return string(result)
}
