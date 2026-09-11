package ods

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"
)

const mimetype = "application/vnd.oasis.opendocument.spreadsheet"

// ODSWriter creates ODS files from scratch.
type ODSWriter struct {
	Sheets       []SheetDef
	FreezeConfig map[string]FreezePane // sheet name -> freeze settings
}

// SheetDef defines a sheet to be written.
type SheetDef struct {
	Name     string
	Hidden   bool
	Columns  []ColDef
	Header   []string
	Rows     [][]Cell
	Validations []Validation
}

// ColDef defines a column width in cm.
type ColDef struct {
	Width float64 // cm
}

// Cell represents a single cell.
type Cell struct {
	Type      string // "text", "float", "date", "formula", "empty"
	Value     string
	Style     string
	Formula   string
	Validation string
	Repeated  int    // for empty cells
	FloatVal  float64
	DateVal   string // YYYY-MM-DD
}

// Validation defines a content validation.
type Validation struct {
	Name        string
	Condition   string
	BaseCell    string
	DisplayList string // "sorted", "unsorted", "none"
	ErrorMsg    string
	HelpMsg     string
	AllowEmpty  bool
}

// FreezePane defines freeze pane settings for a sheet.
type FreezePane struct {
	Rows    int
	Columns int
}

// WriteODS writes an ODS file to disk.
func WriteODS(path string, writer *ODSWriter) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// 1. mimetype (first entry, no compression)
	mw, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		return fmt.Errorf("create mimetype entry: %w", err)
	}
	if _, err := mw.Write([]byte(mimetype)); err != nil {
		return fmt.Errorf("write mimetype: %w", err)
	}

	// 2. META-INF/manifest.xml
	if err := writeEntry(zw, "META-INF/manifest.xml", generateManifestXML()); err != nil {
		return err
	}

	// 3. content.xml
	content := generateContentXML(writer)
	if err := writeEntry(zw, "content.xml", content); err != nil {
		return err
	}

	// 4. styles.xml
	if err := writeEntry(zw, "styles.xml", generateStylesXML()); err != nil {
		return err
	}

	// 5. meta.xml
	if err := writeEntry(zw, "meta.xml", generateMetaXML()); err != nil {
		return err
	}

	// 6. settings.xml
	if err := writeEntry(zw, "settings.xml", generateSettingsXML(writer.FreezeConfig)); err != nil {
		return err
	}

	return nil
}

func writeEntry(zw *zip.Writer, name, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	return nil
}

// --- XML Helpers ---

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func textCell(value, style, validation string) string {
	var attrs strings.Builder
	attrs.WriteString(` office:value-type="string"`)
	if style != "" {
		attrs.WriteString(fmt.Sprintf(` table:style-name="%s"`, style))
	}
	if validation != "" {
		attrs.WriteString(fmt.Sprintf(` table:content-validation-name="%s"`, validation))
	}
	return fmt.Sprintf(`<table:table-cell%s><text:p>%s</text:p></table:table-cell>`,
		attrs.String(), xmlEscape(value))
}

func numCell(value float64, style, validation string) string {
	var attrs strings.Builder
	attrs.WriteString(fmt.Sprintf(` office:value-type="float" office:value="%g"`, value))
	if style != "" {
		attrs.WriteString(fmt.Sprintf(` table:style-name="%s"`, style))
	}
	if validation != "" {
		attrs.WriteString(fmt.Sprintf(` table:content-validation-name="%s"`, validation))
	}
	return fmt.Sprintf(`<table:table-cell%s><text:p>%g</text:p></table:table-cell>`,
		attrs.String(), value)
}

func formulaCell(formula, style, cached string) string {
	var attrs strings.Builder
	attrs.WriteString(fmt.Sprintf(` table:formula="of:=%s"`, xmlEscape(formula)))
	attrs.WriteString(` office:value-type="string"`)
	if style != "" {
		attrs.WriteString(fmt.Sprintf(` table:style-name="%s"`, style))
	}
	return fmt.Sprintf(`<table:table-cell%s><text:p>%s</text:p></table:table-cell>`,
		attrs.String(), xmlEscape(cached))
}

