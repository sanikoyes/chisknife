// Package batteryless 提供 GBA ROM 的 batteryless 补丁功能
// 将游戏存档从电池供电的 SRAM 转换为 Flash 存储
package batteryless

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"

	"chisknife/asset"
)

const (
	maxROMSize = 0x02000000
	alignment  = 0x40000
)

// 定义 payload 中的偏移量
const (
	OriginalEntrypointAddr = iota
	FlushMode
	SaveSize
	PatchedEntrypoint
	WriteSRAMPatched
	WriteEEPROMPatched
	WriteFlashPatched
	WriteEEPROMV111PostHook
)

var (
	// 签名标识
	signature = []byte{'<', '3', ' ', 'f', 'r', 'o', 'm', ' ', 'M', 'a', 'n', 'i', 'a', 'c'}

	// 分支跳转指令
	thumbBranchThunk = []byte{0x00, 0x4b, 0x18, 0x47}
	armBranchThunk   = []byte{0x00, 0x30, 0x9f, 0xe5, 0x13, 0xff, 0x2f, 0xe1}

	// 写入函数签名
	writeSRAMSignature = []byte{
		0x30, 0xB5, 0x05, 0x1C, 0x0C, 0x1C, 0x13, 0x1C, 0x0B, 0x4A, 0x10, 0x88, 0x0B, 0x49, 0x08, 0x40,
	}
	writeSRAM2Signature = []byte{
		0x80, 0xb5, 0x83, 0xb0, 0x6f, 0x46, 0x38, 0x60, 0x79, 0x60, 0xba, 0x60, 0x09, 0x48, 0x09, 0x49,
	}
	writeSRAMRAMSignature = []byte{
		0x04, 0xC0, 0x90, 0xE4, 0x01, 0xC0, 0xC1, 0xE4, 0x2C, 0xC4, 0xA0, 0xE1, 0x01, 0xC0, 0xC1, 0xE4,
	}
	writeEEPROMSignature = []byte{
		0x70, 0xB5, 0x00, 0x04, 0x0A, 0x1C, 0x40, 0x0B, 0xE0, 0x21, 0x09, 0x05, 0x41, 0x18, 0x07, 0x31, 0x00, 0x23, 0x10, 0x78,
	}
	writeFlashSignature = []byte{
		0x70, 0xB5, 0x00, 0x03, 0x0A, 0x1C, 0xE0, 0x21, 0x09, 0x05, 0x41, 0x18, 0x01, 0x23, 0x1B, 0x03,
	}
	writeFlash2Signature = []byte{
		0x7C, 0xB5, 0x90, 0xB0, 0x00, 0x03, 0x0A, 0x1C, 0xE0, 0x21, 0x09, 0x05, 0x09, 0x18, 0x01, 0x23,
	}
	writeFlash3Signature = []byte{
		0xF0, 0xB5, 0x90, 0xB0, 0x0F, 0x1C, 0x00, 0x04, 0x04, 0x0C, 0x03, 0x48, 0x00, 0x68, 0x40, 0x89,
	}
	writeEEPROMV11EpiloguePatch = []byte{0x07, 0x49, 0x08, 0x47}
	writeEEPROMV111Signature    = []byte{
		0x0A, 0x88, 0x80, 0x21, 0x09, 0x06, 0x0A, 0x43, 0x02, 0x60, 0x07, 0x48, 0x00, 0x47, 0x00, 0x00,
	}
)

// 处理 ROM 补丁操作
type ROMPatcher struct {
	romData []byte
	romSize uint32
	payload []byte
}

// 创建新的 ROM 补丁器
func NewROMPatcher() *ROMPatcher {
	return &ROMPatcher{
		payload: asset.PayloadBatteryLess,
	}
}

// 在数据中查找模式
func findPattern(data, pattern []byte, stride int) int {
	if stride < 1 {
		stride = 1
	}
	for i := 0; i <= len(data)-len(pattern); i += stride {
		if bytes.Equal(data[i:i+len(pattern)], pattern) {
			return i
		}
	}
	return -1
}

// 从字节数组加载 ROM 数据
func (p *ROMPatcher) loadROMFromBytes(data []byte) error {
	p.romSize = uint32(len(data))
	if p.romSize > maxROMSize {
		return errors.New("ROM 太大 - 不是 GBA ROM?")
	}

	// 检查对齐并填充
	if p.romSize&0x3ffff != 0 {
		p.romSize &= ^uint32(0x3ffff)
		p.romSize += alignment
	}

	// 创建最大大小的缓冲区并填充 0xFF
	p.romData = make([]byte, maxROMSize)
	for i := range p.romData {
		p.romData[i] = 0xFF
	}
	copy(p.romData, data)

	return nil
}

// 返回修补后的 ROM 字节数据
func (p *ROMPatcher) getROMBytes() []byte {
	result := make([]byte, p.romSize)
	copy(result, p.romData[:p.romSize])
	return result
}

