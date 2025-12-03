// Package regf provides a pure-Go parser for Windows Registry hive files (REGF format).
// This is a forensic-friendly, best-effort parser that exposes raw cell bytes and offsets
// for analysis and plugin development.
package regf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// REGF signature at offset 0
	regfSignature = "regf"
	// HBIN signature for hive bins
	hbinSignature = "hbin"
	// Data section starts at offset 0x1000
	dataOffset = 0x1000
)

var (
	ErrInvalidSignature = errors.New("invalid REGF signature")
	ErrInvalidHive      = errors.New("invalid hive file")
)

// Hive represents an open Windows Registry hive file.
type Hive struct {
	data     []byte           // Complete file contents
	fileSize int64            // Total file size
	cells    map[int64]*Cell  // Map of offset -> Cell
	keys     map[int64]*Key   // Map of offset -> Key (parsed NK cells)
	values   map[int64]*Value // Map of offset -> Value (parsed VK cells)
	rootKey  *Key             // Root key of the hive
	closer   io.Closer        // Optional closer for file handle
}

// OpenFile opens a registry hive file from disk.
func OpenFile(path string) (*Hive, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	hive, err := OpenReader(f)
	if err != nil {
		fileCloseError := f.Close()
		if fileCloseError != nil {
			return nil, fmt.Errorf("could not close the file while handling %v", err)
		}
		return nil, err
	}

	hive.closer = f
	return hive, nil
}

// OpenReader opens a registry hive from an io.Reader.
// The entire hive is read into memory for parsing.
func OpenReader(r io.Reader) (*Hive, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read hive data: %w", err)
	}

	if len(data) < 0x1000 {
		return nil, ErrInvalidHive
	}

	// Validate REGF signature
	if string(data[0:4]) != regfSignature {
		return nil, ErrInvalidSignature
	}

	hive := &Hive{
		data:     data,
		fileSize: int64(len(data)),
		cells:    make(map[int64]*Cell),
		keys:     make(map[int64]*Key),
		values:   make(map[int64]*Value),
	}

	// Scan for cells starting at data offset
	if err := hive.scanCells(); err != nil {
		return nil, fmt.Errorf("failed to scan cells: %w", err)
	}

	// Parse NK and VK cells
	hive.parseCells()

	// Find root key
	hive.findRootKey()

	return hive, nil
}

// scanCells performs a best-effort scan for HBIN blocks and cells starting at offset 0x1000.
func (h *Hive) scanCells() error {
	offset := int64(dataOffset)

	for offset < h.fileSize {
		// Check for HBIN signature
		if offset+4 > h.fileSize {
			break
		}

		sig := string(h.data[offset : offset+4])
		if sig == hbinSignature {
			// Read HBIN header
			if offset+0x20 > h.fileSize {
				break
			}

			hbinSize := binary.LittleEndian.Uint32(h.data[offset+8 : offset+12])
			if hbinSize == 0 || int64(hbinSize) > h.fileSize-offset {
				// Invalid HBIN size, skip
				offset += 0x1000
				continue
			}

			// Parse cells within this HBIN
			cellOffset := offset + 0x20 // Skip HBIN header
			hbinEnd := offset + int64(hbinSize)

			for cellOffset < hbinEnd && cellOffset < h.fileSize {
				if cellOffset+4 > h.fileSize {
					break
				}

				// Read cell size (signed 32-bit integer)
				cellSizeRaw := int32(binary.LittleEndian.Uint32(h.data[cellOffset : cellOffset+4]))
				if cellSizeRaw == 0 {
					break
				}

				var cellSize int64
				var allocated bool
				if cellSizeRaw < 0 {
					// Negative size means allocated cell
					cellSize = int64(-cellSizeRaw)
					allocated = true
				} else {
					// Positive size means free cell
					cellSize = int64(cellSizeRaw)
					allocated = false
				}

				if cellSize < 4 || cellOffset+cellSize > hbinEnd {
					// Invalid cell, move to next possible location
					cellOffset += 4
					continue
				}

				if allocated {
					// Store allocated cell
					cell := &Cell{
						offset:    cellOffset,
						size:      cellSize,
						allocated: allocated,
						hive:      h,
					}
					h.cells[cellOffset] = cell
				}

				cellOffset += cellSize
			}

			offset = hbinEnd
		} else {
			// Not an HBIN, skip to next page boundary
			offset += 0x1000
		}
	}

	return nil
}

// parseCells parses NK and VK cells from the scanned cells.
func (h *Hive) parseCells() {
	for offset, cell := range h.cells {
		payload := cell.Payload()
		if len(payload) < 2 {
			continue
		}

		sig := string(payload[0:2])
		switch sig {
		case "nk":
			if key := parseNK(h, offset, payload); key != nil {
				h.keys[offset] = key
			}
		case "vk":
			if value := parseVK(h, offset, payload); value != nil {
				h.values[offset] = value
			}
		}
	}
}

// findRootKey attempts to find the root key of the hive.
func (h *Hive) findRootKey() {
	// Root key offset is typically stored in the header at offset 0x24
	if len(h.data) < 0x28 {
		return
	}

	rootOffset := binary.LittleEndian.Uint32(h.data[0x24:0x28])
	absOffset := int64(dataOffset) + int64(rootOffset)

	if key, ok := h.keys[absOffset]; ok {
		h.rootKey = key
	}
}

// RootKey returns the root key of the hive.
func (h *Hive) RootKey() *Key {
	return h.rootKey
}

// GetKey retrieves a key by its registry path (e.g., "ControlSet001\\Services\\Tcpip").
// Path separators can be either \\ or /.
func (h *Hive) GetKey(path string) (*Key, error) {
	if h.rootKey == nil {
		return nil, errors.New("no root key found")
	}

	if path == "" || path == "\\" || path == "/" {
		return h.rootKey, nil
	}

	// Normalize path separators
	pathParts := splitPath(path)

	current := h.rootKey
	for _, part := range pathParts {
		if part == "" {
			continue
		}

		found := false
		for _, subkey := range current.Subkeys() {
			if equalsCaseInsensitive(subkey.Name(), part) {
				current = subkey
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("key not found: %s", path)
		}
	}

	return current, nil
}

// RawCellAt returns the raw cell data at the given offset.
func (h *Hive) RawCellAt(offset int64) ([]byte, error) {
	cell, ok := h.cells[offset]
	if !ok {
		return nil, fmt.Errorf("no cell at offset 0x%x", offset)
	}
	return cell.RawData(), nil
}

// IterateCells calls fn for each allocated cell in the hive.
func (h *Hive) IterateCells(fn func(offset int64, cell *Cell) bool) {
	for offset, cell := range h.cells {
		if !fn(offset, cell) {
			break
		}
	}
}

// FileSize returns the size of the hive file in bytes.
func (h *Hive) FileSize() int64 {
	return h.fileSize
}

// Close closes the hive and releases any associated resources.
func (h *Hive) Close() error {
	if h.closer != nil {
		return h.closer.Close()
	}
	return nil
}

// splitPath splits a registry path into parts, handling both \ and / separators.
func splitPath(path string) []string {
	var parts []string
	var current string

	for _, ch := range path {
		if ch == '\\' || ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// equalsCaseInsensitive compares two strings case-insensitively (ASCII only).
func equalsCaseInsensitive(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]

		// Convert to lowercase
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}

		if ca != cb {
			return false
		}
	}

	return true
}
