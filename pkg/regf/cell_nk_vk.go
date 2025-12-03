package regf

import (
	"encoding/binary"
	"time"
	"unicode/utf16"
)

// Cell represents a registry cell.
type Cell struct {
	offset    int64
	size      int64
	allocated bool
	hive      *Hive
}

// Payload returns the cell data without the size header.
func (c *Cell) Payload() []byte {
	start := c.offset + 4 // Skip 4-byte size field
	end := c.offset + c.size
	if start >= int64(len(c.hive.data)) || end > int64(len(c.hive.data)) {
		return nil
	}
	return c.hive.data[start:end]
}

// RawData returns the complete cell data including the size header.
func (c *Cell) RawData() []byte {
	end := c.offset + c.size
	if c.offset >= int64(len(c.hive.data)) || end > int64(len(c.hive.data)) {
		return nil
	}
	return c.hive.data[c.offset:end]
}

// Key represents a registry key (NK cell).
type Key struct {
	offset       int64
	name         string
	timestamp    time.Time
	parentOffset int64
	subkeyCount  uint32
	valueCount   uint32
	subkeyList   int64
	valueList    int64
	hive         *Hive
}

// Name returns the key name.
func (k *Key) Name() string {
	return k.name
}

// Timestamp returns the last modified timestamp.
func (k *Key) Timestamp() time.Time {
	return k.timestamp
}

// Subkeys returns the list of subkeys.
func (k *Key) Subkeys() []*Key {
	var subkeys []*Key

	if k.subkeyList == 0 || k.subkeyCount == 0 {
		return subkeys
	}

	absOffset := int64(dataOffset) + k.subkeyList
	if absOffset < 0 || absOffset >= k.hive.fileSize {
		return subkeys
	}

	// Read the subkey list cell
	if absOffset+4 > k.hive.fileSize {
		return subkeys
	}

	cellSize := int32(binary.LittleEndian.Uint32(k.hive.data[absOffset : absOffset+4]))
	if cellSize >= 0 {
		// Not an allocated cell
		return subkeys
	}

	payloadStart := absOffset + 4
	if payloadStart+2 > k.hive.fileSize {
		return subkeys
	}

	sig := string(k.hive.data[payloadStart : payloadStart+2])

	// Handle different list types: lf, lh, li, ri
	switch sig {
	case "lf", "lh", "li":
		// Direct list of subkeys
		if payloadStart+4 > k.hive.fileSize {
			return subkeys
		}
		count := readUint16(k.hive.data, payloadStart+2)

		for i := uint16(0); i < count && uint32(i) < k.subkeyCount; i++ {
			entryOffset := payloadStart + 4 + int64(i)*8
			if entryOffset+4 > k.hive.fileSize {
				break
			}

			subkeyOffset := readUint32(k.hive.data, entryOffset)
			subkeyAbsOffset := int64(dataOffset) + int64(subkeyOffset)

			if subkey, ok := k.hive.keys[subkeyAbsOffset]; ok {
				subkeys = append(subkeys, subkey)
			}
		}
	case "ri":
		// Indirect list - list of lists
		if payloadStart+4 > k.hive.fileSize {
			return subkeys
		}
		count := readUint16(k.hive.data, payloadStart+2)

		for i := uint16(0); i < count; i++ {
			entryOffset := payloadStart + 4 + int64(i)*4
			if entryOffset+4 > k.hive.fileSize {
				break
			}

			listOffset := readUint32(k.hive.data, entryOffset)
			listAbsOffset := int64(dataOffset) + int64(listOffset)

			// Recursively read the sub-list
			if listAbsOffset+4 > k.hive.fileSize {
				continue
			}

			listCellSize := int32(binary.LittleEndian.Uint32(k.hive.data[listAbsOffset : listAbsOffset+4]))
			if listCellSize >= 0 {
				continue
			}

			listPayloadStart := listAbsOffset + 4
			if listPayloadStart+4 > k.hive.fileSize {
				continue
			}

			listSig := string(k.hive.data[listPayloadStart : listPayloadStart+2])
			if listSig == "lf" || listSig == "lh" || listSig == "li" {
				listCount := readUint16(k.hive.data, listPayloadStart+2)

				for j := uint16(0); j < listCount; j++ {
					subEntryOffset := listPayloadStart + 4 + int64(j)*8
					if subEntryOffset+4 > k.hive.fileSize {
						break
					}

					subkeyOffset := readUint32(k.hive.data, subEntryOffset)
					subkeyAbsOffset := int64(dataOffset) + int64(subkeyOffset)

					if subkey, ok := k.hive.keys[subkeyAbsOffset]; ok {
						subkeys = append(subkeys, subkey)
					}
				}
			}
		}
	}

	return subkeys
}

