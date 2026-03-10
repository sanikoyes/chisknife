// Package gbapatch 提供 GBA ROM 补丁功能
// 完整翻译自 C++ 版本的 gba_patch
package sram

import (
	"chisknife/asset"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

// 补丁选项
type Options struct {
	IPSPath         string
	PatchSRAM       bool
	Uniformize      bool
	PatchComplement bool
	Trim            bool
	InPlace         bool
	DummySave       bool
	SRAMBankType    byte
}

// 字节偏移结果
type ByteOffset struct {
	Valid  bool
	Offset int
}

// ROM 补丁定义
type RomPatch struct {
	FindData        []byte
	ReplacementData []byte
	FindMask        []bool
	ReplacementMask []bool
}

// 修复补码校验
func PatchComplementCheck(romData []byte) error {
	if len(romData) <= 0xbd {
		return errors.New("invalid ROM data; data size too small")
	}

	var sum byte = 0
	for i := 0xa0; i < 0xbd; i++ {
		sum -= romData[i]
	}
	sum -= 0x19

	romData[0xbd] = sum
	return nil
}

// 查找 ROM 数据结束位置
func FindRomEOD(romData []byte, interchangeableEmptyByte bool) int {
	if len(romData) == 0 {
		return 0
	}

	endByte := romData[len(romData)-1]

	if endByte == 0xff || endByte == 0x00 {
		for i := len(romData) - 1; i >= 0; i-- {
			if (!interchangeableEmptyByte && romData[i] != endByte) ||
				(interchangeableEmptyByte && romData[i] != 0xff && romData[i] != 0x00) {
				return i
			}
		}
		return 0
	}

	if len(romData) > 0 {
		return len(romData) - 1
	}
	return 0
}

// 获取下一个对齐地址
func NextAlignedAddress(address, alignment int) int {
	aligned := address
	for aligned%alignment != 0 {
		aligned++
	}
	return aligned
}

// 统一 ROM 填充
func UniformizeRomPadding(romData []byte, alignment int) {
	if len(romData) == 0 {
		return
	}

	empty := romData[len(romData)-1]
	romEOD := FindRomEOD(romData, false)

	var begin int
	if alignment > 0 {
		begin = NextAlignedAddress(romEOD, alignment)
	} else {
		begin = romEOD
	}

	if begin < len(romData) {
		for i := begin; i < len(romData); i++ {
			if romData[i] != empty {
				romData[i] = empty
			}
		}
	}
}

// 修剪 ROM 填充
func TrimPadding(romData *[]byte, alignment int, interchangeableEmptyByte bool) {
	if len(*romData) == 0 {
		return
	}

	romEOD := FindRomEOD(*romData, interchangeableEmptyByte)
	var cutoff int
	if alignment > 0 {
		cutoff = NextAlignedAddress(romEOD, alignment) - 1
	} else {
		cutoff = romEOD
	}

	if cutoff+1 < len(*romData) {
		*romData = (*romData)[:cutoff+1]
	}
}

// 应用 IPS 补丁
func ApplyIPSPatch(data *[]byte, ipsPatch []byte) error {
	if len(ipsPatch) < 5 {
		return errors.New("missing IPS header (5 bytes) at position 0x00")
	}

	// 检查 IPS 头部 "PATCH"
	if string(ipsPatch[0:5]) != "PATCH" {
		return errors.New("IPS patch has invalid header")
	}

	readPos := 5

	for readPos < len(ipsPatch) {
		// 读取 3 字节偏移量（大端序）
		if readPos+3 > len(ipsPatch) {
			return fmt.Errorf("insufficient bytes for offset reading (3 bytes) at IPS patch read position %d", readPos)
		}

		// 检查 EOF 标记
		if ipsPatch[readPos] == 0x45 && ipsPatch[readPos+1] == 0x4F && ipsPatch[readPos+2] == 0x46 {
			readPos += 3
			break
		}

		writePos := uint32(ipsPatch[readPos])<<16 | uint32(ipsPatch[readPos+1])<<8 | uint32(ipsPatch[readPos+2])
		readPos += 3

		// 读取 2 字节长度（大端序）
		if readPos+2 > len(ipsPatch) {
			return fmt.Errorf("insufficient bytes for patch size reading (2 bytes) at IPS patch read position %d", readPos)
		}

		writeSize := uint16(ipsPatch[readPos])<<8 | uint16(ipsPatch[readPos+1])
		readPos += 2

		if writeSize > 0 {
			// 普通补丁
			if readPos+int(writeSize) > len(ipsPatch) {
				return fmt.Errorf("IPS patch data has insufficient bytes for patch at read position %d length %d", readPos, writeSize)
			}

			// 扩展数据切片如果需要
			if int(writePos)+int(writeSize) > len(*data) {
				newData := make([]byte, int(writePos)+int(writeSize))
				copy(newData, *data)
				*data = newData
			}

			copy((*data)[writePos:], ipsPatch[readPos:readPos+int(writeSize)])
			readPos += int(writeSize)
		} else {
			// RLE 补丁
			if readPos+3 > len(ipsPatch) {
				return fmt.Errorf("insufficient bytes for RLE patch at IPS patch read position %d", readPos)
			}

			rleSize := uint16(ipsPatch[readPos])<<8 | uint16(ipsPatch[readPos+1])
			rleValue := ipsPatch[readPos+2]
			readPos += 3

			// 扩展数据切片如果需要
			if int(writePos)+int(rleSize) > len(*data) {
				newData := make([]byte, int(writePos)+int(rleSize))
				copy(newData, *data)
				*data = newData
			}

			for i := 0; i < int(rleSize); i++ {
				(*data)[int(writePos)+i] = rleValue
			}
		}
	}

	// 处理截断扩展（Lunar IPS 兼容）
	if readPos+3 <= len(ipsPatch) {
		truncateLength := uint32(ipsPatch[readPos])<<16 | uint32(ipsPatch[readPos+1])<<8 | uint32(ipsPatch[readPos+2])
		if int(truncateLength) < len(*data) {
			*data = (*data)[:truncateLength]
		}
	}

	return nil
}

// 在数据中查找字节模式（Boyer-Moore 算法）
func FindBytes(data []byte, pattern []byte, wildcardMask []bool) ByteOffset {
	useMask := wildcardMask != nil && len(wildcardMask) == len(pattern)

	if len(pattern) > len(data) {
		return ByteOffset{Valid: false, Offset: 0}
	}

	dataLen := len(data)
	findLen := len(pattern)

	dataIdx := findLen - 1
	findIdx := findLen - 1

	for dataIdx < dataLen {
		if pattern[findIdx] == data[dataIdx] || (useMask && wildcardMask[findIdx]) {
			if findIdx == 0 {
				return ByteOffset{Valid: true, Offset: dataIdx}
			}
			dataIdx--
			findIdx--
		} else {
			lastMatchIdx := 0
			for i := findLen - 1; i > 0; i-- {
				if pattern[i] == data[dataIdx] || (useMask && wildcardMask[i]) {
					lastMatchIdx = i
					break
				}
			}

			if findIdx < lastMatchIdx+1 {
				dataIdx += findLen - findIdx
			} else {
				dataIdx += findLen - lastMatchIdx - 1
			}
			findIdx = findLen - 1
		}
	}

	return ByteOffset{Valid: false, Offset: 0}
}

// 替换字节模式
func ReplaceBytes(data []byte, findData []byte, replacementData []byte, findMask []bool, replacementMask []bool) ByteOffset {
	findResult := FindBytes(data, findData, findMask)

	if !findResult.Valid {
		return findResult
	}

	idx := findResult.Offset
	useReplacementMask := replacementMask != nil && len(replacementMask) == len(replacementData)

	// 替换数据
	for i := 0; i < len(replacementData) && idx+i < len(data); i++ {
		if !useReplacementMask || !replacementMask[i] {
			data[idx+i] = replacementData[i]
		}
	}

	return findResult
}

// 写入虚拟存档文件
func WriteDummySave(filePath string, size int) error {
	data := make([]byte, size)
	for i := range data {
		data[i] = 0xff
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// 处理 ROM 文件
func ProcessROM(inputPath string, outputPath string, opts Options) error {
	// 读取 ROM 文件
	romData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading ROM file: %w", err)
	}

	// 应用 IPS 补丁
	if opts.IPSPath != "" {
		ipsData, err := asset.SramIpsFS.ReadFile(opts.IPSPath)
		if err != nil {
			return fmt.Errorf("error reading IPS patch file: %w", err)
		}
		if err := ApplyIPSPatch(&romData, ipsData); err != nil {
			return fmt.Errorf("failed to apply IPS patch: %w", err)
		}
	}

	// 统一填充
	if opts.Uniformize {
		UniformizeRomPadding(romData, 16)
	}

	// 应用 SRAM 补丁
	if opts.PatchSRAM {
		if err := PatchSRAM(&romData, opts.SRAMBankType); err != nil {
			return fmt.Errorf("error during SRAM patching: %w", err)
		}
	}

	// 修复补码校验
	if opts.PatchComplement {
		if err := PatchComplementCheck(romData); err != nil {
			return fmt.Errorf("error during complement check patch: %w", err)
		}
	}

	// 修剪填充
	if opts.Trim {
		TrimPadding(&romData, 16, true)
	}

	// 写入输出文件
	outPath := outputPath
	if opts.InPlace {
		outPath = inputPath
	}

	if err := os.WriteFile(outPath, romData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// 写入虚拟存档
	if opts.DummySave {
		dir := filepath.Dir(outPath)
		base := filepath.Base(outPath)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		savePath := filepath.Join(dir, "saver", name+".sav")

		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			if err := WriteDummySave(savePath, 512); err != nil {
				return fmt.Errorf("failed to write dummy save: %w", err)
			}
		}
	}

	return nil
}

// 简单的 SRAM 补丁接口
func SRAMPatch(romPath string, outPath string) error {
	opts := Options{
		PatchSRAM:       true,
		PatchComplement: true,
	}
	return ProcessROM(romPath, outPath, opts)
}

// 带 SRAM bank 类型的补丁接口
func SRAMPatchBank(romPath string, outPath string, sramBankType byte) error {
	opts := Options{
		PatchSRAM:       true,
		PatchComplement: true,
		SRAMBankType:    sramBankType,
	}
	return ProcessROM(romPath, outPath, opts)
}

// IPS 补丁接口
func IPSPatch(romPath string, ipsPath string, outPath string) error {
	opts := Options{
		IPSPath:         ipsPath,
		PatchSRAM:       false,
		PatchComplement: true,
	}
	return ProcessROM(romPath, outPath, opts)
}

// 检测系统字节序
func isBigEndian() bool {
	var i uint16 = 0x0102
	buf := (*[2]byte)(unsafe.Pointer(&i))
	return buf[0] == 1
}

func init() {
	// 初始化时检测字节序
	_ = binary.BigEndian
}
