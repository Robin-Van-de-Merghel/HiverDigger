package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&RecentAppsPlugin{})
}

// RecentAppsPlugin extracts recently used applications (Win10+)
type RecentAppsPlugin struct{}

func (p *RecentAppsPlugin) Name() string {
	return "recentapps"
}

func (p *RecentAppsPlugin) Description() string {
	return "Recently used applications (Windows 10+)"
}

func (p *RecentAppsPlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT"}
}

func (p *RecentAppsPlugin) Run(hive *regf.Hive) error {
	appsPath := "Software\\Microsoft\\Windows\\CurrentVersion\\Search\\RecentApps"
	appsKey, err := hive.GetKey(appsPath)
	if err != nil {
		return fmt.Errorf("RecentApps key not found: %w", err)
	}

	fmt.Println("Recently Used Applications:")
	fmt.Println(strings.Repeat("=", 80))

	for _, app := range appsKey.Subkeys() {
		fmt.Printf("\n[%s]\n", app.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range app.Values() {
			if val.Name() == "AppId" || val.Name() == "AppPath" ||
				val.Name() == "LastAccessedTime" {
				fmt.Printf("  %s: %s\n", val.Name(), GetValueString(val))
			}
		}
	}

	return nil
}
