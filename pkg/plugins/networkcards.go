package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&NetworkCardsPlugin{})
}

// NetworkCardsPlugin extracts network adapter details
type NetworkCardsPlugin struct{}

func (p *NetworkCardsPlugin) Name() string {
	return "networkcards"
}

func (p *NetworkCardsPlugin) Description() string {
	return "Network adapter details - SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\NetworkCards"
}

func (p *NetworkCardsPlugin) CompatibleHiveTypes() []string {
	return []string{"SOFTWARE"}
}

func (p *NetworkCardsPlugin) Run(hive *regf.Hive) error {
	netcardsPath := "Microsoft\\Windows NT\\CurrentVersion\\NetworkCards"
	netcardsKey, err := hive.GetKey(netcardsPath)
	if err != nil {
		return fmt.Errorf("NetworkCards key not found: %w", err)
	}

	fmt.Println("Network Adapters:")
	fmt.Println(strings.Repeat("=", 80))

	for _, card := range netcardsKey.Subkeys() {
		fmt.Printf("\nAdapter %s:\n", card.Name())
		for _, val := range card.Values() {
			fmt.Printf("  %s: %s\n", val.Name(), GetValueString(val))
		}
	}

	return nil
}
