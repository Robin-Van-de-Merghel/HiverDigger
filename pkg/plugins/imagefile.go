package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ImageFilePlugin{})
}

// ImageFilePlugin extracts Image File Execution Options (IFEO) - used for debugging and malware persistence
type ImageFilePlugin struct{}

func (p *ImageFilePlugin) Name() string {
	return "imagefile"
}

func (p *ImageFilePlugin) Description() string {
	return "Image File Execution Options (IFEO) - SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Image File Execution Options"
}

func (p *ImageFilePlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *ImageFilePlugin) Run(hive *regf.Hive) error {
	ifeoPath := "Microsoft\\Windows NT\\CurrentVersion\\Image File Execution Options"
	ifeoKey, err := hive.GetKey(ifeoPath)
	if err != nil {
		return fmt.Errorf("IFEO key not found: %w", err)
	}

	fmt.Println("Image File Execution Options (IFEO):")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Note: Debugger entries can indicate malware persistence or legitimate debugging")
	fmt.Println()

	for _, exe := range ifeoKey.Subkeys() {
		hasDebugger := false
		debuggerVal := ""
		
		for _, val := range exe.Values() {
			if val.Name() == "Debugger" {
				hasDebugger = true
				debuggerVal = GetValueString(val)
			}
		}
		
		if hasDebugger {
			fmt.Printf("\n[%s] %s\n", exe.Timestamp().Format("2006-01-02 15:04:05"), exe.Name())
			fmt.Printf("  Debugger: %s\n", debuggerVal)
			
			for _, val := range exe.Values() {
				if val.Name() != "Debugger" && val.Name() != "" {
					fmt.Printf("  %s: %s\n", val.Name(), GetValueString(val))
				}
			}
		}
	}

	return nil
}
