package plugins

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
)

func init() {
	Register(&BAMPlugin{})
}

// BAMPlugin displays Background Activity Moderator (BAM) entries.
// Shows program execution timestamps (Windows 10+).
type BAMPlugin struct{}

func (p *BAMPlugin) Name() string {
	return "bam"
}

func (p *BAMPlugin) Description() string {
	return "Display Background Activity Moderator (BAM) entries from SYSTEM hive"
}

func (p *BAMPlugin) CompatibleHiveTypes() []string {
	return []string{"SYSTEM"}
}

func (p *BAMPlugin) Run(hive *regf.Hive) error {
	// Find current ControlSet
	controlSetName, err := p.findCurrentControlSet(hive)
	if err != nil {
		return fmt.Errorf("failed to find current ControlSet: %w", err)
	}

	paths := []string{
		fmt.Sprintf("%s/Services/bam/State/UserSettings", controlSetName),
		fmt.Sprintf("%s/Services/dam/State/UserSettings", controlSetName),
	}

	fmt.Println("BAM/DAM Entries (Program Execution):")
	fmt.Println(strings.Repeat("=", 80))

	found := false
	for _, path := range paths {
		key, err := hive.GetKey(path)
		if err != nil {
			continue
		}

		found = true
		fmt.Printf("\n%s:\n", path)

		// Enumerate subkeys (SIDs)
		for _, sidKey := range key.Subkeys() {
			fmt.Printf("\n  SID: %s\n", sidKey.Name())
			fmt.Printf("  Last Write: %s\n", sidKey.Timestamp().Format("2006-01-02 15:04:05"))

			// List values (executables)
			for _, val := range sidKey.Values() {
				if val.Name() != "" && len(val.Bytes()) >= 8 {
					// Parse FILETIME from first 8 bytes
					data := val.Bytes()
					timestamp := binary.LittleEndian.Uint64(data[0:8])
					if timestamp > 0 {
						// Convert Windows FILETIME to Unix time
						t := filetimeToTime(timestamp)
						if t.Year() > 1970 {
							fmt.Printf("    %s\n", val.Name())
							fmt.Printf("      Timestamp: %s\n", t.Format("2006-01-02 15:04:05"))
						}
					}
				}
			}
		}
	}

	if !found {
		fmt.Println("No BAM/DAM entries found (Windows 10+ feature)")
	}

	return nil
}

// filetimeToTime converts a Windows FILETIME to time.Time
func filetimeToTime(filetime uint64) time.Time {
	if filetime == 0 {
		return time.Time{}
	}
	// Windows FILETIME is 100-nanosecond intervals since January 1, 1601
	// Unix time is seconds since January 1, 1970
	// Difference between 1601 and 1970 is 11644473600 seconds
	unixTime := int64(filetime/10000000 - 11644473600)
	if unixTime < 0 {
		return time.Time{}
	}
	return time.Unix(unixTime, 0)
}

func (p *BAMPlugin) findCurrentControlSet(hive *regf.Hive) (string, error) {
	selectKey, err := hive.GetKey("Select")
	if err != nil {
		return "", err
	}

	for _, val := range selectKey.Values() {
		if val.Name() == "Current" && len(val.Bytes()) >= 4 {
			current := binary.LittleEndian.Uint32(val.Bytes()[:4])
			return fmt.Sprintf("ControlSet%03d", current), nil
		}
	}

	return "", fmt.Errorf("current value not found in select key")
}
