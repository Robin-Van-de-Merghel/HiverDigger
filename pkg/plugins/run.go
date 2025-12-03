package plugins

import (
	"fmt"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&RunPlugin{})
}

// RunPlugin lists programs that run at startup from SOFTWARE hive.
type RunPlugin struct{}

func (p *RunPlugin) Name() string {
	return "run"
}

func (p *RunPlugin) Description() string {
	return "List programs that run at startup (Run/RunOnce keys)"
}

func (p *RunPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *RunPlugin) Run(hive *regf.Hive) error {
	fmt.Println("Startup Programs")
	fmt.Println("================")
	fmt.Println()

	paths := []string{
		"Microsoft\\Windows\\CurrentVersion\\Run",
		"Microsoft\\Windows\\CurrentVersion\\RunOnce",
		"Microsoft\\Windows\\CurrentVersion\\RunServices",
		"Microsoft\\Windows\\CurrentVersion\\RunServicesOnce",
		"Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\Run",
		"Wow6432Node\\Microsoft\\Windows\\CurrentVersion\\RunOnce",
	}

	for _, path := range paths {
		if err := p.printRunKeys(hive, path); err != nil {
			continue
		}
	}

	return nil
}

func (p *RunPlugin) printRunKeys(hive *regf.Hive, path string) error {
	key, err := hive.GetKey(path)
	if err != nil {
		return err
	}

	fmt.Printf("[%s]\n", path)
	fmt.Printf("Last Modified: %s\n", key.Timestamp().Format("2006-01-02 15:04:05"))

	for _, val := range key.Values() {
		if val.Name() != "" {
			fmt.Printf("  %s = %s\n", val.Name(), GetValueString(val))
		}
	}
	fmt.Println()

	return nil
}
