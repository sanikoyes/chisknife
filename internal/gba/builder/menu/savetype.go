package menu

import (
	"bytes"
	"os"
)

// 存档类型
type SaveType string

const (
	SaveTypeNone    SaveType = "none"
	SaveTypeSRAM    SaveType = "sram"
	SaveTypeFlash   SaveType = "flash"
	SaveTypeFlash1M SaveType = "flash1m"
	SaveTypeEEPROM  SaveType = "eeprom"
)

// 存档类型检测模式
var saveTypePatterns = []struct {
	pattern  []byte
	saveType SaveType
}{
	{[]byte("FLASH1M_V1"), SaveTypeFlash1M},
	{[]byte("EEPROM_V1"), SaveTypeEEPROM},
	{[]byte("FLASH_V1"), SaveTypeFlash},
	{[]byte("FLASH512_V1"), SaveTypeFlash},
	{[]byte("SRAM_V1"), SaveTypeSRAM},
	{[]byte("SRAM_F_V1"), SaveTypeSRAM},
}

// 检测 ROM 的存档类型
func CheckSaveType(romPath string) SaveType {
	data, err := os.ReadFile(romPath)
	if err != nil {
		return SaveTypeNone
	}

	for _, p := range saveTypePatterns {
		if bytes.Contains(data, p.pattern) {
			return p.saveType
		}
	}

	return SaveTypeNone
}
