package regf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFile_NonExistent(t *testing.T) {
	_, err := OpenFile("/nonexistent/file")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestOpenReader_InvalidSignature(t *testing.T) {
	data := make([]byte, 0x1000)
	copy(data[0:4], []byte("INVD"))

	_, err := OpenReader(bytes.NewReader(data))
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestOpenReader_TooSmall(t *testing.T) {
	data := make([]byte, 100)

	_, err := OpenReader(bytes.NewReader(data))
	if err != ErrInvalidHive {
		t.Errorf("expected ErrInvalidHive, got %v", err)
	}
}

func TestOpenReader_ValidSignature(t *testing.T) {
	data := make([]byte, 0x1000)
	copy(data[0:4], []byte("regf"))

	hive, err := OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Errorf("failed to open valid REGF: %v", err)
	}

	if hive == nil {
		t.Error("hive is nil")
		return
	}

	if hive.fileSize != int64(len(data)) {
		t.Errorf("expected fileSize %d, got %d", len(data), hive.fileSize)
	}

	defer func() {
		if err := hive.Close(); err != nil {
			t.Fatalf("could not close a hive")
		}
	}()
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"\\", []string{}},
		{"ControlSet001", []string{"ControlSet001"}},
		{"ControlSet001\\Services", []string{"ControlSet001", "Services"}},
		{"ControlSet001\\Services\\Tcpip", []string{"ControlSet001", "Services", "Tcpip"}},
		{"ControlSet001/Services/Tcpip", []string{"ControlSet001", "Services", "Tcpip"}},
		{"\\ControlSet001\\Services\\", []string{"ControlSet001", "Services"}},
	}

	for _, tt := range tests {
		result := splitPath(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitPath(%q): expected %d parts, got %d", tt.input, len(tt.expected), len(result))
			continue
		}

		for i, part := range result {
			if part != tt.expected[i] {
				t.Errorf("splitPath(%q)[%d]: expected %q, got %q", tt.input, i, tt.expected[i], part)
			}
		}
	}
}

func TestEqualsCaseInsensitive(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected bool
	}{
		{"", "", true},
		{"abc", "abc", true},
		{"abc", "ABC", true},
		{"ABC", "abc", true},
		{"aBc", "AbC", true},
		{"abc", "abd", false},
		{"abc", "ab", false},
		{"ab", "abc", false},
	}

	for _, tt := range tests {
		result := equalsCaseInsensitive(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("equalsCaseInsensitive(%q, %q): expected %v, got %v", tt.a, tt.b, tt.expected, result)
		}
	}
}

func TestOpenFile_WithRealHive(t *testing.T) {
	// Check if example SYSTEM hive exists
	hivePath := filepath.Join("../../example/config/SYSTEM")
	if _, err := os.Stat(hivePath); os.IsNotExist(err) {
		t.Skip("example SYSTEM hive not found")
	}

	hive, err := OpenFile(hivePath)
	if err != nil {
		t.Fatalf("failed to open example SYSTEM hive: %v", err)
	}
	defer func() {
		if err := hive.Close(); err != nil {
			t.Fatalf("failed to close a hive")
		}
	}()

	// Verify we can get root key
	root := hive.RootKey()
	if root == nil {
		t.Error("root key is nil")
	}

	// Verify file size is reasonable
	if hive.FileSize() < 1024 {
		t.Errorf("file size seems too small: %d", hive.FileSize())
	}

	// Try to get Select key
	selectKey, err := hive.GetKey("Select")
	if err != nil {
		t.Errorf("failed to get Select key: %v", err)
	}

	if selectKey == nil {
		t.Error("Select key is nil")
	}
}

func TestGetKey_CaseInsensitive(t *testing.T) {
	hivePath := filepath.Join("../../example/config/SYSTEM")
	if _, err := os.Stat(hivePath); os.IsNotExist(err) {
		t.Skip("example SYSTEM hive not found")
	}

	hive, err := OpenFile(hivePath)
	if err != nil {
		t.Fatalf("failed to open example SYSTEM hive: %v", err)
	}
	defer func() {
		if err := hive.Close(); err != nil {
			t.Fatalf("failed to close a hive")
		}
	}()

	// Try different case variations
	key1, err1 := hive.GetKey("Select")
	key2, err2 := hive.GetKey("select")
	key3, err3 := hive.GetKey("SELECT")

	if err1 != nil || err2 != nil || err3 != nil {
		t.Errorf("case-insensitive lookup failed: %v, %v, %v", err1, err2, err3)
	}

	if key1 == nil || key2 == nil || key3 == nil {
		t.Error("one or more keys are nil")
	}
}

func TestRawCellAt(t *testing.T) {
	hivePath := filepath.Join("../../example/config/SYSTEM")
	if _, err := os.Stat(hivePath); os.IsNotExist(err) {
		t.Skip("example SYSTEM hive not found")
	}

	hive, err := OpenFile(hivePath)
	if err != nil {
		t.Fatalf("failed to open example SYSTEM hive: %v", err)
	}
	defer func() {
		if err := hive.Close(); err != nil {
			t.Fatalf("failed to close a hive")
		}
	}()

	// Try to get raw cell at offset 0x1020 (typical first cell)
	raw, err := hive.RawCellAt(0x1020)
	if err != nil {
		t.Logf("no cell at 0x1020 (this is ok): %v", err)
	} else if len(raw) == 0 {
		t.Error("raw cell data is empty")
	}
}

func TestIterateCells(t *testing.T) {
	hivePath := filepath.Join("../../example/config/SYSTEM")
	if _, err := os.Stat(hivePath); os.IsNotExist(err) {
		t.Skip("example SYSTEM hive not found")
	}

	hive, err := OpenFile(hivePath)
	if err != nil {
		t.Fatalf("failed to open example SYSTEM hive: %v", err)
	}
	defer func() {
		if err := hive.Close(); err != nil {
			t.Fatalf("failed to close a hive")
		}
	}()

	count := 0
	hive.IterateCells(func(offset int64, cell *Cell) bool {
		count++
		return count < 10 // Stop after 10 cells
	})

	if count == 0 {
		t.Error("no cells found during iteration")
	}
}
