package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&MountPointsPlugin{})
}

// MountPointsPlugin displays mounted devices and volumes from SYSTEM hive.
type MountPointsPlugin struct{}

func (p *MountPointsPlugin) Name() string {
	return "mountpoints"
}

func (p *MountPointsPlugin) Description() string {
	return "Display mounted devices and volumes from SYSTEM hive"
}

func (p *MountPointsPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *MountPointsPlugin) Run(hive *regf.Hive) error {
	path := "MountedDevices"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("MountedDevices key not found: %w", err)
	}

	fmt.Println("Mounted Devices")
	fmt.Println("===============")
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()

	for _, val := range key.Values() {
		if val.Name() != "" {
			// The data is binary, but we can show the name
			fmt.Printf("%s\n", val.Name())
		}
	}

	return nil
}