// 加载 ROM 文件
func (p *ROMPatcher) loadROM(romPath string) error {
	data, err := os.ReadFile(romPath)
	if err != nil {
		return fmt.Errorf("无法打开输入文件: %w", err)
	}
	return p.loadROMFromBytes(data)
}

// 保存 ROM 文件
func (p *ROMPatcher) saveROM(outPath string) error {
	return os.WriteFile(outPath, p.romData[:p.romSize], 0644)
}

// 检查是否已经打过补丁
func (p *ROMPatcher) isAlreadyPatched() bool {
	return findPattern(p.romData, signature, 4) != -1
}

// 修补 IRQ 处理程序
func (p *ROMPatcher) patchIRQHandler() error {
	oldIRQAddr := []byte{0xfc, 0x7f, 0x00, 0x03}
	newIRQAddr := []byte{0xf4, 0x7f, 0x00, 0x03}

	foundIRQ := 0
	for i := uint32(0); i < p.romSize; i += 4 {
		if bytes.Equal(p.romData[i:i+4], oldIRQAddr) {
			foundIRQ++
			fmt.Printf("在 0x%x 处找到 IRQ 处理程序地址引用，正在修补\n", i)
			copy(p.romData[i:], newIRQAddr)
		}
	}

	if foundIRQ == 0 {
		return errors.New("找不到 IRQ 处理程序的任何引用。ROM 是否已经打过补丁?")
	}
	return nil
}

// 查找 payload 基地址
func (p *ROMPatcher) findPayloadBase() int {
	payloadLen := len(p.payload)
	for payloadBase := int(p.romSize) - alignment - payloadLen; payloadBase >= 0; payloadBase -= alignment {
		isAllZeroes := true
		isAllOnes := true

		for i := 0; i < alignment+payloadLen; i++ {
			if p.romData[payloadBase+i] != 0 {
				isAllZeroes = false
			}
			if p.romData[payloadBase+i] != 0xFF {
				isAllOnes = false
			}
			if !isAllZeroes && !isAllOnes {
				break
			}
		}

		if isAllZeroes || isAllOnes {
			return payloadBase
		}
	}
	return -1
}

// 修补写入函数
func (p *ROMPatcher) patchWriteFunction(offset int, patchBytes []byte, branchTarget, saveSize uint32, payloadBase int, description string) {
	copy(p.romData[offset:], patchBytes)
	binary.LittleEndian.PutUint32(p.romData[offset+len(patchBytes):], 0x08000000+uint32(payloadBase)+branchTarget)
	binary.LittleEndian.PutUint32(p.romData[payloadBase+SaveSize*4:], saveSize)
	fmt.Printf("%s 在偏移 0x%x 处识别，正在修补\n", description, offset)
}

// 修补所有写入函数
func (p *ROMPatcher) patchWriteFunctions(payloadBase, mode int) error {
	foundWriteLocation := false

	// 辅助函数：检查并修补
	checkAndPatch := func(sig, patchBytes []byte, branchTarget, saveSize uint32, description string, isARM bool) bool {
		stride := 2
		if isARM {
			stride = 4
		}

		for offset := 0; offset <= int(p.romSize)-len(sig); offset += stride {
			if bytes.Equal(p.romData[offset:offset+len(sig)], sig) {
				foundWriteLocation = true
				if mode == 0 {
					if isARM {
						copy(p.romData[offset:], patchBytes)
						binary.LittleEndian.PutUint32(p.romData[offset+8:], 0x08000000+uint32(payloadBase)+branchTarget)
					} else {
						p.patchWriteFunction(offset, patchBytes, branchTarget, saveSize, payloadBase, description)
					}
				}
				binary.LittleEndian.PutUint32(p.romData[payloadBase+SaveSize*4:], saveSize)
				return true
			}
		}
		return false
	}

	// 获取 payload 中的偏移量
	getPayloadOffset := func(index int) uint32 {
		return binary.LittleEndian.Uint32(p.payload[index*4:])
	}

	// 修补各种写入函数
	checkAndPatch(writeSRAMSignature, thumbBranchThunk, getPayloadOffset(WriteSRAMPatched), 0x8000, "WriteSram", false)
	checkAndPatch(writeSRAM2Signature, thumbBranchThunk, getPayloadOffset(WriteSRAMPatched), 0x8000, "WriteSram 2", false)
	checkAndPatch(writeSRAMRAMSignature, armBranchThunk, getPayloadOffset(WriteSRAMPatched), 0x8000, "WriteSramFast", true)
	checkAndPatch(writeEEPROMSignature, thumbBranchThunk, getPayloadOffset(WriteEEPROMPatched), 0x2000, "SRAM-patched ProgramEepromDword", false)
	checkAndPatch(writeFlashSignature, thumbBranchThunk, getPayloadOffset(WriteFlashPatched), 0x10000, "SRAM-patched flash write function 1", false)
	checkAndPatch(writeFlash2Signature, thumbBranchThunk, getPayloadOffset(WriteFlashPatched), 0x10000, "SRAM-patched flash write function 2", false)
	checkAndPatch(writeFlash3Signature, thumbBranchThunk, getPayloadOffset(WriteFlashPatched), 0x20000, "Flash write function 3", false)

	// 特殊处理 EEPROM V111
	for offset := 0; offset <= int(p.romSize)-len(writeEEPROMV111Signature); offset += 2 {
		if bytes.Equal(p.romData[offset:offset+len(writeEEPROMV111Signature)], writeEEPROMV111Signature) {
			foundWriteLocation = true
			if mode == 0 {
				fmt.Printf("SRAM-patched EEPROM_V111 epilogue 在偏移 0x%x 处识别\n", offset)
				copy(p.romData[offset+12:], writeEEPROMV11EpiloguePatch)
				binary.LittleEndian.PutUint32(p.romData[offset+44:], 0x08000000+uint32(payloadBase)+getPayloadOffset(WriteEEPROMV111PostHook))
			}
			binary.LittleEndian.PutUint32(p.romData[payloadBase+SaveSize*4:], 0x2000)
			break
		}
	}

	if !foundWriteLocation {
		if mode == 0 {
			return errors.New("找不到要挂钩的写入函数。您确定游戏具有保存功能并已使用 GBATA 进行 SRAM 修补吗?")
		}
		fmt.Println("不确定这是什么保存类型。默认为 128KB 保存")
	}
	return nil
}

