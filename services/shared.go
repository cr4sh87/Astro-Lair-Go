package services

import (
	"fyne.io/fyne/v2"
)

// =======================
//  Cache in-memory (shared)
// =======================

// imageCache memorizza le immagini remote scaricate
var imageCache = map[string]fyne.Resource{}

// animationCache memorizza le sequenze di animazioni (liste di frame)
var animationCache = map[string][]fyne.Resource{}

// GetImageFromCache ritorna un'immagine dalla cache se esiste
func GetImageFromCache(url string) (fyne.Resource, bool) {
	res, ok := imageCache[url]
	return res, ok
}

// SetImageInCache salva un'immagine nella cache
func SetImageInCache(url string, res fyne.Resource) {
	imageCache[url] = res
}

// GetAnimationFromCache ritorna un'animazione dalla cache se esiste
func GetAnimationFromCache(source string) ([]fyne.Resource, bool) {
	res, ok := animationCache[source]
	return res, ok
}

// SetAnimationInCache salva un'animazione nella cache
func SetAnimationInCache(source string, frames []fyne.Resource) {
	animationCache[source] = frames
}

// ClearCache svuota tutta la cache
func ClearCache() {
	imageCache = make(map[string]fyne.Resource)
	animationCache = make(map[string][]fyne.Resource)
}
