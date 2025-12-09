package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// =======================
//  DSO Catalog (JSON)
// =======================

type DsoCatalog struct {
	Version     int         `json:"version"`
	GeneratedAt string      `json:"generatedAt"`
	Objects     []DsoObject `json:"objects"`
}

type DsoObject struct {
	ID                string   `json:"id"`
	Catalog           string   `json:"catalog"`
	Code              string   `json:"code"`
	Number            *int     `json:"number"`
	NGC               *string  `json:"ngc"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Constellation     string   `json:"constellation"`
	RADeg             *float64 `json:"raDeg"`
	DecDeg            *float64 `json:"decDeg"`
	Mag               *float64 `json:"mag"`
	SurfaceBrightness *float64 `json:"surfaceBrightness"`
	ImageURL          *string  `json:"imageUrl"`
}

type TargetObject struct {
	Catalog       string
	Code          string
	Name          string
	Type          string
	Magnitude     float64
	SurfaceBright *float64
	RA            string
	Dec           string
	Constellation string
}

var targetsByCatalog = map[string][]TargetObject{
	"Messier": {
		{
			Catalog:       "Messier",
			Code:          "M31",
			Name:          "Galassia di Andromeda",
			Type:          "Galassia a spirale",
			Magnitude:     3.4,
			SurfaceBright: floatPtr(22.0),
			RA:            "00h 42m 44s",
			Dec:           "+41° 16′ 09″",
			Constellation: "Andromeda",
		},
		{
			Catalog:       "Messier",
			Code:          "M42",
			Name:          "Nebulosa di Orione",
			Type:          "Nebulosa a emissione",
			Magnitude:     4.0,
			SurfaceBright: floatPtr(21.0),
			RA:            "05h 35m 17s",
			Dec:           "−05° 23′ 28″",
			Constellation: "Orione",
		},
		{
			Catalog:       "Messier",
			Code:          "M45",
			Name:          "Pleiadi",
			Type:          "Ammasso aperto",
			Magnitude:     1.6,
			SurfaceBright: nil,
			RA:            "03h 47m 24s",
			Dec:           "+24° 07′ 00″",
			Constellation: "Toro",
		},
	},
	"NGC": {
		{
			Catalog:       "NGC",
			Code:          "NGC 7000",
			Name:          "Nebulosa Nord America",
			Type:          "Nebulosa a emissione",
			Magnitude:     4.0,
			SurfaceBright: nil,
			RA:            "20h 58m",
			Dec:           "+44° 20′",
			Constellation: "Cigno",
		},
		{
			Catalog:       "NGC",
			Code:          "NGC 253",
			Name:          "Galassia dello Scultore",
			Type:          "Galassia a spirale",
			Magnitude:     8.0,
			SurfaceBright: floatPtr(22.5),
			RA:            "00h 47m 33s",
			Dec:           "−25° 17′ 18″",
			Constellation: "Scultore",
		},
	},
}

func floatPtr(f float64) *float64 { return &f }

// Prova prima a caricare il catalogo DSO dal repo remoto.
// Se fallisce, torna al catalogo embedded (targetsByCatalog).
func buildTargetsCatalog() map[string][]TargetObject {
	url := "https://raw.githubusercontent.com/cr4sh87/Astro-Lair/refs/heads/main/catalog/dso_catalog.json"

	m, err := loadDsoTargetsFromURL(url)
	if err == nil && len(m) > 0 {
		return m
	}

	return targetsByCatalog
}

func parseDsoTargetsJSON(data []byte) (map[string][]TargetObject, error) {
	var cat DsoCatalog
	if err := json.Unmarshal(data, &cat); err != nil {
		return nil, err
	}

	out := make(map[string][]TargetObject)

	for _, o := range cat.Objects {
		if strings.TrimSpace(o.Catalog) == "" {
			continue
		}

		code := o.Code
		if code == "" && o.Number != nil {
			code = fmt.Sprintf("%d", *o.Number)
		}

		var mag float64
		if o.Mag != nil {
			mag = *o.Mag
		} else {
			mag = 0
		}

		raStr := ""
		if o.RADeg != nil {
			raStr = fmt.Sprintf("%.3f°", *o.RADeg)
		}
		decStr := ""
		if o.DecDeg != nil {
			decStr = fmt.Sprintf("%.3f°", *o.DecDeg)
		}

		t := TargetObject{
			Catalog:       o.Catalog,
			Code:          code,
			Name:          o.Name,
			Type:          o.Type,
			Magnitude:     mag,
			SurfaceBright: o.SurfaceBrightness,
			RA:            raStr,
			Dec:           decStr,
			Constellation: o.Constellation,
		}

		out[o.Catalog] = append(out[o.Catalog], t)
	}

	for k, list := range out {
		sortTargetsByNumericCode(list)
		out[k] = list
	}

	return out, nil
}

func loadDsoTargetsFromURL(url string) (map[string][]TargetObject, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseDsoTargetsJSON(data)
}

// helper per ordinare i target in base alla parte numerica del codice
func sortTargetsByNumericCode(list []TargetObject) {
	sort.Slice(list, func(i, j int) bool {
		return extractNumericCode(list[i].Code) < extractNumericCode(list[j].Code)
	})
}

func extractNumericCode(code string) int {
	var digits []rune
	for _, ch := range code {
		if ch >= '0' && ch <= '9' {
			digits = append(digits, ch)
		}
	}
	if len(digits) == 0 {
		return math.MaxInt32
	}
	v, err := strconv.Atoi(string(digits))
	if err != nil {
		return math.MaxInt32
	}
	return v
}

type TargetsView struct {
	allByCatalog    map[string][]TargetObject
	catalogNames    []string
	currentCatalog  string
	allTargets      []TargetObject
	filtered        []TargetObject
	root            fyne.CanvasObject
	list            *widget.List
	searchEntry     *widget.Entry
	detailLabel     *widget.Label
	catalogSelector *widget.Select
}

func NewTargetsView(byCatalog map[string][]TargetObject) *TargetsView {
	tv := &TargetsView{
		allByCatalog: byCatalog,
	}

	for name := range byCatalog {
		tv.catalogNames = append(tv.catalogNames, name)
	}
	sort.Strings(tv.catalogNames)

	if len(tv.catalogNames) > 0 {
		tv.currentCatalog = tv.catalogNames[0]
	}

	for k, v := range tv.allByCatalog {
		copyList := append([]TargetObject(nil), v...)
		sortTargetsByNumericCode(copyList)
		tv.allByCatalog[k] = copyList
	}

	tv.allTargets = append(tv.allTargets, tv.allByCatalog[tv.currentCatalog]...)
	tv.filtered = append(tv.filtered, tv.allTargets...)

	tv.searchEntry = widget.NewEntry()
	tv.searchEntry.SetPlaceHolder("Cerca per nome, codice o costellazione…")
	tv.searchEntry.OnChanged = func(s string) {
		tv.applyFilter(s)
	}

	tv.catalogSelector = widget.NewSelect(tv.catalogNames, func(selected string) {
		if selected == "" {
			return
		}
		tv.switchCatalog(selected)
	})
	tv.catalogSelector.PlaceHolder = "Seleziona catalogo"
	if tv.currentCatalog != "" {
		tv.catalogSelector.Selected = tv.currentCatalog
	}

	tv.detailLabel = widget.NewLabel("Seleziona un target per vedere i dettagli.")
	tv.detailLabel.Wrapping = fyne.TextWrapWord

	tv.list = widget.NewList(
		func() int {
			return len(tv.filtered)
		},
		func() fyne.CanvasObject {
			title := widget.NewLabel("Nome target")
			sub := widget.NewLabel("Catalogo • Costellazione • Magnitudine")
			sub.TextStyle = fyne.TextStyle{Italic: true}
			return container.NewVBox(title, sub)
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			if id < 0 || id >= len(tv.filtered) {
				return
			}
			t := tv.filtered[id]
			box := co.(*fyne.Container)
			title := box.Objects[0].(*widget.Label)
			sub := box.Objects[1].(*widget.Label)

			title.SetText(t.Name)
			sub.SetText(fmt.Sprintf("%s • %s • mag %.1f",
				fmt.Sprintf("%s %s", t.Catalog, t.Code),
				t.Constellation,
				t.Magnitude,
			))
		},
	)

	tv.list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(tv.filtered) {
			return
		}
		t := tv.filtered[id]
		tv.detailLabel.SetText(tv.formatDetails(t))
	}

	topControls := container.NewVBox(
		tv.catalogSelector,
		tv.searchEntry,
	)

	listCard := container.NewBorder(
		topControls,
		nil,
		nil,
		nil,
		tv.list,
	)

	detailCard := container.NewVBox(
		widget.NewLabelWithStyle("Dettagli target", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		tv.detailLabel,
	)

	detailScroll := container.NewVScroll(detailCard)
	detailScroll.SetMinSize(fyne.NewSize(220, 0))

	split := container.NewHSplit(listCard, detailScroll)
	split.Offset = 0.45

	tv.root = split
	return tv
}

func (tv *TargetsView) switchCatalog(catalog string) {
	tv.currentCatalog = catalog
	tv.allTargets = tv.allTargets[:0]
	tv.allTargets = append(tv.allTargets, tv.allByCatalog[catalog]...)
	tv.applyFilter(tv.searchEntry.Text)
}

func (tv *TargetsView) applyFilter(q string) {
	q = strings.TrimSpace(strings.ToLower(q))
	tv.filtered = tv.filtered[:0]

	if q == "" {
		tv.filtered = append(tv.filtered, tv.allTargets...)
		tv.list.Refresh()
		return
	}

	for _, t := range tv.allTargets {
		if strings.Contains(strings.ToLower(t.Name), q) ||
			strings.Contains(strings.ToLower(t.Constellation), q) ||
			strings.Contains(strings.ToLower(t.Code), q) ||
			strings.Contains(strings.ToLower(t.Catalog), q) {
			tv.filtered = append(tv.filtered, t)
		}
	}
	tv.list.Refresh()
}

func (tv *TargetsView) formatDetails(t TargetObject) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "Nome: %s\n", t.Name)
	fmt.Fprintf(sb, "Catalogo: %s %s\n", t.Catalog, t.Code)
	fmt.Fprintf(sb, "Tipo: %s\n", t.Type)
	fmt.Fprintf(sb, "Costellazione: %s\n", t.Constellation)
	fmt.Fprintf(sb, "Magnitudine: %.1f\n", t.Magnitude)
	if t.SurfaceBright != nil {
		fmt.Fprintf(sb, "Luminosità superficiale: %.2f mag/arcsec²\n", *t.SurfaceBright)
	}
	fmt.Fprintf(sb, "\nCoordinate:\n  RA:  %s\n  Dec: %s\n", t.RA, t.Dec)
	return sb.String()
}

func (tv *TargetsView) Widget() fyne.CanvasObject {
	return tv.root
}
