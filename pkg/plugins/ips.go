package plugins

import (
	"fmt"
	"strings"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&IPSPlugin{})
}

// IPSPlugin implements the functionality of keydet89's ips.pl.
// It locates the current ControlSet and reads TCP/IP interface configuration.
type IPSPlugin struct{}

func (p *IPSPlugin) Name() string {
	return "ips"
}

func (p *IPSPlugin) Description() string {
	return "Extract IP configuration from SYSTEM hive (similar to keydet89's ips.pl)"
}

func (p *IPSPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *IPSPlugin) Run(hive *regf.Hive) error {
	// Step 1: Find current ControlSet
	controlSetName, err := p.findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	fmt.Printf("Current ControlSet: %s\n\n", controlSetName)

	// Step 2: Navigate to Services\Tcpip\Parameters\Interfaces
	interfacesPath := fmt.Sprintf("%s\\Services\\Tcpip\\Parameters\\Interfaces", controlSetName)
	interfacesKey, err := hive.GetKey(interfacesPath)
	if err != nil {
		return fmt.Errorf("failed to find Interfaces key: %w", err)
	}

	// Step 3: Iterate through interface subkeys
	subkeys := interfacesKey.Subkeys()
	if len(subkeys) == 0 {
		fmt.Println("No network interfaces found.")
		return nil
	}

	for _, ifaceKey := range subkeys {
		p.printInterface(ifaceKey)
	}

	return nil
}

// findCurrentControlSet determines the current ControlSet by reading Select\Current.
func (p *IPSPlugin) findCurrentControlSet(hive *regf.Hive) (string, error) {
	selectKey, err := hive.GetKey("Select")
	if err != nil {
		return "", err
	}

	// Find the "Current" value
	values := selectKey.Values()
	for _, v := range values {
		if strings.EqualFold(v.Name(), "Current") {
			// Current is a REG_DWORD (4)
			if v.Type() == 4 {
				data := v.Bytes()
				if len(data) >= 4 {
					// Read as little-endian DWORD
					currentNum := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
					return fmt.Sprintf("ControlSet%03d", currentNum), nil
				}
			}
		}
	}

	return "", fmt.Errorf("Current value not found in Select key")
}

// printInterface prints IP configuration for a single interface.
func (p *IPSPlugin) printInterface(ifaceKey *regf.Key) {
	fmt.Printf("Interface: %s\n", ifaceKey.Name())

	values := ifaceKey.Values()
	if len(values) == 0 {
		fmt.Println("  (no values)")
		fmt.Println()
		return
	}

	// Collect interesting values
	dhcpIPAddress := ""
	dhcpDomain := ""
	dhcpNetworkHint := ""
	ipAddress := ""
	domain := ""

	for _, v := range values {
		name := v.Name()
		switch {
		case strings.EqualFold(name, "DhcpIPAddress"):
			dhcpIPAddress = GetValueString(v)
		case strings.EqualFold(name, "DhcpDomain"):
			dhcpDomain = GetValueString(v)
		case strings.EqualFold(name, "DhcpNetworkHint"):
			dhcpNetworkHint = GetValueString(v)
		case strings.EqualFold(name, "IPAddress"):
			ipAddress = GetValueString(v)
		case strings.EqualFold(name, "Domain"):
			domain = GetValueString(v)
		}
	}

	// Print values in the style of ips.pl
	if dhcpIPAddress != "" {
		fmt.Printf("  DhcpIPAddress    : %s\n", dhcpIPAddress)
	}
	if dhcpDomain != "" {
		fmt.Printf("  DhcpDomain       : %s\n", dhcpDomain)
	}
	if dhcpNetworkHint != "" {
		fmt.Printf("  DhcpNetworkHint  : %s\n", dhcpNetworkHint)
	}
	if ipAddress != "" {
		fmt.Printf("  IPAddress        : %s\n", ipAddress)
	}
	if domain != "" {
		fmt.Printf("  Domain           : %s\n", domain)
	}

	fmt.Println()
}
