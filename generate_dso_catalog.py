#!/usr/bin/env python3
"""
Generatore catalogo DSO per Astro-Lair.

- Scarica OpenNGC (NGC + addendum)
- Converte in formato dso_catalog.json
- (Opzionale) fa git add/commit/push

Dipendenze:
    pip install requests
"""

import csv
import io
import json
import subprocess
from datetime import datetime
from pathlib import Path
from typing import List, Dict, Any, Optional

import requests

# =========================
# CONFIG
# =========================

# URL sorgenti OpenNGC
NGC_URL = "https://github.com/mattiaverga/OpenNGC/raw/master/database_files/NGC.csv"
ADDENDUM_URL = "https://github.com/mattiaverga/OpenNGC/raw/master/database_files/addendum.csv"

# Dove salvare il catalogo generato dentro al repo Astro-Lair
OUTPUT_PATH = Path("catalog/dso_catalog.json")

# Se True, alla fine fa git add/commit/push
AUTO_COMMIT_AND_PUSH = True

GIT_COMMIT_MESSAGE = "Update DSO catalog (generated from OpenNGC)"


# =========================
# FUNZIONI UTILI
# =========================

def fetch_csv(url: str) -> List[Dict[str, str]]:
    """Scarica un CSV da URL e lo restituisce come lista di dict."""
    print(f"[INFO] Scarico CSV da: {url}")
    resp = requests.get(url, timeout=60)
    resp.raise_for_status()
    text = resp.text

    # OpenNGC usa attualmente ';' come separatore
    reader = csv.DictReader(io.StringIO(text), delimiter=";")
    rows = list(reader)
    print(f"[INFO]   → {len(rows)} righe lette")
    return rows


def try_parse_float(value: Optional[str]) -> Optional[float]:
    """Parsing float robusto, accetta anche virgole."""
    if value is None:
        return None
    v = value.strip()
    if not v:
        return None
    v = v.replace(",", ".")
    try:
        return float(v)
    except ValueError:
        return None


def parse_angle(value: Optional[str], is_ra: bool) -> Optional[float]:
    """
    Converte una stringa di angolo in gradi.
    - Se contiene ':' o spazi -> interpreta come sessagesimale:
      * RA  = HH:MM:SS → gradi = (H + M/60 + S/3600) * 15
      * Dec = ±DD:MM:SS → gradi con segno
    - Altrimenti prova come float già in gradi.
    """
    if value is None:
        return None
    v = value.strip()
    if not v:
        return None

    # formato sessagesimale?
    if ":" in v or " " in v:
        # normalizza i separatori
        v_norm = v.replace(" ", ":")
        sign = 1.0

        if not is_ra:
            # Dec può avere segno davanti
            if v_norm[0] in "+-":
                if v_norm[0] == "-":
                    sign = -1.0
                v_norm = v_norm[1:]

        parts = v_norm.split(":")
        try:
            h_or_d = float(parts[0])
            m = float(parts[1]) if len(parts) > 1 else 0.0
            s = float(parts[2]) if len(parts) > 2 else 0.0
        except (ValueError, IndexError):
            return None

        if is_ra:
            # RA in ore → gradi
            hours = h_or_d
            deg = (hours + m / 60.0 + s / 3600.0) * 15.0
            return deg
        else:
            # Dec in gradi con segno
            deg_abs = abs(h_or_d) + m / 60.0 + s / 3600.0
            return sign * deg_abs

    # fallback: float in gradi
    return try_parse_float(v)


