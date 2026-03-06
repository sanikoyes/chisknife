package sram

import "bytes"

// GBA 存档类型
type SaveType int

const (
	FLASH_V120 SaveType = iota
	FLASH_V121
	FLASH_V123
	FLASH_V124
	FLASH_V125
	FLASH_V126
	FLASH512_V130
	FLASH512_V131
	FLASH512_V133
	FLASH1M_V102
	FLASH1M_V103

	EEPROM_V111
	EEPROM_V120
	EEPROM_V121
	EEPROM_V122
	EEPROM_V124
	EEPROM_V126

	SRAM_V110
	SRAM_V111
	SRAM_V112
	SRAM_V113
	FRAM_V100
	FRAM_V102
	FRAM_V103
	FRAM_V110
	NO_SAVE
)

// 检查存档类型是否为FLASH类型
func (t SaveType) IsFLASH() bool {
	return t >= FLASH_V120 && t <= FLASH1M_V103
}

// 检查存档类型是否为SRAM类型
func (t SaveType) IsSRAM() bool {
	return t >= SRAM_V110 && t <= SRAM_V113
}

// 检查存档类型是否为EEPROM类型
func (t SaveType) IsEEPROM() bool {
	return t >= EEPROM_V111 && t <= EEPROM_V126
}

// 存档类型字节模式
var saveTypePatterns = map[SaveType][]byte{
	FLASH_V120:    []byte("FLASH_V120"),
	FLASH_V121:    []byte("FLASH_V121"),
	FLASH_V123:    []byte("FLASH_V123"),
	FLASH_V124:    []byte("FLASH_V124"),
	FLASH_V125:    []byte("FLASH_V125"),
	FLASH_V126:    []byte("FLASH_V126"),
	FLASH512_V130: []byte("FLASH512_V130"),
	FLASH512_V131: []byte("FLASH512_V131"),
	FLASH512_V133: []byte("FLASH512_V133"),
	FLASH1M_V102:  []byte("FLASH1M_V102"),
	FLASH1M_V103:  []byte("FLASH1M_V103"),
	EEPROM_V111:   []byte("EEPROM_V111"),
	EEPROM_V120:   []byte("EEPROM_V120"),
	EEPROM_V121:   []byte("EEPROM_V121"),
	EEPROM_V122:   []byte("EEPROM_V122"),
	EEPROM_V124:   []byte("EEPROM_V124"),
	EEPROM_V126:   []byte("EEPROM_V126"),
	SRAM_V110:     []byte("SRAM_V110"),
	SRAM_V111:     []byte("SRAM_V111"),
	SRAM_V112:     []byte("SRAM_V112"),
	SRAM_V113:     []byte("SRAM_V113"),
	FRAM_V100:     []byte("SRAM_F_V100"),
	FRAM_V102:     []byte("SRAM_F_V102"),
	FRAM_V103:     []byte("SRAM_F_V103"),
	FRAM_V110:     []byte("SRAM_F_V110"),
}

var saveTypesList = []SaveType{
	FLASH_V120, FLASH_V121, FLASH_V123, FLASH_V124, FLASH_V125, FLASH_V126,
	FLASH512_V130, FLASH512_V131, FLASH512_V133,
	FLASH1M_V102, FLASH1M_V103,
	EEPROM_V111, EEPROM_V120, EEPROM_V121, EEPROM_V122, EEPROM_V124, EEPROM_V126,
	SRAM_V110, SRAM_V111, SRAM_V112, SRAM_V113,
	FRAM_V100, FRAM_V102, FRAM_V103, FRAM_V110,
	NO_SAVE,
}

// 检测 ROM 的存档类型
func DetectSaveType(romData []byte) SaveType {
	for saveType, pattern := range saveTypePatterns {
		if bytes.Contains(romData, pattern) {
			return saveType
		}
	}
	return NO_SAVE
}
