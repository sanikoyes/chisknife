// Package types 定义了应用程序使用的核心数据类型
package types

// 定义了 ROM 大小选项
type RomSize struct {
	Desc string // 大小描述（如 "4 MB"）
	Size int    // 实际字节大小
}

// ROM 大小选项的集合
type RomSizes []RomSize

// 提取所有 ROM 大小的描述列表
// 用于在下拉菜单中显示
func (t RomSizes) Descs() []string {
	descs := make([]string, len(t))
	for i, v := range t {
		descs[i] = v.Desc
	}
	return descs
}
