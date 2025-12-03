package plugins

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&ServicesPlugin{})
}

// ServicesPlugin lists Windows services from SYSTEM hive.
type ServicesPlugin struct{}

func (p *ServicesPlugin) Name() string {
	return "services"
}

func (p *ServicesPlugin) Description() string {
	return "List Windows services from SYSTEM hive"
}

func (p *ServicesPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *ServicesPlugin) Run(hive *regf.Hive) error {
	// Find current ControlSet
	controlSet, err := p.findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	fmt.Printf("Windows Services (from %s)\n", controlSet)
	fmt.Println("==================================")
	fmt.Println()

	servicesPath := fmt.Sprintf("%s\\Services", controlSet)
	servicesKey, err := hive.GetKey(servicesPath)
	if err != nil {
		return fmt.Errorf("failed to find Services key: %w", err)
	}

	for _, svcKey := range servicesKey.Subkeys() {
		p.printService(svcKey)
	}

	return nil
}

func (p *ServicesPlugin) findCurrentControlSet(hive *regf.Hive) (string, error) {
	selectKey, err := hive.GetKey("Select")
	if err != nil {
		return "", err
	}

	for _, v := range selectKey.Values() {
		if strings.EqualFold(v.Name(), "Current") {
			if v.Type() == 4 { // REG_DWORD
				data := v.Bytes()
				if len(data) >= 4 {
					currentNum := binary.LittleEndian.Uint32(data)
					return fmt.Sprintf("ControlSet%03d", currentNum), nil
				}
			}
		}
	}

	return "", fmt.Errorf("Current value not found")
}

func (p *ServicesPlugin) printService(svcKey *regf.Key) {
	displayName := ""
	imagePath := ""
	startType := ""
	serviceType := ""

	for _, val := range svcKey.Values() {
		name := val.Name()
		switch {
		case strings.EqualFold(name, "DisplayName"):
			displayName = GetValueString(val)
		case strings.EqualFold(name, "ImagePath"):
			imagePath = GetValueString(val)
		case strings.EqualFold(name, "Start"):
			if val.Type() == 4 {
				data := val.Bytes()
				if len(data) >= 4 {
					start := binary.LittleEndian.Uint32(data)
					switch start {
					case 0:
						startType = "Boot"
					case 1:
						startType = "System"
					case 2:
						startType = "Automatic"
					case 3:
						startType = "Manual"
					case 4:
						startType = "Disabled"
					default:
						startType = fmt.Sprintf("Unknown (%d)", start)
					}
				}
			}
		case strings.EqualFold(name, "Type"):
			if val.Type() == 4 {
				data := val.Bytes()
				if len(data) >= 4 {
					svcType := binary.LittleEndian.Uint32(data)
					switch svcType & 0xFF {
					case 1:
						serviceType = "Kernel Driver"
					case 2:
						serviceType = "File System Driver"
					case 16:
						serviceType = "Win32 Own Process"
					case 32:
						serviceType = "Win32 Share Process"
					default:
						serviceType = fmt.Sprintf("Type 0x%x", svcType)
					}
				}
			}
		}
	}

	fmt.Printf("Service: %s\n", svcKey.Name())
	if displayName != "" {
		fmt.Printf("  Display Name: %s\n", displayName)
	}
	if imagePath != "" {
		fmt.Printf("  Image Path: %s\n", imagePath)
	}
	if startType != "" {
		fmt.Printf("  Start Type: %s\n", startType)
	}
	if serviceType != "" {
		fmt.Printf("  Service Type: %s\n", serviceType)
	}
	fmt.Printf("  Last Modified: %s\n", svcKey.Timestamp().Format("2006-01-02 15:04:05"))
	fmt.Println()
}
