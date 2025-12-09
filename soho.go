package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

// =======================
//  SOHO / SUVI endpoints
// =======================

const (
	sohoC2JPEG = "https://soho.nascom.nasa.gov/data/realtime/c2/512/latest.jpg"
	sohoC3JPEG = "https://soho.nascom.nasa.gov/data/realtime/c3/512/latest.jpg"

	sohoC2GIF = "https://sohowww.nascom.nasa.gov/data/LATEST/current_c2.gif"
	sohoC3GIF = "https://sohowww.nascom.nasa.gov/data/LATEST/current_c3.gif"

	suvi304URL    = "https://services.swpc.noaa.gov/images/animations/suvi/primary/304/latest.png"
	suvi304DirURL = "https://services.swpc.noaa.gov/images/animations/suvi/primary/304/"
)

// =======================
//  Cache in-memory
// =======================

var imageCache = map[string]fyne.Resource{}
var animationCache = map[string][]fyne.Resource{}

// =======================
//  Widget cliccabili
// =======================

type ClickableImage struct {
	widget.BaseWidget
	img      *canvas.Image
	onTapped func()
}

func NewClickableImage(img *canvas.Image, onTapped func()) *ClickableImage {
	c := &ClickableImage{
		img:      img,
		onTapped: onTapped,
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *ClickableImage) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.img)
}

func (c *ClickableImage) Tapped(*fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}

func (c *ClickableImage) TappedSecondary(*fyne.PointEvent) {}

type ClickableObject struct {
	widget.BaseWidget
	obj      fyne.CanvasObject
	onTapped func()
}

func NewClickableObject(obj fyne.CanvasObject, onTapped func()) *ClickableObject {
	c := &ClickableObject{
		obj:      obj,
		onTapped: onTapped,
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *ClickableObject) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.obj)
}

func (c *ClickableObject) Tapped(*fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}

func (c *ClickableObject) TappedSecondary(*fyne.PointEvent) {}

// =======================
//  Funzione unica: immagini
// =======================

// loadRemoteImage scarica un'immagine da url, usa cache e aggiorna il canvas.Image.
func loadRemoteImage(url, name string, img *canvas.Image, status *widget.Label) {
	// Cache hit
	if res, ok := imageCache[url]; ok {
		img.Resource = res
		img.Refresh()
		if status != nil {
			status.SetText(name + " caricata dalla cache.")
		}
		return
	}

	if status != nil {
		status.SetText(fmt.Sprintf("Scarico immagine… (%s)", name))
	}

	go func() {
		resp, err := http.Get(url)
		if err != nil {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Errore download " + name + ":\n" + err.Error())
				}
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fyne.Do(func() {
				if status != nil {
					status.SetText(fmt.Sprintf("HTTP %d per %s", resp.StatusCode, name))
				}
			})
			return
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil || len(data) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Dati non validi per " + name)
				}
			})
			return
		}

		res := fyne.NewStaticResource(name, data)

		fyne.Do(func() {
			imageCache[url] = res
			img.Resource = res
			img.Refresh()
			if status != nil {
				status.SetText(name + " aggiornata.")
			}
		})
	}()
}

// =======================
//  Funzione unica: animazioni
// =======================

