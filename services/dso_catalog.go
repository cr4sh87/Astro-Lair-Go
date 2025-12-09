package services

import (
	"io"
	"log"
	"net/http"
	"os"
)

// DSOCatalogURL è l'URL pubblico del catalogo DSO.
const DSOCatalogURL = "https://raw.githubusercontent.com/cr4sh87/astro-lair-go/main/catalog/dso_catalog.json"

// DSOCatalogLocalPath è il path locale dove viene salvato il catalogo.
const DSOCatalogLocalPath = "catalog/dso_catalog.json"

// UpdateDSOCatalogFromGitHub scarica l'ultima versione di dso_catalog.json dal repo GitHub
// e la salva nella cartella locale "catalog", così buildTargetsCatalog() può leggerla.
func UpdateDSOCatalogFromGitHub() {
	log.Println("[DSO] Aggiorno dso_catalog.json da GitHub...")

	resp, err := http.Get(DSOCatalogURL)
	if err != nil {
		log.Printf("[DSO] Errore nel download: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[DSO] HTTP status non valido: %s\n", resp.Status)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[DSO] Errore lettura risposta: %v\n", err)
		return
	}

	if err := os.MkdirAll("catalog", 0o755); err != nil {
		log.Printf("[DSO] Impossibile creare cartella catalog/: %v\n", err)
		return
	}

	if err := os.WriteFile(DSOCatalogLocalPath, data, 0o644); err != nil {
		log.Printf("[DSO] Errore nel salvataggio di %s: %v\n", DSOCatalogLocalPath, err)
		return
	}

	log.Printf("[DSO] Catalogo aggiornato (%d byte) → %s\n", len(data), DSOCatalogLocalPath)
}
