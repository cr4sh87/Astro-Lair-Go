package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
	"github.com/cr4sh87/astro-lair-go/services"
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
//  showResourcesDialog — Unified resource viewer (UI helper)
// =======================

// showResourcesDialog displays one or more resources in a dialog.
// - single resource: if GIF/AVI, play animation; else show static image.
// - multiple resources: cycle through frames as animation.
func showResourcesDialog(parent fyne.CanvasObject, resources []fyne.Resource, title string) {
	if len(resources) == 0 {
		return
	}

	fyne.Do(func() {
		app := fyne.CurrentApp()
		if app == nil {
			return
		}
		wins := app.Driver().AllWindows()
		if len(wins) == 0 {
			return
		}
		win := wins[0]

		// Single resource
		if len(resources) == 1 {
			res := resources[0]
			name := strings.ToLower(res.Name())
			if strings.HasSuffix(name, ".gif") || strings.HasSuffix(name, ".avi") {
				gifWidget, err := xwidget.NewAnimatedGifFromResource(res)
				if err != nil {
					// Fall back to static image if GIF decoding fails
					img := canvas.NewImageFromResource(res)
					img.FillMode = canvas.ImageFillContain
					img.SetMinSize(fyne.NewSize(400, 400))
					content := container.NewVBox(img)
					popup := dialog.NewCustom(title, "Chiudi", content, win)
					popup.Show()
					return
				}
				gifWidget.Start()

				content := container.NewVBox(gifWidget)
				popup := dialog.NewCustom(title, "Chiudi", content, win)
				popup.SetOnClosed(func() {
					gifWidget.Stop()
				})
				popup.Show()
				return
			}

			// static image
			img := canvas.NewImageFromResource(res)
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(400, 400))

			content := container.NewVBox(img)
			popup := dialog.NewCustom(title, "Chiudi", content, win)
			popup.Show()
			return
		}

		// Multiple frames -> cycle
		animImg := canvas.NewImageFromResource(nil)
		animImg.FillMode = canvas.ImageFillContain
		animImg.SetMinSize(fyne.NewSize(400, 400))

		content := container.NewVBox(animImg)
		popup := dialog.NewCustom(title, "Chiudi", content, win)

		running := true
		popup.SetOnClosed(func() {
			running = false
		})

		popup.Show()

		go func() {
			for running {
				for _, res := range resources {
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
}

// BuildSohoView — SOHO UI
// =======================

func BuildSohoView() fyne.CanvasObject {
	return buildSohoView()
}

func buildSohoView() fyne.CanvasObject {
	status := widget.NewLabel("")

	// Helper riutilizzabile per colonne immagine + animazione
	buildRemoteImageColumn := func(title, imageURL, animSource, imageName, animName string) (fyne.CanvasObject, *canvas.Image) {
		img := canvas.NewImageFromResource(nil)
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(0, 240))

		jpgBtn := widget.NewButton("JPG", func() {
			if img.Resource != nil {
				showResourcesDialog(img, []fyne.Resource{img.Resource}, imageName)
				return
			}
			status.SetText(fmt.Sprintf("Immagine %s non ancora caricata.", title))
		})

		var animBtn fyne.CanvasObject
		if animSource != "" {
			animBtnWidget := widget.NewButton("GIF/Anim", func() {
				services.LoadRemoteAnimation(animSource, animName, status, func(frames []fyne.Resource) {
					if len(frames) == 0 {
						status.SetText("Nessun frame disponibile per " + title)
						return
					}

					// Show frames (single GIF or multiple frames)
					showResourcesDialog(img, frames, animName)
				})
			})
			animBtn = animBtnWidget
		} else {
			animBtn = widget.NewLabel("")
		}

		// Carica immagine iniziale
		services.LoadRemoteImage(imageURL, imageName, img, status)

		buttons := container.NewHBox(jpgBtn, animBtn)

		col := container.NewVBox(
			widget.NewLabel(title),
			img,
			buttons,
		)

		return col, img
	}

	// Colonne riutilizzate per C2, C3 e SUVI
	c2Column, c2Img := buildRemoteImageColumn("LASCO C2", sohoC2JPEG, sohoC2GIF, "LASCO C2 (JPG)", "LASCO C2 – GIF 48h")
	c3Column, c3Img := buildRemoteImageColumn("LASCO C3", sohoC3JPEG, sohoC3GIF, "LASCO C3 (JPG)", "LASCO C3 – GIF 48h")
	imageRow := container.NewGridWithColumns(2,
		c2Column,
		c3Column,
	)

	suviColumn, _ := buildRemoteImageColumn("SUVI 304 Å – NOAA GOES-16", suvi304URL, suvi304DirURL, "SUVI 304 Å", "SUVI 304 Å (frames)")

	// Pulsanti globali
	refreshBtn := widget.NewButton("🔄 Aggiorna immagini LASCO", func() {
		services.LoadRemoteImage(sohoC2JPEG, "LASCO C2 (JPG)", c2Img, status)
		services.LoadRemoteImage(sohoC3JPEG, "LASCO C3 (JPG)", c3Img, status)
	})
	downloadBtn := widget.NewButton("📥 Scarica immagini", func() {
		services.DownloadSOHOImages(status)
	})

	buttonRow := container.NewHBox(refreshBtn, downloadBtn)

	page := container.NewVBox(
		widget.NewLabel("SOHO – Sole in tempo quasi reale"),
		buttonRow,
		widget.NewSeparator(),
		imageRow,
		widget.NewSeparator(),
		suviColumn,
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

	return root
}
