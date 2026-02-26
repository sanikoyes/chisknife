package asset

import (
	embed "embed"
)

//go:embed bg.png
var Background []byte

//go:embed translations/*.json
var TranslationsFS embed.FS
