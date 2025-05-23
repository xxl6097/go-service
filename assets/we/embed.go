package we

import (
	"embed"
	"github.com/xxl6097/go-service/assets"
)

//go:embed static/*
var content embed.FS

func init() {
	assets.Register(content)
}
