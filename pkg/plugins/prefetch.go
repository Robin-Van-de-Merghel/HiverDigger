package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&PrefetchPlugin{})
}

// PrefetchPlugin displays prefetch configuration from SYSTEM hive.
type PrefetchPlugin struct{}

func (p *PrefetchPlugin) Name() string {
	return "prefetch"
}

func (p *PrefetchPlugin) Description() string {
	return "Display prefetch configuration from SYSTEM hive"
}

func (p *PrefetchPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *PrefetchPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"ControlSet001\\Control\\Session Manager\\Memory Management\\PrefetchParameters",
		"ControlSet002\\Control\\Session Manager\\Memory Management\\PrefetchParameters",
	}

	fmt.Println("Prefetch Configuration")
	fmt.Println("======================")
	fmt.Println()

	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		fmt.Printf("[%s]\n", path)
		fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		for _, val := range key.Values() {
			if val.Name() == "EnablePrefetcher" || val.Name() == "EnableSuperfetch" {
				data := val.Bytes()
				if len(data) >= 4 {
					value := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
					fmt.Printf("  %s: %d\n", val.Name(), value)
				}
			}
		}
		fmt.Println()

		return nil
	}

	return fmt.Errorf("prefetch key not found")
}
