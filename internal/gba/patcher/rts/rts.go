// Package rts 提供 GBA ROM 的 RTS (Real-Time Save) 补丁功能
// 实现实时存档/读档功能
//
// License / 许可声明
// 未经授权，禁止用于商业行为。使用该代码的衍生项目需要保持开源，并且需要指明该项目的原始仓库地址（https://github.com/ArcheyChen/GBA-RTS-PATCH）。
// 代码中的 "Ausar'S-RTSFILE." 和 "<3 from Maniac" 等识别用字符串不应修改，而应当原样保留。
//
// Commercial use is prohibited without authorization. Any derivative project using this code must remain open source
// and clearly indicate the original repository address (https://github.com/ArcheyChen/GBA-RTS-PATCH).
// Identification strings in the code such as "Ausar'S-RTSFILE." and "<3 from Maniac" must not be altered and should be preserved as is.
package rts

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"chisknife/asset"
)

const (
	maxROMSize = 0x02000000 // 32MB
	rtsSize    = 448 * 1024 // 448KB
	alignment  = 0x40000    // 256KB
)

var (
	signature = []byte("<3 from Maniac")

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
	writeEEPROMV111Signature = []byte{
		0x0A, 0x88, 0x80, 0x21, 0x09, 0x06, 0x0A, 0x43, 0x02, 0x60, 0x07, 0x48, 0x00, 0x47, 0x00, 0x00,
	}

	oldIRQAddr = []byte{0xfc, 0x7f, 0x00, 0x03}
	newIRQAddr = []byte{0xf4, 0x7f, 0x00, 0x03}
)

// RTS payload 头部结构
type PayloadHeader struct {
	OriginalEntrypoint    uint32
	CtrlFlag              uint32
	RtsSize               uint32
	SaveSize              uint32
	WbufSize              uint32
	PatchedEntrypointAddr uint32
}

// RTS 补丁器
type RTSPatcher struct {
	romData []byte
	romSize uint32
	payload []byte
}

