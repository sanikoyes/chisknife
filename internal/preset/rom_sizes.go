// Package preset 提供预定义的卡带类型和 ROM 大小配置
package preset

import "chisknife/internal/types"

// 定义了支持的最小 ROM 大小选项
// 用户可以选择不同的 ROM 容量来构建卡带
var RomSizes = types.RomSizes{
	{
		Desc: "4 MB",
		Size: 4 * 1024 * 1024,
	},
	{
		Desc: "512 KB",
		Size: 512 * 1024,
	},
	{
		Desc: "8 MB",
		Size: 8 * 1024 * 1024,
	},
}
