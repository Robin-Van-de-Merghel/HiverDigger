package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&CachedPlugin{})
}

// CachedPlugin extracts cached domain logon information
type CachedPlugin struct{}

func (p *CachedPlugin) Name() string {
	return "cached"
}

func (p *CachedPlugin) Description() string {
	return "Cached domain logons - SECURITY\\Cache"
}

func (p *CachedPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM", "SECURITY"}
}

func (p *CachedPlugin) Run(hive *regf.Hive) error {
	cachePath := "Policy\\Secrets\\NL$*"
	cacheKey, err := hive.GetKey(cachePath)
	if err != nil {
		// Try alternate location
		cachePath = "Cache"
		cacheKey, err = hive.GetKey(cachePath)
		if err != nil {
			return fmt.Errorf("cache key not found: %w", err)
		}
	}

	fmt.Println("Cached Domain Logons:")
	fmt.Println(strings.Repeat("=", 80))

	for _, entry := range cacheKey.Subkeys() {
		fmt.Printf("\nEntry: %s\n", entry.Name())
		fmt.Printf("Last Write: %s\n", entry.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range entry.Values() {
			if len(val.Bytes()) > 0 && len(val.Bytes()) < 1000 {
				fmt.Printf("  %s: %s\n", val.Name(), GetValueString(val))
			}
		}
	}

	return nil
}
