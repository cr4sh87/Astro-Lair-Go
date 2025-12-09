package models

// =======================
//  DSO Catalog (JSON)
// =======================

type DsoCatalog struct {
	Version     int         `json:"version"`
	GeneratedAt string      `json:"generated_at"`
	Source      string      `json:"source"`
	ObjectCount int         `json:"object_count"`
	Objects     []DsoObject `json:"objects"`
}

type DsoObject struct {
	ID                string   `json:"id"`
	Catalog           string   `json:"catalog"`
	Code              string   `json:"code"`
	Number            *int     `json:"number"`
	NGC               *string  `json:"ngc"`
	IC                *string  `json:"ic"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Constellation     string   `json:"constellation"`
	RADeg             *float64 `json:"ra_deg"`
	DecDeg            *float64 `json:"dec_deg"`
	Mag               *float64 `json:"mag"`
	SurfaceBrightness *float64 `json:"surface_brightness"`
	SizeMajor         *float64 `json:"size_major"`
	SizeMinor         *float64 `json:"size_minor"`
	ImageURL          *string  `json:"image_url"`
}
