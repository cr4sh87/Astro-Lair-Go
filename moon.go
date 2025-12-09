package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// =======================
//  MoonPhaseCalculator
// =======================

// Durata del mese sinodico in giorni (approssimazione standard)
const synodicMonth = 29.53058867

// Data di riferimento di una Luna Nuova (2000-01-06 18:14 UTC)
var refNewMoon = time.Date(2000, 1, 6, 18, 14, 0, 0, time.UTC)

// Restituisce:
//   - phase: valore tra 0.0 e 1.0 (0 = luna nuova, 0.5 = piena, ecc.)
//   - ageDays: età della Luna in giorni dall'ultima luna nuova
func moonPhaseFraction(t time.Time) (phase float64, ageDays float64) {
	days := t.Sub(refNewMoon).Hours() / 24.0

	age := math.Mod(days, synodicMonth)
	if age < 0 {
		age += synodicMonth
	}

	phase = age / synodicMonth
	return phase, age
}

func moonPhaseName(phase float64) string {
	// normalizziamo nel range [0,1)
	f := math.Mod(phase, 1.0)
	if f < 0 {
		f += 1.0
	}

	switch {
	case isNearPhase(f, 0.0):
		return "Luna Nuova"
	case isNearPhase(f, 0.25):
		return "Primo Quarto"
	case isNearPhase(f, 0.5):
		return "Luna Piena"
	case isNearPhase(f, 0.75):
		return "Ultimo Quarto"
	case f >= 0.0 && f < 0.25:
		return "Falce Crescente"
	case f >= 0.25 && f < 0.5:
		return "Gibbosa Crescente"
	case f >= 0.5 && f < 0.75:
		return "Gibbosa Calante"
	default:
		return "Falce Calante"
	}
}

// approssimazione dell’illuminazione: 0..1
func moonIllumination(phase float64) float64 {
	// formula semplice: (1 - cos(2πphase)) / 2
	return 0.5 * (1.0 - math.Cos(2.0*math.Pi*phase))
}

func isNearPhase(value, target float64) bool {
	const tol = 0.03
	diff := math.Abs(value - target)
	return diff <= tol || diff >= 1.0-tol
}

// =======================
//  Mapping fase → sprite immagine
// =======================

// Restituisce il path dell'immagine della Luna in base alla fase (0..1).
// Aspetta di trovare i file:
// assets/moon/moon_00.jpg
// assets/moon/moon_12.jpg
// assets/moon/moon_25.jpg
// assets/moon/moon_37.jpg
// assets/moon/moon_50.jpg
// assets/moon/moon_62.jpg
// assets/moon/moon_75.jpg
// assets/moon/moon_87.jpg
func moonPhaseImageFile(phase float64) string {
	// 8 "bucket" da 0 a 7
	idx := int(math.Round(phase * 8))
	if idx == 8 {
		idx = 0
	}

	switch idx {
	case 0:
		return "assets/moon/moon_00.jpg" // Luna Nuova
	case 1:
		return "assets/moon/moon_12.jpg" // Falce crescente
	case 2:
		return "assets/moon/moon_25.jpg" // Primo Quarto
	case 3:
		return "assets/moon/moon_37.jpg" // Gibbosa crescente
	case 4:
		return "assets/moon/moon_50.jpg" // Piena
	case 5:
		return "assets/moon/moon_62.jpg" // Gibbosa calante
	case 6:
		return "assets/moon/moon_75.jpg" // Ultimo Quarto
	case 7:
		return "assets/moon/moon_87.jpg" // Falce calante
	default:
		return "assets/moon/moon_00.jpg"
	}
}

// =======================
//  Best nights per osservazione
// =======================

type BestNightEntry struct {
	Date      time.Time // user locale, ora centrale considerata
	Illum     float64   // 0..1
	PhaseName string
	Quality   string // "Eccellente", "Buona", ...
}

// qualità molto grezza basata solo su illuminazione
func qualityFromIllum(illum float64) string {
	switch {
	case illum < 0.10:
		return "Eccellente"
	case illum < 0.25:
		return "Buona"
	case illum < 0.40:
		return "Discreta"
	default:
		return "Scarsa"
	}
}

// Calcola le "migliori notti" nei prossimi `days` giorni.
// Consideriamo la Luna alle 22:00 locali di ogni giorno.
func computeBestNights(start time.Time, days int) []BestNightEntry {
	loc := start.Location()
	var nights []BestNightEntry

	for i := 0; i < days; i++ {
		d := start.AddDate(0, 0, i)
		night := time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, loc)

		phase, _ := moonPhaseFraction(night.UTC())
		illum := moonIllumination(phase)
		name := moonPhaseName(phase)

		// tieni solo notti con poca Luna (≤ 30% circa)
		if illum <= 0.30 {
			nights = append(nights, BestNightEntry{
				Date:      night,
				Illum:     illum,
				PhaseName: name,
				Quality:   qualityFromIllum(illum),
			})
		}
	}

	return nights
}

