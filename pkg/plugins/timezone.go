package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&TimeZonePlugin{})
}

// TimeZonePlugin displays timezone information from SYSTEM hive.
type TimeZonePlugin struct{}

func (p *TimeZonePlugin) Name() string {
	return "timezone"
}

func (p *TimeZonePlugin) Description() string {
	return "Display timezone information from SYSTEM hive"
}

func (p *TimeZonePlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *TimeZonePlugin) Run(hive *regf.Hive) error {
	// Find current ControlSet
	selectKey, err := hive.GetKey("Select")
	if err != nil {
		return err
	}

	var currentNum uint32 = 1
	for _, v := range selectKey.Values() {
		if strings.EqualFold(v.Name(), "Current") && v.Type() == 4 {
			data := v.Bytes()
			if len(data) >= 4 {
				currentNum = uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
			}
		}
	}

	controlSet := fmt.Sprintf("ControlSet%03d", currentNum)
	tzPath := fmt.Sprintf("%s\\Control\\TimeZoneInformation", controlSet)

	tzKey, err := hive.GetKey(tzPath)
	if err != nil {
		return fmt.Errorf("timezone key not found: %w", err)
	}

	fmt.Println("Timezone Information")
	fmt.Println("====================")
	fmt.Println()

	for _, val := range tzKey.Values() {
		name := val.Name()
		switch {
		case strings.EqualFold(name, "TimeZoneKeyName"):
			fmt.Printf("Timezone: %s\n", GetValueString(val))
		case strings.EqualFold(name, "StandardName"):
			fmt.Printf("Standard Name: %s\n", GetValueString(val))
		case strings.EqualFold(name, "DaylightName"):
			fmt.Printf("Daylight Name: %s\n", GetValueString(val))
		}
	}

	fmt.Printf("Last Modified: %s\n", tzKey.Timestamp().Format("2006-01-02 15:04:05"))

	return nil
}
