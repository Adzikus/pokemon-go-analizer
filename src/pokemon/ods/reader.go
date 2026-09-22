package ods

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Sheet represents a single ODS sheet with rows of cells.
type Sheet struct {
	Name string
	Rows [][]string
}

// ODS represents the parsed ODS document.
type ODS struct {
	Sheets []Sheet
}

// Read parses an ODS file and returns all sheets.
func Read(path string) (*ODS, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("otworz ZIP: %w", err)
	}
	defer r.Close()

	var contentFile *zip.File
	for _, f := range r.File {
		if f.Name == "content.xml" {
			contentFile = f
			break
		}
	}
	if contentFile == nil {
		return nil, fmt.Errorf("brak content.xml w ODS")
	}

	rc, err := contentFile.Open()
	if err != nil {
		return nil, fmt.Errorf("otworz content.xml: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("czytaj content.xml: %w", err)
	}

	return parseContentXML(data)
}

// parseContentXML uses a streaming decoder to parse ODS XML.
// Go's XML decoder strips namespace prefixes automatically.
func parseContentXML(data []byte) (*ODS, error) {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	ods := &ODS{}

	var curSheet *Sheet
	var inTable bool

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch start := tok.(type) {
		case xml.StartElement:
			local := start.Name.Local

			switch local {
			case "table":
				name := ""
				for _, a := range start.Attr {
					if a.Name.Local == "name" {
						name = a.Value
					}
				}
				ods.Sheets = append(ods.Sheets, Sheet{Name: name})
				curSheet = &ods.Sheets[len(ods.Sheets)-1]
				inTable = true

			case "table-row":
				if inTable && curSheet != nil {
					curSheet.Rows = append(curSheet.Rows, []string{})
				}

			case "table-cell":
				if inTable && curSheet != nil && len(curSheet.Rows) > 0 {
					// Get repeat count
					repeat := 1
					for _, a := range start.Attr {
						if a.Name.Local == "number-columns-repeated" {
							fmt.Sscanf(a.Value, "%d", &repeat)
						}
					}

					// Read cell content (text inside text:p)
					cellVal := ""
					depth := 1
					for depth > 0 {
						t2, err := dec.Token()
						if err != nil {
							break
						}
						switch t2t := t2.(type) {
						case xml.StartElement:
							depth++
						case xml.EndElement:
							depth--
						case xml.CharData:
							if depth >= 2 {
								cellVal = strings.TrimSpace(string(t2t))
							}
						}
					}

					// Cap repeat for sparse rows
					if repeat > 50 {
						repeat = 1
						cellVal = ""
					}

					rowIdx := len(curSheet.Rows) - 1
					for i := 0; i < repeat; i++ {
						curSheet.Rows[rowIdx] = append(curSheet.Rows[rowIdx], cellVal)
					}
				}
			}

		case xml.EndElement:
			if start.Name.Local == "table" {
				inTable = false
				curSheet = nil
			}
		}
	}

	return ods, nil
}

// GetSheet returns a sheet by name, or nil if not found.
func (o *ODS) GetSheet(name string) *Sheet {
	for i := range o.Sheets {
		if o.Sheets[i].Name == name {
			return &o.Sheets[i]
		}
	}
	return nil
}

// SheetNames returns all sheet names.
func (o *ODS) SheetNames() []string {
	names := make([]string, len(o.Sheets))
	for i := range o.Sheets {
		names[i] = o.Sheets[i].Name
	}
	return names
}

// VerifyBackupZIP checks if a file is a valid ODS (ZIP with content.xml).
func VerifyBackupZIP(path string) error {
	_, err := Read(path)
	if err != nil {
		return fmt.Errorf("otworz ODS: %w", err)
	}
	return nil
}