def build_dso_objects(rows: List[Dict[str, str]], default_catalog: str) -> List[Dict[str, Any]]:
    """
    Converte le righe del CSV OpenNGC (NGC.csv o addendum.csv)
    nel formato interno usato da Astro-Lair.
    Struttura attuale del CSV:

    Name;Type;RA;Dec;Const;MajAx;MinAx;PosAng;B-Mag;V-Mag;J-Mag;H-Mag;K-Mag;SurfBr;Hubble;Pax;Pm-RA;Pm-Dec;RadVel;Redshift;Cstar U-Mag;Cstar B-Mag;Cstar V-Mag;M;NGC;IC;Cstar Names;Identifiers;Common names;NED notes;OpenNGC notes;Sources
    """
    objects: List[Dict[str, Any]] = []

    for row in rows:
        # Nome "catalogo"
        name = (
            row.get("Name")
            or row.get("NAME")
            or row.get("name")
        )

        # Nome comune
        common = (
            row.get("Common names")
            or row.get("Common name")
            or row.get("CommonName")
            or row.get("Common")
        )

        obj_type = (
            row.get("Type")
            or row.get("TYPE")
        )

        constellation = (
            row.get("Const")
            or row.get("CONST")
            or row.get("Constellation")
        )

        # RA/Dec:
        #  - nel CSV attuale: RA/Dec in formato HH:MM:SS / ±DD:MM:SS
        #  - fallback per eventuali versioni future con RAJ2000/DEJ2000/RAdeg/DEdeg
        ra_deg = parse_angle(
            row.get("RA")
            or row.get("RAJ2000")
            or row.get("RAdeg"),
            is_ra=True,
        )
        dec_deg = parse_angle(
            row.get("Dec")
            or row.get("DEJ2000")
            or row.get("DEdeg"),
            is_ra=False,
        )

        # Se non ha coordinate, lo scartiamo
        if ra_deg is None or dec_deg is None:
            continue

        # Magnitudine: preferisci V, fallback B
        mag_v = try_parse_float(
            row.get("V-Mag")
            or row.get("Vmag")
            or row.get("m_V")
            or row.get("V_MAG")
        )
        mag_b = try_parse_float(
            row.get("B-Mag")
            or row.get("Bmag")
            or row.get("m_B")
            or row.get("B_MAG")
        )
        mag = mag_v if mag_v is not None else mag_b

        # Luminosità superficiale, se presente
        surface_brightness = try_parse_float(
            row.get("SurfBr")
            or row.get("SurfBr_V")
            or row.get("SurfaceBrightness")
        )

        # Dimensioni angolari (in arcmin, tipicamente)
        maj_ax = try_parse_float(row.get("MajAx"))
        min_ax = try_parse_float(row.get("MinAx"))

        # Messier, se presente
        messier_code = (
            row.get("M")
            or row.get("Messier")
            or ""
        )
        messier_code = messier_code.strip()

        if messier_code:
            catalog = "Messier"
            code = f"M{messier_code}"
            try:
                number = int(messier_code)
            except ValueError:
                number = None
        else:
            catalog = default_catalog
            code = (name or "").strip()
            digits = "".join(ch for ch in code if ch.isdigit())
            number = int(digits) if digits else None

        # ID interno dell'oggetto: uso code (Mxx / NGCxxxx ecc) come chiave primaria
        obj_id = code or (name or "").strip() or "UNKNOWN"

        # Alcune colonne aggiuntive (se esistono)
        ngc_designation = row.get("NGC") or None
        ic_designation = row.get("IC") or None

        dso = {
            "id": obj_id,
            "catalog": catalog,
            "code": code,
            "number": number,
            "ngc": ngc_designation,
            "ic": ic_designation,
            "name": (common or name or code or "").strip(),
            "type": obj_type or "",
            "constellation": (constellation or "").strip(),
            "ra_deg": ra_deg,
            "dec_deg": dec_deg,
            "mag": mag,
            "surface_brightness": surface_brightness,
            "size_major": maj_ax,
            "size_minor": min_ax,
            "image_url": None,  # placeholder per uso futuro
        }

        objects.append(dso)

    print(f"[INFO]   → convertiti {len(objects)} oggetti ({default_catalog})")
    return objects


def generate_catalog_json() -> Dict[str, Any]:
    """Scarica NGC + addendum, li converte e restituisce il dict JSON completo."""
    ngc_rows = fetch_csv(NGC_URL)
    add_rows = fetch_csv(ADDENDUM_URL)

    ngc_objects = build_dso_objects(ngc_rows, default_catalog="NGC/IC")
    add_objects = build_dso_objects(add_rows, default_catalog="Addendum")

    all_objects = ngc_objects + add_objects

    catalog = {
        "version": 1,
        "generated_at": datetime.utcnow().isoformat(timespec="seconds") + "Z",
        "source": "OpenNGC (NGC.csv + addendum.csv)",
        "object_count": len(all_objects),
        "objects": all_objects,
    }

    return catalog


def save_catalog_to_file(catalog: Dict[str, Any], path: Path) -> None:
    """Salva il catalogo in JSON pretty."""
    if not path.parent.exists():
        path.parent.mkdir(parents=True, exist_ok=True)

    print(f"[INFO] Salvo catalogo in: {path}")
    with path.open("w", encoding="utf-8") as f:
        json.dump(catalog, f, ensure_ascii=False, indent=2)

    print(f"[INFO] File scritto con successo ({path.stat().st_size} byte)")


def git_commit_and_push(path: Path, message: str) -> None:
    """Esegue git add/commit/push sul file indicato.

    Richiede che:
      - lo script venga eseguito dentro la repo astro-lair
      - git sia configurato (remote origin, credenziali, ecc.)
    """
    print("[INFO] Eseguo git add/commit/push...")

    def run(cmd: list[str]) -> None:
        print(f"[CMD] {' '.join(cmd)}")
        subprocess.run(cmd, check=True)

    run(["git", "add", str(path)])
    # commit può fallire se non ci sono cambi, gestiamo il caso
    try:
        run(["git", "commit", "-m", message])
    except subprocess.CalledProcessError as e:
        print("[WARN] git commit fallito (probabilmente nessuna modifica). Dettagli:")
        print(f"       {e}")
        return

    run(["git", "push"])
    print("[INFO] git push completato.")


def main() -> None:
    print("=== Astro-Lair DSO Catalog Generator ===")

    catalog = generate_catalog_json()
    save_catalog_to_file(catalog, OUTPUT_PATH)

    if AUTO_COMMIT_AND_PUSH:
        try:
            git_commit_and_push(OUTPUT_PATH, GIT_COMMIT_MESSAGE)
        except Exception as e:
            print(f"[ERROR] Problema durante git add/commit/push: {e}")
            print("        Controlla di essere nella repo corretta e con git configurato.")


if __name__ == "__main__":
    main()