// Values returns the list of values.
func (k *Key) Values() []*Value {
	var values []*Value

	if k.valueList == 0 || k.valueCount == 0 {
		return values
	}

	absOffset := int64(dataOffset) + k.valueList
	if absOffset < 0 || absOffset+4 > k.hive.fileSize {
		return values
	}

	// Read the value list cell
	cellSize := int32(binary.LittleEndian.Uint32(k.hive.data[absOffset : absOffset+4]))
	if cellSize >= 0 {
		// Not an allocated cell
		return values
	}

	payloadStart := absOffset + 4

	for i := uint32(0); i < k.valueCount; i++ {
		entryOffset := payloadStart + int64(i)*4
		if entryOffset+4 > k.hive.fileSize {
			break
		}

		valueOffset := readUint32(k.hive.data, entryOffset)
		valueAbsOffset := int64(dataOffset) + int64(valueOffset)

		if value, ok := k.hive.values[valueAbsOffset]; ok {
			values = append(values, value)
		}
	}

	return values
}

// Value represents a registry value (VK cell).
type Value struct {
	offset     int64
	name       string
	dataType   uint32
	dataSize   uint32
	dataOffset int64
	hive       *Hive
}

// Name returns the value name.
func (v *Value) Name() string {
	return v.name
}

// Type returns the value data type.
func (v *Value) Type() uint32 {
	return v.dataType
}

// Bytes returns the raw value data.
func (v *Value) Bytes() []byte {
	if v.dataSize == 0 {
		return nil
	}

	// Check for inline data (size & 0x80000000)
	if v.dataSize&0x80000000 != 0 {
		// Data is stored inline in the offset field
		actualSize := v.dataSize & 0x7FFFFFFF
		if actualSize > 4 {
			actualSize = 4
		}

		// Extract bytes from the dataOffset field
		inlineData := make([]byte, 4)
		binary.LittleEndian.PutUint32(inlineData, uint32(v.dataOffset))
		return inlineData[:actualSize]
	}

	// Data is stored in a separate cell
	absOffset := int64(dataOffset) + v.dataOffset
	if absOffset < 0 || absOffset+4 > v.hive.fileSize {
		return nil
	}

	cellSize := int32(binary.LittleEndian.Uint32(v.hive.data[absOffset : absOffset+4]))
	if cellSize >= 0 {
		// Not an allocated cell
		return nil
	}

	payloadSize := int64(-cellSize) - 4
	payloadStart := absOffset + 4

	actualSize := int64(v.dataSize)
	if actualSize > payloadSize {
		actualSize = payloadSize
	}

	if payloadStart+actualSize > v.hive.fileSize {
		return nil
	}

	return v.hive.data[payloadStart : payloadStart+actualSize]
}

