# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Überblick

`github.com/nakami-lounge-GmbH/tools` ist eine reine Library (kein `main`-Package, kein Binary), die als
gemeinsame Werkzeugsammlung in anderen nakami-Projekten eingebunden wird. Änderungen an exportierten
Signaturen brechen daher stillschweigend die Konsumenten-Repos — Signaturen möglichst additiv erweitern.

## Kommandos

```bash
go build ./...                              # muss sauber durchlaufen
go vet ./...
go test ./...                               # läuft sauber durch, inklusive vet
go test ./importer/ -run TestNewExcelPageImporter -v   # einzelner Test
go generate ./importer/excel_cols/          # stringer für ExcelPos neu erzeugen
go install golang.org/x/tools/cmd/stringer@latest      # Voraussetzung für go generate
go get -u ./...                             # Dependency-Upgrade (Repo-Historie besteht großteils daraus)
```

Zum Testen:

- `go build ./...`, `go vet ./...`, `gofmt -l .` und `go test ./...` laufen sauber durch. Der früher
  nötige Umweg über `-vet=off` ist weg: der non-constant format string in `helpers/file.go` ist
  korrigiert.
- Die `importer`-Tests brauchen keine Lizenz und keine Env-Variablen mehr; die Lizenz-Dependencies
  unidoc/unioffice sind mit dem Wechsel auf excelize entfallen. Fixture: `importer/test1.xlsx`,
  Sheet „Daten“ (Header in Zeile 3, Daten in Zeile 4 und 5, Spalten B, C, E, F, H).

## Architektur

### importer — tag-gesteuerter generischer Excel/CSV-Import

Kern ist `Importer[C any]` (`importer/main.go`): aus dem Typparameter `C` wird per Reflection eine
`map[string]*field` aufgebaut, die Struct-Tags auf Spaltenpositionen abbildet. Drei Betriebsmodi,
unterschieden durch `ImportType` und jeweils einen eigenen Tag — pro Modus existiert ein Konstruktor in
einer eigenen Datei, der die Zeilen durchläuft und `Data []C` füllt:

| Modus | Konstruktor | Tag | Bedeutung |
|---|---|---|---|
| `ExcelLine` | `NewExcelLineImporter` (`excel_line.go`) | `header:"Anzahl Kinder"` | Spalte wird über die Header-Zeile aufgelöst (case-insensitive) |
| `ExcelLineCol` | `NewExcelLineColImporter` (`excel_line_col.go`) | `col:"B"` | Feste Spaltenbuchstaben, Header-Zeile nur zur Spaltenzählung |
| `ExcelPage` | `NewExcelPageImporter` (`excel_page.go`) | `excel_pos:"B4"` | Einzelner Datensatz aus festen Zellkoordinaten |

Zusätzlich: `format:"02.01.2006"` ist bei `time.Time`/`null.Time` im CSV-Pfad **Pflicht**, sonst
Validierungsfehler.

**Fehlerbehandlung ist der wichtigste Vertrag dieses Packages:** Nahezu alle Fehler landen im
mitgegebenen `*ErrorList`, nicht im zurückgegebenen `error`. Konstruktoren geben in vielen Fällen
`(importer, nil)` zurück, obwohl der Import fehlgeschlagen ist. Aufrufer müssen immer
`el.HasAny()` / `el.HasErrors()` prüfen, nicht nur `err`. `ErrorList` (`error.go`) trennt Errors und
Warnings, hat ein eigenes `MarshalJSON` mit `has_any`/`has_errors`/`has_warnings` für API-Antworten und
`AddValidation(line, column, header, err)` für zeilenbezogene Importfehler.