// loadRemoteAnimation:
//   - se source termina con .gif / .avi → scarica il file e ritorna []Resource con un solo elemento
//   - altrimenti considera source una directory e scarica i frame (png/jpg) in ordine.
//
// Usa cache e chiama callback(frames) quando è pronto.
func loadRemoteAnimation(source, name string, status *widget.Label, callback func([]fyne.Resource)) {
	// Cache
	if frames, ok := animationCache[source]; ok {
		if status != nil {
			status.SetText(name + " caricata dalla cache.")
		}
		if callback != nil {
			fyne.Do(func() {
				callback(frames)
			})
		}
		return
	}

	if status != nil {
		status.SetText(fmt.Sprintf("Scarico animazione… (%s)", name))
	}

	lower := strings.ToLower(source)
	if strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".avi") {
		// File singolo (GIF/AVI)
		go func() {
			resp, err := http.Get(source)
			if err != nil {
				fyne.Do(func() {
					if status != nil {
						status.SetText("Errore download " + name + ":\n" + err.Error())
					}
				})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fyne.Do(func() {
					if status != nil {
						status.SetText(fmt.Sprintf("HTTP %d per %s", resp.StatusCode, name))
					}
				})
				return
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil || len(data) == 0 {
				fyne.Do(func() {
					if status != nil {
						status.SetText("Dati non validi per " + name)
					}
				})
				return
			}

			res := fyne.NewStaticResource(name, data)
			frames := []fyne.Resource{res}

			fyne.Do(func() {
				animationCache[source] = frames
				if status != nil {
					status.SetText(name + " animazione pronta.")
				}
				if callback != nil {
					callback(frames)
				}
			})
		}()
		return
	}

	// Directory con frame
	go func() {
		resp, err := http.Get(source)
		if err != nil {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Errore HTTP su directory animazione " + name + ":\n" + err.Error())
				}
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fyne.Do(func() {
				if status != nil {
					status.SetText(fmt.Sprintf("Risposta %d per %s: %s", resp.StatusCode, name, string(body)))
				}
			})
			return
		}

		htmlBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Errore lettura listing " + name + ":\n" + err.Error())
				}
			})
			return
		}
		html := string(htmlBytes)

		re := regexp.MustCompile(`href="([^"]+\.(?:png|jpg|jpeg))"`)
		matches := re.FindAllStringSubmatch(html, -1)
		if len(matches) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Nessun frame trovato per " + name)
				}
			})
			return
		}

		seen := make(map[string]bool)
		var names []string
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			n := m[1]
			if strings.HasSuffix(n, "/") {
				continue
			}
			if seen[n] {
				continue
			}
			seen[n] = true
			names = append(names, n)
		}

		if len(names) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Nessun frame valido per " + name)
				}
			})
			return
		}

		sort.Strings(names)

		const maxFrames = 60
		if len(names) > maxFrames {
			names = names[len(names)-maxFrames:]
		}

		var frames []fyne.Resource
		for _, n := range names {
			u := n
			if !strings.HasPrefix(u, "http") {
				u = source + n
			}

			r, err := http.Get(u)
			if err != nil {
				continue
			}
			data, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil || len(data) == 0 {
				continue
			}
			frames = append(frames, fyne.NewStaticResource(n, data))
		}

		if len(frames) == 0 {
			fyne.Do(func() {
				if status != nil {
					status.SetText("Impossibile scaricare i frame per " + name)
				}
			})
			return
		}

		fyne.Do(func() {
			animationCache[source] = frames
			if status != nil {
				status.SetText(fmt.Sprintf("%s: %d frame pronti.", name, len(frames)))
			}
			if callback != nil {
				callback(frames)
			}
		})
	}()
}

// =======================
//  Download massivo SOHO (immagini statiche)
// =======================

func downloadSOHOImages(status *widget.Label) {
	status.SetText("Scaricamento immagini SOHO...")

	type item struct {
		url, name string
	}

	files := []item{
		{sohoC2JPEG, "lasco_c2_latest.jpg"},
		{sohoC3JPEG, "lasco_c3_latest.jpg"},
		{sohoC2GIF, "lasco_c2_48h.gif"},
		{sohoC3GIF, "lasco_c3_48h.gif"},
	}

	go func() {
		var errs []string

		home, err := os.UserHomeDir()
		if err != nil {
			fyne.Do(func() {
				status.SetText("Errore: impossibile determinare la home utente")
			})
			return
		}
		dir := filepath.Join(home, "AstroLair", "SOHO")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fyne.Do(func() {
				status.SetText("Errore: impossibile creare directory " + dir)
			})
			return
		}

		for _, f := range files {
			resp, err := http.Get(f.url)
			if err != nil {
				errs = append(errs, f.name)
				continue
			}
			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil || len(data) == 0 {
				errs = append(errs, f.name)
				continue
			}

			path := filepath.Join(dir, f.name)
			if err := os.WriteFile(path, data, 0o644); err != nil {
				errs = append(errs, f.name)
				continue
			}
		}

		fyne.Do(func() {
			if len(errs) > 0 {
				status.SetText("Scaricato con errori per: " + strings.Join(errs, ", "))
			} else {
				status.SetText("Immagini salvate in: " + dir)
			}
		})
	}()
}