// parseNK parses an NK (key) cell.
func parseNK(h *Hive, offset int64, payload []byte) *Key {
	if len(payload) < 0x50 {
		return nil
	}

	if string(payload[0:2]) != "nk" {
		return nil
	}

	key := &Key{
		offset: offset,
		hive:   h,
	}

	// Flags at offset 0x02
	flags := readUint16(payload, 0x02)

	// Timestamp at 0x04 (8 bytes, Windows FILETIME)
	timestamp := readUint64(payload, 0x04)
	key.timestamp = filetimeToTime(timestamp)

	// Parent offset at 0x10
	key.parentOffset = int64(readUint32(payload, 0x10))

	// Subkey count at 0x14
	key.subkeyCount = readUint32(payload, 0x14)

	// Subkey list offset at 0x1C
	key.subkeyList = int64(readUint32(payload, 0x1C))

	// Value count at 0x24
	key.valueCount = readUint32(payload, 0x24)

	// Value list offset at 0x28
	key.valueList = int64(readUint32(payload, 0x28))

	// Name length at 0x48
	nameLen := readUint16(payload, 0x48)

	// Name at 0x4C
	if len(payload) >= 0x4C+int(nameLen) {
		nameBytes := payload[0x4C : 0x4C+nameLen]
		key.name = parseNameBytes(nameBytes, flags&0x0020 != 0)
	}

	return key
}

// parseVK parses a VK (value) cell.
func parseVK(h *Hive, offset int64, payload []byte) *Value {
	if len(payload) < 0x14 {
		return nil
	}

	if string(payload[0:2]) != "vk" {
		return nil
	}

	value := &Value{
		offset: offset,
		hive:   h,
	}

	// Name length at 0x02
	nameLen := readUint16(payload, 0x02)

	// Data size at 0x04
	value.dataSize = readUint32(payload, 0x04)

	// Data offset at 0x08
	value.dataOffset = int64(readUint32(payload, 0x08))

	// Data type at 0x0C
	value.dataType = readUint32(payload, 0x0C)

	// Flags at 0x10
	flags := readUint16(payload, 0x10)

	// Name at 0x14
	if nameLen > 0 && len(payload) >= 0x14+int(nameLen) {
		nameBytes := payload[0x14 : 0x14+nameLen]
		value.name = parseNameBytes(nameBytes, flags&0x0001 != 0)
	}

	return value
}

// Helper functions

func readUint16(data []byte, offset int64) uint16 {
	if offset+2 > int64(len(data)) {
		return 0
	}
	return binary.LittleEndian.Uint16(data[offset : offset+2])
}

func readUint32(data []byte, offset int64) uint32 {
	if offset+4 > int64(len(data)) {
		return 0
	}
	return binary.LittleEndian.Uint32(data[offset : offset+4])
}

func readUint64(data []byte, offset int64) uint64 {
	if offset+8 > int64(len(data)) {
		return 0
	}
	return binary.LittleEndian.Uint64(data[offset : offset+8])
}

// parseNameBytes parses a name byte array, detecting UTF-16LE encoding.
func parseNameBytes(data []byte, asciiFlag bool) string {
	if len(data) == 0 {
		return ""
	}

	// Check if ASCII flag is set or heuristically detect encoding
	isUTF16 := !asciiFlag && len(data) >= 2 && len(data)%2 == 0

	if isUTF16 {
		// Check for UTF-16LE pattern (every other byte is likely 0x00 for ASCII chars)
		zeroCount := 0
		for i := 1; i < len(data) && i < 20; i += 2 {
			if data[i] == 0x00 {
				zeroCount++
			}
		}

		if zeroCount > len(data)/4 {
			// Likely UTF-16LE
			u16 := make([]uint16, len(data)/2)
			for i := 0; i < len(u16); i++ {
				u16[i] = binary.LittleEndian.Uint16(data[i*2 : i*2+2])
			}
			return string(utf16.Decode(u16))
		}
	}

	// Treat as ASCII/Latin-1
	return string(data)
}

// filetimeToTime converts a Windows FILETIME to time.Time.
// FILETIME is 100-nanosecond intervals since January 1, 1601 UTC.
func filetimeToTime(filetime uint64) time.Time {
	// Windows epoch: January 1, 1601
	// Unix epoch: January 1, 1970
	// Difference: 116444736000000000 * 100ns = 11644473600 seconds
	const windowsToUnixEpoch = 116444736000000000

	if filetime < windowsToUnixEpoch {
		return time.Time{}
	}

	unixNano := int64((filetime - windowsToUnixEpoch) * 100)
	return time.Unix(0, unixNano)
}