// 执行核心补丁逻辑
func (p *ROMPatcher) applyPatch(autoMode bool) error {
	// 检查是否已打补丁
	if p.isAlreadyPatched() {
		return errors.New("找到签名。ROM 已经打过补丁!")
	}

	// 修补 IRQ 处理程序
	if err := p.patchIRQHandler(); err != nil {
		return err
	}

	// 查找 payload 基地址
	payloadBase := p.findPayloadBase()
	if payloadBase < 0 {
		if p.romSize+0x80000 > maxROMSize {
			return errors.New("ROM 已达到最大大小。无法扩展。无法安装 payload")
		}
		p.romSize += 0x80000
		payloadBase = int(p.romSize) - alignment - len(p.payload)
	}

	// 复制 payload
	copy(p.romData[payloadBase:], p.payload)

	// 设置刷新模式
	mode := 1
	if autoMode {
		mode = 0
	}
	binary.LittleEndian.PutUint32(p.romData[payloadBase+FlushMode*4:], uint32(mode))

	// 修补入口点
	if p.romData[3] != 0xea {
		return errors.New("意外的入口点指令")
	}

	originalEntrypointOffset := uint32(p.romData[0]) | (uint32(p.romData[1]) << 8) | (uint32(p.romData[2]) << 16)
	originalEntrypointAddress := 0x08000000 + 8 + (originalEntrypointOffset << 2)

	binary.LittleEndian.PutUint32(p.romData[payloadBase+OriginalEntrypointAddr*4:], originalEntrypointAddress)

	newEntrypointAddress := 0x08000000 + uint32(payloadBase) + binary.LittleEndian.Uint32(p.payload[PatchedEntrypoint*4:])
	binary.LittleEndian.PutUint32(p.romData[0:], 0xea000000|((newEntrypointAddress-0x08000008)>>2))

	// 修补写入函数
	if err := p.patchWriteFunctions(payloadBase, mode); err != nil && mode == 0 {
		return err
	}

	return nil
}

// 对 ROM 字节数据执行补丁操作
func (p *ROMPatcher) PatchBytes(romData []byte, autoMode bool) ([]byte, error) {
	// 加载 ROM 数据
	if err := p.loadROMFromBytes(romData); err != nil {
		return nil, err
	}

	// 执行补丁
	if err := p.applyPatch(autoMode); err != nil {
		return nil, err
	}

	// 返回修补后的 ROM 数据
	return p.getROMBytes(), nil
}

// 执行完整的补丁操作（文件版本）
func (p *ROMPatcher) Patch(romPath, outPath string, autoMode bool) error {
	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(romPath), ".gba") {
		return errors.New("文件没有 .gba 扩展名")
	}

	// 加载 ROM
	if err := p.loadROM(romPath); err != nil {
		return err
	}

	// 执行补丁
	if err := p.applyPatch(autoMode); err != nil {
		return err
	}

	// 保存 ROM
	if err := p.saveROM(outPath); err != nil {
		return fmt.Errorf("无法打开输出文件: %w", err)
	}

	return nil
}

// 是包级别的便捷函数（字节版本）
func PatchBytes(romData []byte, autoMode bool) ([]byte, error) {
	patcher := NewROMPatcher()
	return patcher.PatchBytes(romData, autoMode)
}

// 是包级别的便捷函数（文件版本）
func Patch(romPath, outPath string, autoMode bool) error {
	patcher := NewROMPatcher()
	return patcher.Patch(romPath, outPath, autoMode)
}