// =======================
//  Fullscreen helpers
// =======================

func showGifFullscreen(parent fyne.CanvasObject, res fyne.Resource, title string) {
	if res == nil {
		return
	}

	fyne.Do(func() {
		canvasObj := fyne.CurrentApp().Driver().CanvasForObject(parent)
		if canvasObj == nil {
			return
		}

		gifWidget, err := xwidget.NewAnimatedGifFromResource(res)
		if err != nil {
			return
		}
		gifWidget.Start()

		var popup *widget.PopUp
		closeBtn := widget.NewButton("X", func() {
			if popup != nil {
				gifWidget.Stop()
				popup.Hide()
			}
		})
		closeBtn.Importance = widget.DangerImportance

		topPad := canvas.NewRectangle(color.Transparent)
		topPad.SetMinSize(fyne.NewSize(0, 32)) // padding per status bar

		topRow := container.NewHBox(layout.NewSpacer(), closeBtn)
		topBar := container.NewVBox(topPad, topRow)

		content := container.NewBorder(topBar, nil, nil, nil, gifWidget)

		popup = widget.NewModalPopUp(content, canvasObj)
		popup.Resize(canvasObj.Size())
		popup.Show()
	})
}

func showImageFullscreen(parent fyne.CanvasObject, res fyne.Resource, title string) {
	if res == nil {
		return
	}

	fyne.Do(func() {
		canvasObj := fyne.CurrentApp().Driver().CanvasForObject(parent)
		if canvasObj == nil {
			return
		}

		full := canvas.NewImageFromResource(res)
		full.FillMode = canvas.ImageFillContain

		var popup *widget.PopUp
		closeBtn := widget.NewButton("X", func() {
			if popup != nil {
				popup.Hide()
			}
		})
		closeBtn.Importance = widget.DangerImportance

		topPad := canvas.NewRectangle(color.Transparent)
		topPad.SetMinSize(fyne.NewSize(0, 32))

		topRow := container.NewHBox(layout.NewSpacer(), closeBtn)
		topBar := container.NewVBox(topPad, topRow)

		content := container.NewBorder(topBar, nil, nil, nil, full)

		popup = widget.NewModalPopUp(content, canvasObj)
		popup.Resize(canvasObj.Size())
		popup.Show()
	})
}

// =======================
//  Vista SOHO completa
// =======================