func emptyCell(style, validation string, repeated int) string {
	var attrs strings.Builder
	if style != "" {
		attrs.WriteString(fmt.Sprintf(` table:style-name="%s"`, style))
	}
	if validation != "" {
		attrs.WriteString(fmt.Sprintf(` table:content-validation-name="%s"`, validation))
	}
	if repeated > 1 {
		attrs.WriteString(fmt.Sprintf(` table:number-columns-repeated="%d"`, repeated))
	}
	return fmt.Sprintf(`<table:table-cell%s/>`, attrs.String())
}

func dateCell(date, style, validation string) string {
	var attrs strings.Builder
	attrs.WriteString(fmt.Sprintf(` office:value-type="date" office:date-value="%s"`, date))
	if style != "" {
		attrs.WriteString(fmt.Sprintf(` table:style-name="%s"`, style))
	}
	if validation != "" {
		attrs.WriteString(fmt.Sprintf(` table:content-validation-name="%s"`, validation))
	}
	return fmt.Sprintf(`<table:table-cell%s><text:p>%s</text:p></table:table-cell>`,
		attrs.String(), date)
}

// --- XML Generation ---

func generateContentXML(w *ODSWriter) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-content
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
  xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"
  xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0"
  xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"
  xmlns:of="urn:oasis:names:tc:opendocument:xmlns:of:1.2"
  office:version="1.2">

<office:automatic-styles>
`)
	b.WriteString(generateAutomaticStyles(w))
	b.WriteString(`</office:automatic-styles>

<office:body>
<office:spreadsheet>

`)
	// Content validations
	b.WriteString(`<table:content-validations>
`)
	for _, sheet := range w.Sheets {
		for _, v := range sheet.Validations {
			b.WriteString(generateValidation(v))
		}
	}
	b.WriteString(`</table:content-validations>

`)
	// Sheets
	for _, sheet := range w.Sheets {
		b.WriteString(generateSheet(sheet))
	}

	b.WriteString(`</office:spreadsheet>