**Excel-Zugriff liegt vollständig in `importer/excel.go`** — die einzige Datei, die
`github.com/xuri/excelize/v2` importiert. Sie stellt `Sheet`, `Row`, `Cell` und `CellType` bereit; der
restliche Importer arbeitet nur gegen diese Typen. `openSheet` liest das Blatt in einem Zug ein, einmal
mit den Rohwerten (`Cell.Value`, Basis aller typisierten Konvertierungen) und einmal mit den von Excel
angezeigten Werten (`Cell.String()`, für Text- und Bool-Felder). `Sheet.Rows` ist von `0` an
zeilennummern-treu indiziert, leere Zeilen sind als Zeile ohne Zellen enthalten — darauf baut die
`OffsetRow`-Logik auf. Leere Zellen am Zeilenende fehlen in `Row.Cells`, deshalb prüfen die Getter
immer gegen `len(row.Cells)`. `Cell.Type()` fragt den Typ erst bei Bedarf ab: Zahlen tragen in der
Datei keinen eigenen Typ und kommen als `CellTypeUnset` zurück, sie werden deshalb über ihren Wert
klassifiziert. Datums- und Zeitwerte sind in Excel Zahlen, also `CellTypeNumeric`; `Cell.GetTime()`
rechnet sie mit dem Epochen-Flag der Arbeitsmappe (`date1904`) um. **Wer ein `Sheet` über `GetSheet()`
holt, muss es `Close()`en.**

**Zwei getrennte Konvertierungspfade** — wer einen neuen Feldtyp unterstützen will, muss beide anfassen:

- `getCellValue` — typisierte `*Cell`-Werte, von allen drei Excel-Modi genutzt. Gibt zusätzlich zum
  Fehler ein `bool` zurück, das sagt, ob überhaupt ein Wert zugewiesen wurde; daran hängt das
  `Valid`-Flag der Nullables.
- `setLineValue` — `[]string`-Werte, für den CSV-Import; macht eigenes Zahlen-Cleanup (Punkte
  entfernen, `,` → `.`, `%`/Whitespace trimmen) und verlangt den `format`-Tag. `GetLineValues`
  ruft es pro Feld auf und hängt den Fehler als Validierungsfehler an die `ErrorList`.

Unterstützt werden die Go-Basistypen und **Nullables, die über ihre Form erkannt werden** statt über
ihren Typ: jedes Struct `{Wert T; Valid bool}` wird über sein Wertfeld gefüllt und danach als gesetzt
markiert (`importer/nullable.go`). Damit funktionieren die `null.*`-Typen aus
`github.com/aarondl/null/v8`, die unsere älteren Projekte benutzen, und die `sql.Null*`-Typen der
Standardbibliothek, **ohne dass die Library eine Abhängigkeit dazu hat**. Wichtig beim Ändern:
`Valid` darf nur gesetzt werden, wenn wirklich ein Wert zugewiesen wurde. Eine Textzelle in einem
`null.Time` bleibt deshalb ungesetzt, statt auf die Null-Zeit mit `Valid: true` zu laufen.

Bool-Parsing ist deutsch-lastig und tolerant: `wahr`, `true`, `ja`, `yes`, `1`, `x` → true, alles
andere → false (`isTrueString`).

Validierung läuft über `go-playground/validator/v10` nach dem Füllen des Structs
(`Importer.ValidateStruct`), inklusive dreier eigener Regeln in `validators.go`: `eitherrequired`,
`withrequired` und `one_of_str` (Parameter-Trennzeichen ist `#-#`, da `,` im Tag belegt ist).
Achtung: `validate` ist eine Package-Level-Variable, die in `newImorter` neu gesetzt wird — paralleles
Erzeugen von Importern ist nicht thread-safe.

### importer/excel_cols

`ExcelPos` ist ein Enum A…DG mit stringer-generierter Datei `excelpos_string.go` (nicht editieren).
`GetPosFromString` berechnet die Position aus dem Spaltennamen (Basis 26) und deckt damit alle Namen
von `A` bis `XFD` ab, auch solche jenseits der `ExcelPos`-Konstanten; Klein- und Großschreibung sind
gleichwertig, ungültige Namen liefern `-1`. Vorher war es ein `switch`, in dem `CJ`–`CU` und `DC`
fehlten und `-1` lieferten — im `ExcelLineCol`-Modus führte das zu `row.Cells[-1]` und damit zum
Panic.