func buildSohoView() fyne.CanvasObject {
	status := widget.NewLabel("")

	// Immagini LASCO
	c2Img := canvas.NewImageFromResource(nil)
	c2Img.FillMode = canvas.ImageFillContain
	c2Img.SetMinSize(fyne.NewSize(0, 160))

	c3Img := canvas.NewImageFromResource(nil)
	c3Img.FillMode = canvas.ImageFillContain
	c3Img.SetMinSize(fyne.NewSize(0, 160))

	// SUVI
	suvi304Img := canvas.NewImageFromResource(nil)
	suvi304Img.FillMode = canvas.ImageFillContain
	suvi304Img.SetMinSize(fyne.NewSize(320, 320))

	// Pulsanti globali
	refreshBtn := widget.NewButton("🔄 Aggiorna immagini LASCO", func() {
		loadRemoteImage(sohoC2JPEG, "LASCO C2 (JPG)", c2Img, status)
		loadRemoteImage(sohoC3JPEG, "LASCO C3 (JPG)", c3Img, status)
	})
	downloadBtn := widget.NewButton("📥 Scarica immagini", func() {
		downloadSOHOImages(status)
	})

	buttonRow := container.NewHBox(refreshBtn, downloadBtn)

	// Colonna C2
	c2Buttons := container.NewHBox(
		widget.NewButton("JPG", func() {
			if c2Img.Resource != nil {
				showImageFullscreen(c2Img, c2Img.Resource, "LASCO C2 (JPG)")
				return
			}
			status.SetText("Immagine C2 non ancora caricata.")
		}),
		widget.NewButton("GIF", func() {
			loadRemoteAnimation(sohoC2GIF, "LASCO C2 (GIF)", status, func(frames []fyne.Resource) {
				if len(frames) == 0 {
					status.SetText("Nessun dato GIF per C2.")
					return
				}
				showGifFullscreen(c2Img, frames[0], "LASCO C2 – GIF 48h")
			})
		}),
	)

	c2Column := container.NewVBox(
		widget.NewLabel("LASCO C2"),
		c2Img,
		c2Buttons,
	)

	// Colonna C3
	c3Buttons := container.NewHBox(
		widget.NewButton("JPG", func() {
			if c3Img.Resource != nil {
				showImageFullscreen(c3Img, c3Img.Resource, "LASCO C3 (JPG)")
				return
			}
			status.SetText("Immagine C3 non ancora caricata.")
		}),
		widget.NewButton("GIF", func() {
			loadRemoteAnimation(sohoC3GIF, "LASCO C3 (GIF)", status, func(frames []fyne.Resource) {
				if len(frames) == 0 {
					status.SetText("Nessun dato GIF per C3.")
					return
				}
				showGifFullscreen(c3Img, frames[0], "LASCO C3 – GIF 48h")
			})
		}),
	)

	c3Column := container.NewVBox(
		widget.NewLabel("LASCO C3"),
		c3Img,
		c3Buttons,
	)

	imageRow := container.NewGridWithColumns(2,
		c2Column,
		c3Column,
	)

	// SUVI 304 – frame singolo + animazione
	btnLoadSuvi304 := widget.NewButton("Aggiorna SUVI 304 Å (frame singolo)", func() {
		loadRemoteImage(suvi304URL, "SUVI 304 Å", suvi304Img, status)
	})

	btnShowSuviAnim := widget.NewButton("Mostra animazione SUVI 304 Å (tutti i frame)", func() {
		app := fyne.CurrentApp()
		if app == nil {
			return
		}
		wins := app.Driver().AllWindows()
		if len(wins) == 0 {
			return
		}
		win := wins[0]

		animImg := canvas.NewImageFromResource(nil)
		animImg.FillMode = canvas.ImageFillContain
		animImg.SetMinSize(fyne.NewSize(400, 400))

		animStatus := widget.NewLabel("Scarico frames SUVI 304 Å da NOAA…")

		content := container.NewVBox(
			animImg,
			animStatus,
		)

		popup := dialog.NewCustom("Animazione SUVI 304 Å", "Chiudi", content, win)

		running := true
		popup.SetOnClosed(func() {
			running = false
		})

		popup.Show()

		loadRemoteAnimation(suvi304DirURL, "SUVI 304 Å (frames)", animStatus, func(frames []fyne.Resource) {
			if len(frames) == 0 {
				animStatus.SetText("Nessun frame disponibile.")
				return
			}

			animStatus.SetText(fmt.Sprintf("Scaricati %d frame. Avvio animazione…", len(frames)))

			go func() {
				for running {
					for _, res := range frames {
						if !running {
							break
						}
						r := res
						fyne.Do(func() {
							animImg.Resource = r
							animImg.Refresh()
						})
						time.Sleep(100 * time.Millisecond)
					}
				}
			}()
		})
	})

	suvi304Box := container.NewVBox(
		widget.NewLabelWithStyle("SUVI 304 Å – NOAA GOES-16", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		suvi304Img,
		btnLoadSuvi304,
		btnShowSuviAnim,
		status,
	)

	page := container.NewVBox(
		widget.NewLabel("SOHO – Sole in tempo quasi reale"),
		buttonRow,
		widget.NewSeparator(),
		imageRow,
		widget.NewSeparator(),
		suvi304Box,
	)

	scroll := container.NewVScroll(page)
	scroll.SetMinSize(fyne.NewSize(0, 0))
	scroll.Direction = container.ScrollVerticalOnly

	root := container.NewBorder(
		nil,
		status,
		nil,
		nil,
		scroll,
	)

	// Carica anteprime iniziali
	loadRemoteImage(sohoC2JPEG, "LASCO C2 (JPG)", c2Img, status)
	loadRemoteImage(sohoC3JPEG, "LASCO C3 (JPG)", c3Img, status)
	loadRemoteImage(suvi304URL, "SUVI 304 Å", suvi304Img, status)

	return root
}