</office:body>
</office:document-content>
`)
	return b.String()
}

func generateAutomaticStyles(w *ODSWriter) string {
	var b strings.Builder
	seen := make(map[string]bool)

	for _, sheet := range w.Sheets {
		for i, col := range sheet.Columns {
			name := fmt.Sprintf("co%d", i+1)
			if !seen[name] {
				b.WriteString(fmt.Sprintf(`<style:style style:name="%s" style:family="table-column"><style:table-column-properties style:column-width="%.1fcm" style:use-optimal-column-width="false"/></style:style>`+"\n", name, col.Width))
				seen[name] = true
			}
		}
	}

	// Cell styles
	cellStyles := []struct{ name, xml string }{
		{"ce_body", `<style:style style:name="ce_body" style:family="table-cell"><style:table-cell-properties fo:border="0.05cm solid #C9D3E0" style:cell-protect="none"/></style:style>`},
		{"ce_center", `<style:style style:name="ce_center" style:family="table-cell"><style:table-cell-properties fo:border="0.05cm solid #C9D3E0" fo:text-align="center"/></style:style>`},
		{"ce_ro", `<style:style style:name="ce_ro" style:family="table-cell"><style:table-cell-properties fo:border="0.05cm solid #C9D3E0" style:cell-protect="protected"/></style:style>`},
		{"hdr", `<style:style style:name="hdr" style:family="table-cell"><style:table-cell-properties fo:background-color="#1F4E79" fo:border="0.05cm solid #1F4E79" fo:padding="0.15cm" style:wrap-option="wrap" style:vertical-align="middle" fo:text-align="center"/><style:text-properties fo:color="#FFFFFF" fo:font-weight="bold"/></style:style>`},
		{"ce_iv", `<style:style style:name="ce_iv" style:family="table-cell"><style:table-cell-properties fo:border="0.05cm solid #C9D3E0" fo:text-align="center"><style:data-style-name>pct</style:data-style-name></style:table-cell-properties></style:style>`},
		{"ce_date", `<style:style style:name="ce_date" style:family="table-cell"><style:table-cell-properties fo:border="0.05cm solid #C9D3E0"><style:data-style-name>dat</style:data-style-name></style:table-cell-properties></style:style>`},
		{"howto_h1", `<style:style style:name="howto_h1" style:family="table-cell"><style:text-properties fo:font-weight="bold" fo:font-size="16pt" fo:color="#1F4E79"/></style:style>`},
		{"howto_h2", `<style:style style:name="howto_h2" style:family="table-cell"><style:text-properties fo:font-weight="bold" fo:font-size="12pt" fo:color="#1F4E79"/></style:style>`},
		{"howto_text", `<style:style style:name="howto_text" style:family="table-cell"><style:table-cell-properties fo:padding="0.1cm" style:wrap-option="wrap" style:line-height="130%"/></style:style>`},
	}
	for _, cs := range cellStyles {
		b.WriteString(cs.xml + "\n")
	}

	// Row style
	b.WriteString(`<style:style style:name="ro_hdr" style:family="table-row"><style:table-row-properties style:row-height="0.7cm" style:use-optimal-row-height="false"/></style:style>` + "\n")

	// Table styles
	b.WriteString(`<style:style style:name="tbl" style:family="table"><style:table-properties style:display="true"/></style:style>` + "\n")
	b.WriteString(`<style:style style:name="tbl_hidden" style:family="table"><style:table-properties style:display="false"/></style:style>` + "\n")

	// Data styles
	b.WriteString(`<number:number-style style:name="pct"><number:number number:min-integer-digits="1" number:decimal-places="1" number:min-decimal-places="1"/><number:text>%</number:text></number:number-style>` + "\n")
	b.WriteString(`<number:date-style style:name="dat"><number:year number:style="long"/><number:text>-</number:text><number:month number:style="long"/><number:text>-</number:text><number:day number:style="long"/></number:date-style>` + "\n")

	return b.String()
}

func generateValidation(v Validation) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<table:content-validation table:name="%s"`, v.Name))
	if v.AllowEmpty {
		b.WriteString(` table:allow-empty-cell="true"`)
	}
	b.WriteString(fmt.Sprintf(` table:condition="%s"`, xmlEscape(v.Condition)))
	b.WriteString(fmt.Sprintf(` table:base-cell-address="%s"`, v.BaseCell))
	if v.DisplayList != "" {
		b.WriteString(fmt.Sprintf(` table:display-list="%s"`, v.DisplayList))
	}
	b.WriteString(">\n")
	if v.ErrorMsg != "" {
		b.WriteString(fmt.Sprintf(`  <table:error-message table:display="true" table:message-type="stop"><text:p>%s</text:p></table:error-message>`+"\n", xmlEscape(v.ErrorMsg)))
	}
	if v.HelpMsg != "" {
		b.WriteString(fmt.Sprintf(`  <table:help-message table:display="true"><text:p>%s</text:p></table:help-message>`+"\n", xmlEscape(v.HelpMsg)))
	}
	b.WriteString("</table:content-validation>\n")
	return b.String()
}

func generateSheet(sheet SheetDef) string {
	var b strings.Builder

	style := "tbl"
	if sheet.Hidden {
		style = "tbl_hidden"
	}

	b.WriteString(fmt.Sprintf(`<table:table table:name="%s" table:style-name="%s">`+"\n", xmlEscape(sheet.Name), style))

	// Columns
	for _, col := range sheet.Columns {
		idx := 1
		for i, c := range sheet.Columns {
			if c.Width == col.Width {
				idx = i + 1
				break
			}
		}
		b.WriteString(fmt.Sprintf(`  <table:table-column table:style-name="co%d" table:default-cell-style-name="ce_body"/>`+"\n", idx))
	}

	// Header row
	if len(sheet.Header) > 0 {
		b.WriteString("  <table:table-header-rows>\n")
		b.WriteString("    <table:table-row table:style-name=\"ro_hdr\">\n")
		for _, h := range sheet.Header {
			b.WriteString(fmt.Sprintf(`      <table:table-cell table:style-name="hdr" office:value-type="string"><text:p>%s</text:p></table:table-cell>`+"\n", xmlEscape(h)))
		}
		b.WriteString("    </table:table-row>\n")
		b.WriteString("  </table:table-header-rows>\n")
	}

	// Data rows
	for _, row := range sheet.Rows {
		b.WriteString("    <table:table-row>\n")
		for _, cell := range row {
			b.WriteString("      ")
			switch cell.Type {
			case "text":
				b.WriteString(textCell(cell.Value, cell.Style, cell.Validation))
			case "float":
				b.WriteString(numCell(cell.FloatVal, cell.Style, cell.Validation))
			case "date":
				b.WriteString(dateCell(cell.DateVal, cell.Style, cell.Validation))
			case "formula":
				b.WriteString(formulaCell(cell.Formula, cell.Style, cell.Value))
			case "empty":
				b.WriteString(emptyCell(cell.Style, cell.Validation, cell.Repeated))
			default:
				b.WriteString(emptyCell("", "", 0))
			}
			b.WriteString("\n")
		}
		b.WriteString("    </table:table-row>\n")
	}

	b.WriteString("</table:table>\n\n")
	return b.String()
}

func generateManifestXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest
  xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0"
  manifest:version="1.2">
  <manifest:file-entry manifest:media-type="application/vnd.oasis.opendocument.spreadsheet" manifest:full-path="/"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="content.xml"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="styles.xml"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="meta.xml"/>
  <manifest:file-entry manifest:media-type="text/xml" manifest:full-path="settings.xml"/>
</manifest:manifest>
`
}

func generateMetaXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<office:document-meta
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0"
  xmlns:dc="http://purl.org/dc/elements/1.1/"
  office:version="1.2">
  <office:meta>
    <meta:generator>Pokemon GO ODS Generator</meta:generator>
  </office:meta>
</office:document-meta>
`
}

func generateStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<office:document-styles
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"
  xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"
  office:version="1.2">
  <office:styles>
    <style:default-style style:family="table-cell">
      <style:table-cell-properties fo:background-color="transparent"/>
      <style:text-properties fo:font-family="Liberation Sans" fo:font-size="10pt"/>
    </style:default-style>
  </office:styles>
</office:document-styles>
`
}

func generateSettingsXML(freezeConfig map[string]FreezePane) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-settings
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:config="urn:oasis:names:tc:opendocument:xmlns:config:1.0"
  xmlns:ooo="http://openoffice.org/2004/office"
  office:version="1.2">
  <office:settings>
    <config:config-item-set config:name="ooo:configuration">
      <config:config-item-map-indexed config:name="Views">
        <config:config-item-map-entry>
`)
	// Active table (first sheet)
	b.WriteString(`          <config:config-item config:name="ActiveTable" config:type="string">Kolekcja</config:config-item>` + "\n")

	for name, fp := range freezeConfig {
		hSplitMode := "0"
		vSplitMode := "0"
		activeSplitRange := "0"
		if fp.Columns > 0 {
			hSplitMode = "2"
			activeSplitRange = "3"
		}
		if fp.Rows > 0 {
			vSplitMode = "2"
			activeSplitRange = "3"
		}
		if hSplitMode == "2" && vSplitMode == "2" {
			activeSplitRange = "3"
		}

		b.WriteString(fmt.Sprintf(`          <config:config-item-map-entry config:name="%s">`+"\n", xmlEscape(name)))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="CursorPositionX" config:type="int">0</config:config-item>` + "\n"))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="CursorPositionY" config:type="int">0</config:config-item>` + "\n"))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="HorizontalSplitMode" config:type="short">%s</config:config-item>`+"\n", hSplitMode))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="VerticalSplitMode" config:type="short">%s</config:config-item>`+"\n", vSplitMode))
		if fp.Columns > 0 {
			b.WriteString(fmt.Sprintf(`            <config:config-item config:name="HorizontalSplitPosition" config:type="int">%d</config:config-item>`+"\n", fp.Columns))
		}
		if fp.Rows > 0 {
			b.WriteString(fmt.Sprintf(`            <config:config-item config:name="VerticalSplitPosition" config:type="int">%d</config:config-item>`+"\n", fp.Rows))
		}
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="ActiveSplitRange" config:type="short">%s</config:config-item>`+"\n", activeSplitRange))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="PositionLeft" config:type="int">0</config:config-item>` + "\n"))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="PositionRight" config:type="int">0</config:config-item>` + "\n"))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="PositionTop" config:type="int">0</config:config-item>` + "\n"))
		b.WriteString(fmt.Sprintf(`            <config:config-item config:name="PositionBottom" config:type="int">0</config:config-item>` + "\n"))
		b.WriteString(`          </config:config-item-map-entry>` + "\n")
	}

	b.WriteString(`        </config:config-item-map-entry>
      </config:config-item-map-indexed>
      <config:config-item config:name="ZoomValue" config:type="int">100</config:config-item>
    </config:config-item-set>
  </office:settings>
</office:document-settings>
`)
	return b.String()
}
