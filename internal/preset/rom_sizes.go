package preset

import "chisknife/internal/types"

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
