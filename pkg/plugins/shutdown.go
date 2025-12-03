package plugins

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ShutdownPlugin{})
}

// ShutdownPlugin displays shutdown information from SYSTEM hive.
type ShutdownPlugin struct{}

func (p *ShutdownPlugin) Name() string {
	return "shutdown"
}

func (p *ShutdownPlugin) Description() string {
	return "Display shutdown information from SYSTEM hive"
}

func (p *ShutdownPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *ShutdownPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Windows",
		"ControlSet002\\Control\\Windows",
	}

	fmt.Println("Shutdown Information")
	fmt.Println("====================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			name := val.Name()
			if strings.EqualFold(name, "ShutdownTime") {
				data := val.Bytes()
				if len(data) >= 8 {
					timestamp := binary.LittleEndian.Uint64(data)
					// This is a FILETIME, but we'll just show raw for now
					fmt.Printf("  ShutdownTime: 0x%x\n", timestamp)
				}
			}
		}
		fmt.Println()

		return nil
	}

	return fmt.Errorf("shutdown key not found")
}