### Gotcha: der Tag ist der Map-Key

`ii.fields` ist eine `map[string]*field`, geschlüsselt über den **kleingeschriebenen Tag-Wert**. Zwei
Struct-Felder mit demselben Tag (`excel_pos:"B4"` zweimal, derselbe `header:`-Name zweimal) kollidieren
deshalb stillschweigend: das später deklarierte Feld überschreibt das frühere, und das erste bleibt
beim Import leer. Kein Fehler, keine Warnung.

### Gotchas bei der Sheet-/Zeilenauswahl (`Config`)

- `SheetNumber` ist trotz Kommentar „0-indexed“ **1-basiert**; `0` bedeutet „nicht gesetzt“. Ist
  weder `SheetName` noch `SheetNumber` gesetzt, liefert `GetSheet()` einen Fehler (früher `(nil, nil)`
  und damit einen nil-Dereference beim nächsten Zugriff).
- `OffsetRow` ist die **1-basierte Nummer der Header-Zeile**; Datenzeilen sind alle mit Index
  `>= OffsetRow`. `0` bedeutet Header in der ersten Zeile.
- `LineCountToRead = 0` heißt „alle Zeilen“. `EmptyValueStrings` definiert Werte, die wie leer behandelt
  werden (Feld bleibt Zero-Value).

### Weitere Packages

- `uploader` — Validierung von Datei-Uploads: Größenlimit in MB und MIME-Typ-Prüfung per
  `gabriel-vasile/mimetype` (Content-Sniffing, nicht Dateiendung) gegen eine `AllowedFileTypes`-Liste
  (`FileTypePDF`, `FileTypePNG`, `FileTypeJPG`, `FileTypeExcel`, `FileTypeCSV`). Fehler sind
  `*ErrUploadMessage` mit **deutschen, endnutzertauglichen Texten** — diese Meldungen gehen bis ins UI,
  deshalb bei Änderungen Wortlaut und Sprache beibehalten. `GetUploadFile` arbeitet auf
  `*multipart.FileHeader`, `CheckFileTypeAndSize` auf bereits gelesenen Bytes.
- `build` — `GetBuildInfo()` liefert Laufzeit-/Build-Metadaten für Health- bzw. Info-Endpoints:
  Goroutinen, Speicher in MB, Go-Version, `vcs.revision` aus `debug.ReadBuildInfo()` und als Build-Datum
  die mtime des eigenen Executables.
- `helpers` — kleine generische Utilities (Slices, Encoding-Erkennung via `x/net/html/charset`,
  Verzeichnis-/Datei-Checks, `Round2`, `RandomString`, Tages-Zeitgrenzen `Time0`/`Time24`/`Time4` in UTC).
- `time_tracker` — manuelles Schritt-Profiling (`TrackStep`/`TrackFromStart`) mit Ausgabe auf stdout,
  Einheit über `TimeTrackesUnit*`-Konstanten, per `IsDisabled` abschaltbar.
- `scripts/del_go.sh` — Deployment-Helfer auf dem Server: listet alte Verzeichnisse unter `./versions`
  außer dem aktuell verlinkten. Nutzt bewusst nur `find -print`, löscht also nicht selbst.

## Konventionen

- Slice-Helfer kommen in der Library meist aus `golang.org/x/exp/slices`, nicht aus dem stdlib-`slices`
  (Ausnahme: `build/build_info.go`). In bestehenden Dateien dem jeweiligen Import folgen.
- Generics werden aktiv genutzt (`Importer[C any]`, `helpers.CompareSlices[T comparable]`).
- Reflection-Code arbeitet durchgehend mit `*reflect.Value`-Rückgaben, die der Aufrufer per
  `mV.Interface().(C)` casten muss.
