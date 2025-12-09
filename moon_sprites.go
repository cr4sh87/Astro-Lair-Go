package main

import (
	"embed"
	"fmt"

	"fyne.io/fyne/v2"
)

//go:embed assets/moon/*.jpg
var moonSpriteFS embed.FS

var moonSpriteCache map[string]fyne.Resource

func initMoonSprites() {
	if moonSpriteCache != nil {
		return
	}

	moonSpriteCache = make(map[string]fyne.Resource)

	// nomi dei file da caricare (adatta se usi altri nomi)
	for i := 0; i < 16; i++ {
		name := fmt.Sprintf("moon_%02d.jpg", i)
		path := "assets/moon/" + name

		data, err := moonSpriteFS.ReadFile(path)
		if err != nil {
			// se manca qualche file, lo saltiamo
			continue
		}
		moonSpriteCache[name] = fyne.NewStaticResource(name, data)
	}
}

// Restituisce lo sprite in base all'indice 0..15
func getMoonSpriteByIndex(idx int) fyne.Resource {
	initMoonSprites() // lazy init

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

	// fallback: primo sprite
	if res, ok := moonSpriteCache["moon_00.jpg"]; ok {
		return res
	}
	return nil
}
