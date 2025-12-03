// Package plugins provides a registry for HiveDigger plugins.
package plugins

import (
	"errors"
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

var (
	// ErrPluginNotFound is returned when a plugin is not registered.
	ErrPluginNotFound = errors.New("plugin not found")
)

// Plugin is the interface that all plugins must implement.
type Plugin interface {
	// Name returns the plugin name.
	Name() string
	// Description returns a brief description of what the plugin does.
	Description() string
	// Run executes the plugin logic on the given hive.
	Run(hive *regf.Hive) error
}

// HiveTypePlugin is an optional interface that plugins can implement
// to specify which hive types they are compatible with.
type HiveTypePlugin interface {
	Plugin
	// CompatibleHiveTypes returns a list of hive types this plugin works with.
	// Common types: "SYSTEM", "SOFTWARE", "SAM", "SECURITY", "NTUSER.DAT", "USRCLASS.DAT"
	// Return nil or empty slice to indicate compatibility with all hive types.
	CompatibleHiveTypes() []string
}

var registry = make(map[string]Plugin)
var hiveTypeMap = make(map[string][]string) // plugin name -> compatible hive types

// Register registers a plugin with the given name.
func Register(p Plugin) {
	registry[p.Name()] = p

	// Check if plugin implements HiveTypePlugin
	if htp, ok := p.(HiveTypePlugin); ok {
		types := htp.CompatibleHiveTypes()
		if len(types) > 0 {
			hiveTypeMap[p.Name()] = types
		}
	}
}

// Get retrieves a plugin by name.
func Get(name string) (Plugin, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrPluginNotFound, name)
	}
	return p, nil
}

// List returns a list of all registered plugin names.
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// ListForHiveType returns a list of plugin names compatible with the given hive type.
// If a plugin doesn't specify compatible types, it's considered compatible with all types.
func ListForHiveType(hiveType string) []string {
	names := make([]string, 0)
	for name := range registry {
		if IsCompatibleWithHiveType(name, hiveType) {
			names = append(names, name)
		}
	}
	return names
}

// IsCompatibleWithHiveType checks if a plugin is compatible with the given hive type.
func IsCompatibleWithHiveType(pluginName, hiveType string) bool {
	types, exists := hiveTypeMap[pluginName]
	if !exists || len(types) == 0 {
		// Plugin doesn't specify types, so it's compatible with all
		return true
	}

	// Check if hiveType matches any of the compatible types
	for _, t := range types {
		if t == hiveType {
			return true
		}
	}
	return false
}

// GetValueString is a helper function to extract a string value from a registry value.
// It handles REG_SZ (1), REG_EXPAND_SZ (2), and REG_MULTI_SZ (7) types.
func GetValueString(v *regf.Value) string {
	if v == nil {
		return ""
	}

	data := v.Bytes()
	if len(data) == 0 {
		return ""
	}

	// REG_SZ (1), REG_EXPAND_SZ (2)
	if v.Type() == 1 || v.Type() == 2 {
		return parseNullTerminatedString(data)
	}

	// REG_MULTI_SZ (7) - return first string
	if v.Type() == 7 {
		return parseNullTerminatedString(data)
	}

	// For other types, try to parse as ASCII
	return string(data)
}

// parseNullTerminatedString parses a null-terminated string from byte data.
// It handles both ASCII and UTF-16LE encoding.
func parseNullTerminatedString(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Check if UTF-16LE (every other byte might be 0x00 for ASCII chars)
	isUTF16 := len(data) >= 2 && len(data)%2 == 0
	if isUTF16 {
		zeroCount := 0
		for i := 1; i < len(data) && i < 20; i += 2 {
			if data[i] == 0x00 {
				zeroCount++
			}
		}

		if zeroCount > len(data)/4 {
			// UTF-16LE
			var result []uint16
			for i := 0; i+1 < len(data); i += 2 {
				ch := uint16(data[i]) | uint16(data[i+1])<<8
				if ch == 0 {
					break
				}
				result = append(result, ch)
			}

			// Convert UTF-16 to string
			runes := make([]rune, 0, len(result))
			for _, r := range result {
				runes = append(runes, rune(r))
			}
			return string(runes)
		}
	}

	// ASCII/Latin-1
	end := len(data)
	for i, b := range data {
		if b == 0 {
			end = i
			break
		}
	}

	return string(data[:end])
}

// findCurrentControlSet is a helper to determine the current ControlSet
func findCurrentControlSet(hive *regf.Hive) (string, error) {
	selectKey, err := hive.GetKey("Select")
	if err != nil {
		return "", err
	}

	// Find the "Current" value
	values := selectKey.Values()
	for _, v := range values {
		if v.Name() == "Current" || v.Name() == "current" {
			// Current is a REG_DWORD (4)
			if v.Type() == 4 {
				data := v.Bytes()
				if len(data) >= 4 {
					// Read as little-endian DWORD
					currentNum := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
					return fmt.Sprintf("ControlSet%03d", currentNum), nil
				}
			}
		}
	}

	return "", fmt.Errorf("current value not found in select key")
}

// getSubkey is a helper to get a subkey by name
func getSubkey(key *regf.Key, name string) (*regf.Key, error) {
	for _, sk := range key.Subkeys() {
		if sk.Name() == name {
			return sk, nil
		}
	}
	return nil, fmt.Errorf("subkey %s not found", name)
}

// getValue is a helper to get a value by name
func getValue(key *regf.Key, name string) (*regf.Value, error) {
	for _, v := range key.Values() {
		if v.Name() == name {
			return v, nil
		}
	}
	return nil, fmt.Errorf("value %s not found", name)
}
