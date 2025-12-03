package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&AmCachePlugin{})
}

// AmCachePlugin displays AmCache entries showing program execution artifacts.
type AmCachePlugin struct{}

func (p *AmCachePlugin) Name() string {
	return "amcache"
}

func (p *AmCachePlugin) Description() string {
	return "Display AmCache entries (program execution artifacts) from AmCache.hve"
}

func (p *AmCachePlugin) CompatibleHiveTypes() []string {
	return []string{"AMCACHE.HVE", "SYSCACHE.HVE"}
}

func (p *AmCachePlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Root/File",
		"Root/InventoryApplicationFile",
	}

	fmt.Println("AmCache Entries:")
	fmt.Println(strings.Repeat("=", 80))

	found := false
	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		found = true
		fmt.Printf("\n%s:\n", path)

		count := 0
		for _, subkey := range key.Subkeys() {
			var fileName, filePath, sha1 string

			for _, val := range subkey.Values() {
				valStr := GetValueString(val)
				switch val.Name() {
				case "FileName", "Name":
					fileName = valStr
				case "FullPath", "LowerCaseLongPath", "LongPathHash":
					filePath = valStr
				case "FileId", "SHA1", "Sha1":
					sha1 = valStr
				}
			}

			if fileName != "" || filePath != "" {
				count++
				if count <= 50 { // Limit output
					fmt.Printf("\n  %s\n", subkey.Name())
					if fileName != "" {
						fmt.Printf("    File: %s\n", fileName)
					}
					if filePath != "" {
						fmt.Printf("    Path: %s\n", filePath)
					}
					if sha1 != "" {
						fmt.Printf("    SHA1: %s\n", sha1)
					}
					fmt.Printf("    Last Write: %s\n", subkey.Timestamp().Format("2006-01-02 15:04:05"))
				}
			}
		}

		if count > 50 {
			fmt.Printf("\n  ... and %d more entries (showing first 50)\n", count-50)
		}
		fmt.Printf("\n  Total entries: %d\n", count)
	}

	if !found {
		fmt.Println("No AmCache entries found")
		fmt.Println("Note: This plugin requires AmCache.hve file")
	}

	return nil
}
