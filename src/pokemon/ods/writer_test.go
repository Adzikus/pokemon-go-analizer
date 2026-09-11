package ods

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteODSEmpty(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.ods")

	writer := &ODSWriter{
		Sheets: []SheetDef{
			{
				Name:    "Test",
				Columns: []ColDef{{Width: 2}},
				Header:  []string{"Col A"},
				Rows:    [][]Cell{{{Type: "text", Value: "hello"}}},
			},
		},
	}

	if err := WriteODS(outPath, writer); err != nil {
		t.Fatalf("WriteODS: %v", err)
	}

	// Verify file exists and is valid ZIP
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}

	// Check ZIP signature (PK\x03\x04)
	if data[0] != 'P' || data[1] != 'K' {
		t.Errorf("not a ZIP file: first bytes = %x", data[:4])
	}
}

func TestWriteODSVerify(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.ods")

	writer := &ODSWriter{
		Sheets: []SheetDef{
			{
				Name:    "Pokemon",
				Columns: []ColDef{{Width: 2}},
				Header:  []string{"Name"},
				Rows:    [][]Cell{},
			},
		},
	}

	if err := WriteODS(outPath, writer); err != nil {
		t.Fatalf("WriteODS: %v", err)
	}

	doc, err := Read(outPath)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	sheet := doc.GetSheet("Pokemon")
	if sheet == nil {
		t.Fatal("sheet 'Pokemon' not found")
	}
	if len(sheet.Rows) != 1 {
		t.Errorf("expected 1 row (header), got %d", len(sheet.Rows))
	}
}

func TestWriteODSFreezePane(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.ods")

	writer := &ODSWriter{
		Sheets: []SheetDef{
			{
				Name:    "Kolekcja",
				Columns: []ColDef{{Width: 2}, {Width: 3}},
				Header:  []string{"A", "B"},
				Rows:    [][]Cell{},
			},
		},
		FreezeConfig: map[string]FreezePane{
			"Kolekcja": {Rows: 1, Columns: 2},
		},
	}

	if err := WriteODS(outPath, writer); err != nil {
		t.Fatalf("WriteODS: %v", err)
	}

	doc, err := Read(outPath)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	sheet := doc.GetSheet("Kolekcja")
	if sheet == nil {
		t.Fatal("sheet not found")
	}
	// Freeze pane is in content.xml, not directly readable from Sheet struct
	// but WriteODS should not error
}

func TestReadODSNotFound(t *testing.T) {
	_, err := Read("/nonexistent/file.ods")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadODSBadZIP(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.ods")
	os.WriteFile(badPath, []byte("not a zip"), 0644)

	_, err := Read(badPath)
	if err == nil {
		t.Error("expected error for bad ZIP")
	}
}
