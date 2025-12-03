package plugins

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ShimCachePlugin{})
}

// ShimCachePlugin displays Application Compatibility Cache (ShimCache) entries.
// Based on RegRipper's shimcache.pl plugin.
type ShimCachePlugin struct{}

func (p *ShimCachePlugin) Name() string {
	return "shimcache"
}

func (p *ShimCachePlugin) Description() string {
	return "Display Application Compatibility Cache (ShimCache) entries from SYSTEM hive"
}

func (p *ShimCachePlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *ShimCachePlugin) Run(hive *regf.Hive) error {
	// Find current ControlSet
	controlSetName, err := p.findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current controlset: %w", err)
	}

	// Path to AppCompatCache
	path := fmt.Sprintf("%s/Control/Session Manager/AppCompatCache", controlSetName)
	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("appcompatcache key not found: %w", err)
	}

	fmt.Printf("AppCompatCache entries from %s:\n", path)
	fmt.Println(strings.Repeat("=", 80))

	// Get AppCompatCache value
	for _, val := range key.Values() {
		if val.Name() == "AppCompatCache" {
			data := val.Bytes()
			fmt.Printf("\nFound AppCompatCache data (%d bytes)\n", len(data))

			if len(data) < 16 {
				fmt.Println("Data too small to parse")
				return nil
			}

			// Parse header (varies by Windows version)
			// This is a simplified parser
			if len(data) >= 4 {
				signature := binary.LittleEndian.Uint32(data[0:4])
				fmt.Printf("Signature: 0x%08x\n", signature)

				// Note: Full ShimCache parsing is complex and version-dependent
				// This is a basic implementation showing the concept
				fmt.Println("\nNote: Full ShimCache parsing requires version-specific logic.")
				fmt.Println("Data is present but detailed parsing not fully implemented.")
			}
		}
	}

	return nil
}

func (p *ShimCachePlugin) findCurrentControlSet(hive *regf.Hive) (string, error) {
	selectKey, err := hive.GetKey("Select")
	if err != nil {
		return "", err
	}

	for _, val := range selectKey.Values() {
		if val.Name() == "Current" && len(val.Bytes()) >= 4 {
			current := binary.LittleEndian.Uint32(val.Bytes()[:4])
			return fmt.Sprintf("ControlSet%03d", current), nil
		}
	}

	return "", fmt.Errorf("current value not found in select key")
}
