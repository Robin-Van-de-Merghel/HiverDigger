package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&RecentDocsPlugin{})
}

// RecentDocsPlugin displays recently opened documents from NTUSER.DAT.
type RecentDocsPlugin struct{}

func (p *RecentDocsPlugin) Name() string {
	return "recentdocs"
}

func (p *RecentDocsPlugin) Description() string {
	return "Display recently opened documents from NTUSER.DAT"
}

func (p *RecentDocsPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT"}
}

func (p *RecentDocsPlugin) Run(hive *regf.Hive) error {
	path := "Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\RecentDocs"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("RecentDocs key not found: %w", err)
	}

	fmt.Println("Recently Opened Documents")
	fmt.Println("=========================")
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// List extensions
	for _, extKey := range key.Subkeys() {
		fmt.Printf("Extension: .%s\n", extKey.Name())
		fmt.Printf("  Last Modified: %s\n", extKey.Timestamp().Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}
