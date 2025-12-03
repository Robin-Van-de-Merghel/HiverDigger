package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/plugins"
	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func main() {
	var hivePath string
	var pluginName string
	var listPlugins bool

	flag.StringVar(&hivePath, "hive", "", "Path to registry hive file")
	flag.StringVar(&pluginName, "plugin", "", "Plugin to run")
	flag.BoolVar(&listPlugins, "list", false, "List available plugins")
	flag.Parse()

	if listPlugins {
		printAvailablePlugins()
		return
	}

	if hivePath == "" {
		fmt.Fprintf(os.Stderr, "Error: -hive flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if pluginName == "" {
		fmt.Fprintf(os.Stderr, "Error: -plugin flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Open the hive file
	hive, err := regf.OpenFile(hivePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening hive: %v\n", err)
		os.Exit(1)
	}
	defer hive.Close()

	// Get the plugin
	plugin, err := plugins.Get(pluginName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nAvailable plugins:\n")
		printAvailablePlugins()
		os.Exit(1)
	}

	// Run the plugin
	if err := plugin.Run(hive); err != nil {
		fmt.Fprintf(os.Stderr, "Plugin execution failed: %v\n", err)
		os.Exit(1)
	}
}

func printAvailablePlugins() {
	pluginNames := plugins.List()
	if len(pluginNames) == 0 {
		fmt.Println("No plugins registered.")
		return
	}

	fmt.Println("Available plugins:")
	for _, name := range pluginNames {
		plugin, err := plugins.Get(name)
		if err != nil {
			fmt.Printf("  %-15s (error: %v)\n", name, err)
			continue
		}
		fmt.Printf("  %-15s %s\n", name, plugin.Description())
	}
}
