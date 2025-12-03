package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&TasksPlugin{})
}

// TasksPlugin displays scheduled tasks from SOFTWARE hive.
type TasksPlugin struct{}

func (p *TasksPlugin) Name() string {
	return "tasks"
}

func (p *TasksPlugin) Description() string {
	return "Display scheduled tasks information from SOFTWARE hive"
}

func (p *TasksPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *TasksPlugin) Run(hive *regf.Hive) error {
	paths := []string{
		"Microsoft/Windows NT/CurrentVersion/Schedule/TaskCache/Tasks",
		"Microsoft/Windows NT/CurrentVersion/Schedule/TaskCache/Tree",
	}

	fmt.Println("Scheduled Tasks:")
	fmt.Println(strings.Repeat("=", 80))

	found := false
	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		found = true
		fmt.Printf("\n%s:\n", path)
		fmt.Printf("Last Write: %s\n\n", key.Timestamp().Format("2006-01-02 15:04:05"))

		// Enumerate task GUIDs or names
		for _, subkey := range key.Subkeys() {
			fmt.Printf("  Task: %s\n", subkey.Name())
			fmt.Printf("  Last Write: %s\n", subkey.Timestamp().Format("2006-01-02 15:04:05"))

			for _, val := range subkey.Values() {
				valStr := GetValueString(val)
				switch val.Name() {
				case "Path":
					fmt.Printf("    Path: %s\n", valStr)
				case "Author":
					fmt.Printf("    Author: %s\n", valStr)
				case "URI":
					fmt.Printf("    URI: %s\n", valStr)
				case "Actions":
					if len(val.Bytes()) > 0 {
						fmt.Printf("    Actions: (binary data, %d bytes)\n", len(val.Bytes()))
					}
				}
			}
			fmt.Println()
		}
	}

	if !found {
		fmt.Println("No scheduled tasks found")
	}

	return nil
}