// =======================
//  Moon tab UI
// =======================

func buildMoonView() fyne.CanvasObject {
	now := time.Now()
	loc := now.Location()

	// stato corrente selezionato (data + ora)
	selected := now

	moonImg := canvas.NewImageFromResource(nil)
	moonImg.FillMode = canvas.ImageFillContain
	moonImg.SetMinSize(fyne.NewSize(160, 160)) // o quanto ti piace

	// Label che descrivono la fase per la data/ora selezionate
	nameLabel := widget.NewLabelWithStyle(
		"",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)
	phaseLabel := widget.NewLabel("")
	illumLabel := widget.NewLabel("")
	ageLabel := widget.NewLabel("")
	timeLabel := widget.NewLabel("")

	infoModel := widget.NewLabel("Modello semplice basato su mese sinodico medio.\nVa benissimo per uso visuale / planning; non è un'efemeride ad alta precisione.")

	// -----------------------
	// Funzione che aggiorna le label + immagine
	// -----------------------

	updateForSelected := func() {
		tUTC := selected.In(time.UTC)
		phase, age := moonPhaseFraction(tUTC)
		illum := moonIllumination(phase)
		name := moonPhaseName(phase)

		nameLabel.SetText(fmt.Sprintf("Fase: %s", name))
		phaseLabel.SetText(fmt.Sprintf("Frazione lunare: %.3f (0 = nuova, 0.5 = piena)", phase))
		illumLabel.SetText(fmt.Sprintf("Illuminazione: %.1f%%", illum*100.0))
		ageLabel.SetText(fmt.Sprintf("Età della Luna: %.1f giorni", age))
		timeLabel.SetText(fmt.Sprintf("Data/ora selezionata: %s", selected.In(loc).Format("2006-01-02 15:04")))

		// indice sprite 0..15 in base alla fase 0..1
		idx := int(phase * 8.0)
		if idx == 8 {
			idx = 7
		}
		res := getMoonSpriteByIndex(idx)
		if res != nil {
			moonImg.Resource = res
			moonImg.Refresh()
		}
	}

	// -----------------------
	// TextBox per data + pulsante calendario
	// -----------------------
	dateEntry := widget.NewEntry()
	dateEntry.SetPlaceHolder("YYYY-MM-DD")
	dateEntry.SetText(now.Format("2006-01-02"))
	dateEntry.Disable() // se la vuoi non editabile

	timeEntry := widget.NewEntry()
	timeEntry.SetPlaceHolder("HH:MM")
	timeEntry.SetText(now.Format("15:04"))

	// Funzione che legge dateEntry/timeEntry e aggiorna "selected"
	updateSelectedFromFields := func() {
		dtStr := dateEntry.Text
		tStr := timeEntry.Text
		d, err1 := time.Parse("2006-01-02", dtStr)
		tm, err2 := time.Parse("15:04", tStr)
		if err1 != nil || err2 != nil {
			// input non valido -> non aggiorno
			return
		}
		selected = time.Date(
			d.Year(), d.Month(), d.Day(),
			tm.Hour(), tm.Minute(), 0, 0, loc,
		)
		updateForSelected()
	}

	openCalendarBtn := widget.NewButton("📅", func() {
		showCalendarPopup(dateEntry, func(d time.Time) {
			dateEntry.SetText(d.Format("2006-01-02"))
			updateSelectedFromFields()
		})
	})

	// Pulsante "Ora attuale"
	useNowBtn := widget.NewButton("Ora attuale", func() {
		now := time.Now()
		dateEntry.SetText(now.Format("2006-01-02"))
		timeEntry.SetText(now.Format("15:04"))
		selected = now
		updateForSelected()
	})

	// Pulsante "Aggiorna"
	updateBtn := widget.NewButton("Aggiorna", func() {
		updateSelectedFromFields()
	})

	dateColumn := container.NewVBox(
		widget.NewLabel("Data"),
		dateEntry,
	)

	timeColumn := container.NewVBox(
		widget.NewLabel("Ora"),
		timeEntry,
	)

	calendarColumn := container.NewVBox(
		widget.NewLabel(" "),
		openCalendarBtn,
	)

	spacer1 := newSpacer()
	spacer2 := newSpacer()

	dateTimeRow := container.New(
		NewPercentLayout(0.15, 0.03, 0.10, 0.03, 0.05),
		dateColumn,     // 35%
		spacer1,        // 5%
		timeColumn,     // 25%
		spacer2,        // 5%
		calendarColumn, // 30%
	)

	pickersBox := container.NewVBox(
		dateTimeRow,
		container.NewHBox(useNowBtn, updateBtn),
	)

	// primo aggiornamento (label + immagine)
	updateForSelected()

	// -----------------------
	// Tabella "Prossime notti buone"
	// -----------------------

	bestNights := computeBestNights(now, 30) // prossimi 30 giorni

	headers := []string{"Data", "Ora", "Illuminazione", "Qualità"}

	table := widget.NewTable(
		func() (int, int) {
			return len(bestNights) + 1, len(headers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			lbl := obj.(*widget.Label)

			if id.Row == 0 {
				lbl.TextStyle = fyne.TextStyle{Bold: true}
				if id.Col >= 0 && id.Col < len(headers) {
					lbl.SetText(headers[id.Col])
				} else {
					lbl.SetText("")
				}
				return
			}

			lbl.TextStyle = fyne.TextStyle{}
			entryIndex := id.Row - 1
			if entryIndex < 0 || entryIndex >= len(bestNights) {
				lbl.SetText("")
				return
			}
			e := bestNights[entryIndex]

			switch id.Col {
			case 0:
				lbl.SetText(e.Date.Format("02/01/2006"))
			case 1:
				lbl.SetText(e.Date.Format("15:04"))
			case 2:
				lbl.SetText(fmt.Sprintf("%.1f%%", e.Illum*100.0))
			case 3:
				lbl.SetText(e.Quality)
			default:
				lbl.SetText("")
			}
		},
	)

	table.SetColumnWidth(0, 110)
	table.SetColumnWidth(1, 70)
	table.SetColumnWidth(2, 110)
	table.SetColumnWidth(3, 100)

	tableScroll := container.NewVScroll(table)
	tableScroll.SetMinSize(fyne.NewSize(0, 200))

	tableTitle := widget.NewLabelWithStyle(
		"Prossime notti favorevoli (22:00 locali, poca Luna)",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	topRow := container.NewVBox(
		widget.NewLabelWithStyle("Fase Lunare", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		pickersBox,
		widget.NewSeparator(),
		container.NewHBox(
			container.NewVBox(
				nameLabel,
				phaseLabel,
				illumLabel,
				ageLabel,
				timeLabel,
			),
			container.NewCenter(moonImg),
		),
		widget.NewSeparator(),
		infoModel,
	)

	topBox := topRow

	bottomBox := container.NewVBox(
		tableTitle,
		tableScroll,
	)

	content := container.NewVBox(
		topBox,
		widget.NewSeparator(),
		bottomBox,
	)

	// Rende scrollabile tutta la pagina Moon
	return container.NewVScroll(content)
}

// showCalendarPopup mostra un piccolo calendario del mese corrente
// e richiama onSelect con la data scelta.
func showCalendarPopup(anchor fyne.CanvasObject, onSelect func(time.Time)) {
	canvas := fyne.CurrentApp().Driver().CanvasForObject(anchor)
	if canvas == nil {
		return
	}

	today := time.Now()
	year, month, _ := today.Date()
	loc := today.Location()

	first := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	firstWeekday := int(first.Weekday())
	if firstWeekday == 0 {
		firstWeekday = 7
	}
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()

	grid := container.NewGridWithColumns(7)

	weekdays := []string{"Lun", "Mar", "Mer", "Gio", "Ven", "Sab", "Dom"}
	for _, w := range weekdays {
		grid.Add(widget.NewLabelWithStyle(w, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	}

	for i := 1; i < firstWeekday; i++ {
		grid.Add(widget.NewLabel(""))
	}

	var popup *widget.PopUp

	for day := 1; day <= daysInMonth; day++ {
		d := day
		btn := widget.NewButton(fmt.Sprintf("%d", d), func() {
			selectedDate := time.Date(year, month, d, 0, 0, 0, 0, loc)
			if onSelect != nil {
				onSelect(selectedDate)
			}
			if popup != nil {
				popup.Hide()
			}
		})
		grid.Add(btn)
	}

	content := container.NewVBox(grid)
	popup = widget.NewPopUp(content, canvas)

	// posizione "vicino" al campo (grezza ma funziona)
	popup.ShowAtPosition(fyne.NewPos(100, 100))
}

type PercentLayout struct {
	cols []float32 // percentuali: devono sommare ~1.0
}

func NewPercentLayout(cols ...float32) *PercentLayout {
	return &PercentLayout{cols: cols}
}

func (l *PercentLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for i, obj := range objects {
		w := size.Width * l.cols[i]
		obj.Resize(fyne.NewSize(w, size.Height))
		obj.Move(fyne.NewPos(x, 0))
		x += w
	}
}

func (l *PercentLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	// minima: la somma della minsize delle colonne
	var minW, minH float32
	for _, o := range objects {
		s := o.MinSize()
		minW += s.Width
		if s.Height > minH {
			minH = s.Height
		}
	}
	return fyne.NewSize(minW, minH)
}

func newSpacer() fyne.CanvasObject {
	// rettangolo trasparente che non disegna nulla,
	// ma può essere ridimensionato dal layout
	r := canvas.NewRectangle(color.Transparent)
	return r
}
