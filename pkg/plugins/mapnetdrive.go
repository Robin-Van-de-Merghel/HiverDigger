package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&MapNetworkDrivePlugin{})
}

// MapNetworkDrivePlugin displays mapped network drives from NTUSER.DAT.
type MapNetworkDrivePlugin struct{}

func (p *MapNetworkDrivePlugin) Name() string {
	return "mapnetdrive"
}

func (p *MapNetworkDrivePlugin) Description() string {
	return "Display mapped network drives from NTUSER.DAT"
}

func (p *MapNetworkDrivePlugin) CompatibleHiveTypes() []string {
	return []string{"NTUSER.DAT"}
}

func (p *MapNetworkDrivePlugin) Run(hive *regf.Hive) error {
	path := "Network"

	key, err := hive.GetKey(path)
	if err != nil {
		return fmt.Errorf("Network key not found: %w", err)
	}

	fmt.Println("Mapped Network Drives")
	fmt.Println("=====================")
	fmt.Println()

	for _, driveKey := range key.Subkeys() {
		driveLetter := driveKey.Name()
		remotePath := ""
		userName := ""

		for _, val := range driveKey.Values() {
			switch val.Name() {
			case "RemotePath":
				remotePath = GetValueString(val)
			case "UserName":
				userName = GetValueString(val)
			}
		}

		fmt.Printf("Drive %s:\n", driveLetter)
		if remotePath != "" {
			fmt.Printf("  Remote Path: %s\n", remotePath)
		}
		if userName != "" {
			fmt.Printf("  User Name: %s\n", userName)
		}
		fmt.Printf("  Last Modified: %s\n", driveKey.Timestamp().Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}
