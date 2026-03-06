package sram

import (
	"encoding/binary"
	"errors"
	"unsafe"
)

// 应用 SRAM 补丁
func PatchSRAM(romData *[]byte, sramBankType byte) error {
	for _, saveType := range saveTypesList {
		pattern, ok := saveTypePatterns[saveType]
		if !ok || len(pattern) == 0 {
			continue
		}

		if FindBytes(*romData, pattern, nil).Valid {
			if err := PatchSRAMByType(romData, saveType, sramBankType, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// 根据存档类型应用 SRAM 补丁
func PatchSRAMByType(romData *[]byte, saveType SaveType, sramBankType byte, interchangeableEmptyByte bool) error {
	switch saveType {
	case FLASH_V120, FLASH_V121:
		return patchSRAMBySet(romData, SRAM_PATCHES_FLASH_V12X)

	case FLASH_V123, FLASH_V124:
		return patchSRAMBySet(romData, SRAM_PATCHES_FLASH_V12Y)

	case FLASH_V125, FLASH_V126:
		return patchSRAMBySet(romData, SRAM_PATCHES_FLASH_V12Z)

	case FLASH512_V130, FLASH512_V131, FLASH512_V133:
		return patchSRAMBySet(romData, SRAM_PATCHES_FLASH512_V13X)

	case FLASH1M_V102:
		if sramBankType == 1 {
			return patchSRAMBySet(romData, SRAM_BANK_1_PATCHES_FLASH1M_V102)
		}
		return patchSRAMBySet(romData, SRAM_PATCHES_FLASH1M_V102)

	case FLASH1M_V103:
		if sramBankType == 1 {
			return patchSRAMBySet(romData, SRAM_BANK_1_PATCHES_FLASH1M_V103)
		}
		return patchSRAMBySet(romData, SRAM_PATCHES_FLASH1M_V103)

	case EEPROM_V111:
		return patchEEPROMV111(romData, interchangeableEmptyByte)

	case EEPROM_V120, EEPROM_V121, EEPROM_V122:
		return patchSRAMBySet(romData, SRAM_PATCHES_EEPROM_V12X)

	case EEPROM_V124:
		return patchSRAMBySet(romData, SRAM_PATCHES_EEPROM_V124)

	case EEPROM_V126:
		return patchSRAMBySet(romData, SRAM_PATCHES_EEPROM_V126)

	case NO_SAVE, SRAM_V110, SRAM_V111, SRAM_V112, SRAM_V113,
		FRAM_V100, FRAM_V102, FRAM_V103, FRAM_V110:
		// 这些类型不需要补丁
		return nil
	}

	return nil
}

// 应用一组补丁
func patchSRAMBySet(romData *[]byte, patchSet []RomPatch) error {
	for _, patch := range patchSet {
		ReplaceBytes(*romData, patch.FindData, patch.ReplacementData, patch.FindMask, patch.ReplacementMask)
	}
	return nil
}

// EEPROM V111 特殊补丁
func patchEEPROMV111(romData *[]byte, interchangeableEmptyByte bool) error {
	find1 := []byte{0x0e, 0x48, 0x39, 0x68, 0x01, 0x60, 0x0e, 0x48}
	replacement1 := []byte{0x00, 0x48, 0x00, 0x47, 0, 0, 0, 0x08}
	find2 := []byte{0x27, 0xe0, 0xd0, 0x20, 0x00, 0x05, 0x01, 0x88}
	replacement2 := []byte{0x27, 0xe0, 0xe0, 0x20, 0x00, 0x05, 0x01, 0x88}

	footer := []byte{
		0x39, 0x68, 0x27, 0x48, 0x81, 0x42, 0x23, 0xd0, 0x89, 0x1c, 0x08, 0x88, 0x01, 0x28, 0x02, 0xd1,
		0x24, 0x48, 0x78, 0x60, 0x33, 0xe0, 0x00, 0x23, 0x00, 0x22, 0x89, 0x1c, 0x10, 0xb4, 0x01, 0x24,
		0x08, 0x68, 0x20, 0x40, 0x5b, 0x00, 0x03, 0x43, 0x89, 0x1c, 0x52, 0x1c, 0x06, 0x2a, 0xf7, 0xd1,
		0x10, 0xbc, 0x39, 0x60, 0xdb, 0x01, 0x02, 0x20, 0x00, 0x02, 0x1b, 0x18, 0x0e, 0x20, 0x00, 0x06,
		0x1b, 0x18, 0x7b, 0x60, 0x39, 0x1c, 0x08, 0x31, 0x08, 0x88, 0x09, 0x38, 0x08, 0x80, 0x16, 0xe0,
		0x15, 0x49, 0x00, 0x23, 0x00, 0x22, 0x10, 0xb4, 0x01, 0x24, 0x08, 0x68, 0x20, 0x40, 0x5b, 0x00,
		0x03, 0x43, 0x89, 0x1c, 0x52, 0x1c, 0x06, 0x2a, 0xf7, 0xd1, 0x10, 0xbc, 0xdb, 0x01, 0x02, 0x20,
		0x00, 0x02, 0x1b, 0x18, 0x0e, 0x20, 0x00, 0x06, 0x1b, 0x18, 0x08, 0x3b, 0x3b, 0x60, 0x0b, 0x48,
		0x39, 0x68, 0x01, 0x60, 0x0a, 0x48, 0x79, 0x68, 0x01, 0x60, 0x0a, 0x48, 0x39, 0x1c, 0x08, 0x31,
		0x0a, 0x88, 0x80, 0x21, 0x09, 0x06, 0x0a, 0x43, 0x02, 0x60, 0x07, 0x48, 0x00, 0x47, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x04, 0x00, 0x00, 0x0e, 0xd4, 0x00, 0x00, 0x04,
		0xd8, 0x00, 0x00, 0x04, 0xdc, 0x00, 0x00, 0x04, 0, 0, 0, 0x08,
	}

	romEOD := FindRomEOD(*romData, interchangeableEmptyByte)
	footerOffset := NextAlignedAddress(romEOD+1, 16)

	replace1Result := ReplaceBytes(*romData, find1, replacement1, nil, nil)
	ReplaceBytes(*romData, find2, replacement2, nil, nil)

	if !replace1Result.Valid {
		return errors.New("failed to find leading pattern in ROM for EEPROM_V111 to SRAM patch")
	}

	off1 := replace1Result.Offset
	footerOff := int32(footerOffset)
	off1Int := int32(off1)

	// 检测字节序
	isBigEndian := func() bool {
		var i uint16 = 0x0102
		buf := (*[2]byte)(unsafe.Pointer(&i))
		return buf[0] == 1
	}()

	// 确保有足够空间
	if footerOffset+len(footer) > len(*romData) {
		newData := make([]byte, footerOffset+len(footer))
		copy(newData, *romData)
		*romData = newData
	}

	// 写入 footer
	copy((*romData)[footerOffset:], footer)

	// 修改第一个补丁
	if !isBigEndian {
		(*romData)[off1+4] = byte((footerOff + 1) >> 0 & 0xff)
		(*romData)[off1+5] = byte((footerOff + 1) >> 8 & 0xff)
		(*romData)[off1+6] = byte((footerOff + 1) >> 16 & 0xff)
	} else {
		(*romData)[off1+4] = byte((footerOff + 1) >> 24 & 0xff)
		(*romData)[off1+5] = byte((footerOff + 1) >> 16 & 0xff)
		(*romData)[off1+6] = byte((footerOff + 1) >> 8 & 0xff)
	}

	// 修改 footer 补丁
	if !isBigEndian {
		(*romData)[footerOffset+184] = byte((off1Int + 33) >> 0 & 0xff)
		(*romData)[footerOffset+185] = byte((off1Int + 33) >> 8 & 0xff)
		(*romData)[footerOffset+186] = byte((off1Int + 33) >> 16 & 0xff)
	} else {
		(*romData)[footerOffset+184] = byte((off1Int + 33) >> 24 & 0xff)
		(*romData)[footerOffset+185] = byte((off1Int + 33) >> 16 & 0xff)
		(*romData)[footerOffset+186] = byte((off1Int + 33) >> 8 & 0xff)
	}

	return nil
}

func init() {
	// 初始化时检测字节序
	_ = binary.BigEndian
}
