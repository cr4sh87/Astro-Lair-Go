package services

import (
	"embed"
	"fmt"

	"fyne.io/fyne/v2"
)

// Sprite della Luna (embedded nel binary)
//
//go:embed assets/moon/*.jpg
var moonSpriteFS embed.FS

var moonSpriteCache map[string]fyne.Resource

func initMoonSprites() {
	if moonSpriteCache != nil {
		return
	}

	moonSpriteCache = make(map[string]fyne.Resource)

	for i := 0; i < 16; i++ {
		name := fmt.Sprintf("moon_%02d.jpg", i)
		path := "assets/moon/" + name

		data, err := moonSpriteFS.ReadFile(path)
		if err != nil {
			continue
		}
		moonSpriteCache[name] = fyne.NewStaticResource(name, data)
	}
}

// GetMoonSpriteByIndex restituisce lo sprite in base all'indice 0..15.
func GetMoonSpriteByIndex(idx int) fyne.Resource {
	initMoonSprites()

	if idx < 0 {
		idx = 0
	}
	if idx > 15 {
		idx = 15
	}

	name := fmt.Sprintf("moon_%02d.jpg", idx)
	if res, ok := moonSpriteCache[name]; ok {
		return res
	}

	if res, ok := moonSpriteCache["moon_00.jpg"]; ok {
		return res
	}
	return nil
}