// 创建新的 RTS 补丁器
func NewRTSPatcher() *RTSPatcher {
	return &RTSPatcher{
		payload: asset.PayloadRts,
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
func (p *RTSPatcher) loadROMFromBytes(data []byte) error {
	p.romSize = uint32(len(data))
	if p.romSize > maxROMSize {
		return errors.New("ROM 太大 - 不是 GBA ROM?")
	}

	// 检查对齐并填充
	if p.romSize&0x3ffff != 0 {
		p.romSize &= ^uint32(0x3ffff)
		p.romSize += alignment
	}

	// 创建缓冲区并填充 0xFF
	p.romData = make([]byte, p.romSize)
	for i := range p.romData {
		p.romData[i] = 0xFF
	}
	copy(p.romData, data)

	return nil
}

// 返回修补后的 ROM 字节数据
func (p *RTSPatcher) getROMBytes() []byte {
	result := make([]byte, p.romSize)
	copy(result, p.romData[:p.romSize])
	return result
}

// 检查是否已经打过补丁
func (p *RTSPatcher) isAlreadyPatched() bool {
	return findPattern(p.romData, signature, 4) != -1
}

// 修补 IRQ 引用
func (p *RTSPatcher) patchIRQReferences() (int, error) {
	foundCount := 0
	for i := uint32(0); i < p.romSize; i += 4 {
		if bytes.Equal(p.romData[i:i+4], oldIRQAddr) {
			foundCount++
			copy(p.romData[i:], newIRQAddr)
		}
	}

	if foundCount == 0 {
		return 0, errors.New("找不到 IRQ 处理程序的任何引用。ROM 是否已经打过补丁?")
	}
	return foundCount, nil
}

// 检测存档类型
func (p *RTSPatcher) detectSaveType() (uint32, string) {
	signatures := []struct {
		sig      []byte
		saveSize uint32
		saveType string
	}{
		{writeSRAMSignature, 0x8000, "SRAM (32KB)"},
		{writeSRAM2Signature, 0x8000, "SRAM (32KB)"},
		{writeSRAMRAMSignature, 0x8000, "SRAM (32KB)"},
		{writeEEPROMSignature, 0x2000, "EEPROM (8KB)"},
		{writeEEPROMV111Signature, 0x2000, "EEPROM (8KB)"},
		{writeFlashSignature, 0x10000, "Flash (64KB)"},
		{writeFlash2Signature, 0x10000, "Flash (64KB)"},
		{writeFlash3Signature, 0x20000, "Flash (128KB)"},
	}

	for _, s := range signatures {
		if pos := findPattern(p.romData, s.sig, 2); pos != -1 {
			return s.saveSize, s.saveType
		}
	}

	return 0x20000, "Default (128KB)"
}

// 查找 payload 安装位置
func (p *RTSPatcher) findPayloadLocation(reservedSpace uint32) int {
	requiredSpace := reservedSpace + uint32(len(p.payload))

	for payloadBase := int(p.romSize) - int(requiredSpace) - alignment; payloadBase >= 0; payloadBase -= alignment {
		region := p.romData[payloadBase : payloadBase+int(requiredSpace)]
		isAllZeroes := true
		isAllOnes := true

		for _, b := range region {
			if b != 0 {
				isAllZeroes = false
			}
			if b != 0xFF {
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

// 解析 ARM 分支指令
func parseARMBranchInstruction(instruction []byte) uint32 {
	if len(instruction) != 4 {
		return 0
	}

	inst := binary.LittleEndian.Uint32(instruction)
	if (inst & 0xFF000000) != 0xEA000000 {
		return 0
	}

	offset := inst & 0x00FFFFFF
	if offset&0x00800000 != 0 {
		offset |= 0xFF000000
	}

	targetAddress := 0x08000000 + 8 + (offset << 2)
	return targetAddress
}

// 创建 ARM 分支指令
func createARMBranchInstruction(targetAddress uint32) ([]byte, error) {
	offset := (targetAddress - 0x08000000 - 8) >> 2

	if offset > 0x00FFFFFF {
		return nil, errors.New("分支目标地址超出范围")
	}

	offset &= 0x00FFFFFF
	instruction := 0xEA000000 | offset

	result := make([]byte, 4)
	binary.LittleEndian.PutUint32(result, instruction)
	return result, nil
}

// 执行核心补丁逻辑
func (p *RTSPatcher) applyPatch(wbufSize, sectorSize uint32, rtsData []byte) error {
	// 检查是否已打补丁
	if p.isAlreadyPatched() {
		return errors.New("找到签名。ROM 已经打过补丁!")
	}

	// 修补 IRQ 引用
	if _, err := p.patchIRQReferences(); err != nil {
		return err
	}

	// 检测存档类型
	detectedSaveSize, _ := p.detectSaveType()

	// 计算保留空间
	reservedSpace := uint32(0x70000) // 448KB
	reservedSpace += detectedSaveSize
	if reservedSpace%sectorSize != 0 {
		reservedSpace = reservedSpace - (reservedSpace % sectorSize) + sectorSize
	}

	// 查找 payload 位置
	payloadBase := p.findPayloadLocation(reservedSpace)
	if payloadBase == -1 {
		if p.romSize+reservedSpace > maxROMSize {
			return errors.New("ROM 已达到最大大小。无法扩展。无法安装 payload")
		}
		// 扩展 ROM
		newSize := p.romSize + reservedSpace
		newData := make([]byte, newSize)
		for i := range newData {
			newData[i] = 0xFF
		}
		copy(newData, p.romData[:p.romSize])
		p.romData = newData
		p.romSize = newSize
		payloadBase = int(p.romSize) - int(reservedSpace) - len(p.payload)
	}

	// 安装 payload
	copy(p.romData[payloadBase:], p.payload)

	// 更新 payload 头部
	header := PayloadHeader{}
	if len(p.payload) >= 24 {
		header.OriginalEntrypoint = binary.LittleEndian.Uint32(p.payload[0:4])
		header.CtrlFlag = binary.LittleEndian.Uint32(p.payload[4:8])
		header.RtsSize = binary.LittleEndian.Uint32(p.payload[8:12])
		header.SaveSize = binary.LittleEndian.Uint32(p.payload[12:16])
		header.WbufSize = binary.LittleEndian.Uint32(p.payload[16:20])
		header.PatchedEntrypointAddr = binary.LittleEndian.Uint32(p.payload[20:24])
	}

	header.RtsSize = reservedSpace
	header.SaveSize = detectedSaveSize
	header.WbufSize = wbufSize

	// 写入更新后的头部
	binary.LittleEndian.PutUint32(p.romData[payloadBase+0:], header.OriginalEntrypoint)
	binary.LittleEndian.PutUint32(p.romData[payloadBase+4:], header.CtrlFlag)
	binary.LittleEndian.PutUint32(p.romData[payloadBase+8:], header.RtsSize)
	binary.LittleEndian.PutUint32(p.romData[payloadBase+12:], header.SaveSize)
	binary.LittleEndian.PutUint32(p.romData[payloadBase+16:], header.WbufSize)
	binary.LittleEndian.PutUint32(p.romData[payloadBase+20:], header.PatchedEntrypointAddr)

	// 嵌入 RTS 文件（如果提供）
	sramSaveBase := payloadBase + len(p.payload)
	if rtsData != nil {
		if len(rtsData) != rtsSize {
			return fmt.Errorf("RTS 文件大小必须是 448KB (458752 字节)，但得到 %d 字节", len(rtsData))
		}
		copy(p.romData[sramSaveBase:], rtsData)
	}

	// 修补入口点
	if p.romData[3] != 0xEA {
		return errors.New("意外的入口点指令")
	}

	originalEntrypointAddress := parseARMBranchInstruction(p.romData[0:4])
	binary.LittleEndian.PutUint32(p.romData[payloadBase+0:], originalEntrypointAddress)

	newEntrypointAddress := 0x08000000 + uint32(payloadBase) + header.PatchedEntrypointAddr
	newBranchInstruction, err := createARMBranchInstruction(newEntrypointAddress)
	if err != nil {
		return err
	}
	copy(p.romData[0:4], newBranchInstruction)

	return nil
}

// 对 ROM 字节数据执行 RTS 补丁操作
func (p *RTSPatcher) PatchBytes(romData []byte, wbufSize, sectorSize uint32, rtsData []byte) ([]byte, error) {
	// 参数验证
	if wbufSize > 0xFFF {
		return nil, fmt.Errorf("无效的写缓冲区大小: %d (必须是 0-4095)", wbufSize)
	}
	if sectorSize < 0x10000 || sectorSize > 0x40000 {
		return nil, fmt.Errorf("无效的扇区大小: 0x%X (必须是 0x10000-0x40000)", sectorSize)
	}

	// 加载 ROM 数据
	if err := p.loadROMFromBytes(romData); err != nil {
		return nil, err
	}

	// 执行补丁
	if err := p.applyPatch(wbufSize, sectorSize, rtsData); err != nil {
		return nil, err
	}

	// 返回修补后的 ROM 数据
	return p.getROMBytes(), nil
}

// 是包级别的便捷函数
func PatchBytes(romData []byte, wbufSize, sectorSize uint32, rtsData []byte) ([]byte, error) {
	patcher := NewRTSPatcher()
	return patcher.PatchBytes(romData, wbufSize, sectorSize, rtsData)
}
