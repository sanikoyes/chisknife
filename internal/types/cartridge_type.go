// Package types 定义了应用程序使用的核心数据类型
package types

// 定义了卡带的硬件规格
type CartridgeType struct {
	Name       string // 卡带芯片型号名称
	FlashSize  int    // 闪存总容量（字节）
	SectorSize int    // 扇区大小（字节）
	BlockSize  int    // 块大小（字节）
}

// 卡带类型的集合
type CartridgeTypes []CartridgeType

// 提取所有卡带类型的名称列表
// 用于在下拉菜单中显示
func (t CartridgeTypes) Names() []string {
	names := make([]string, len(t))
	for i, v := range t {
		names[i] = v.Name
	}
	return names
}
