# HiveDigger

HiveDigger is **vibe-coded** a pure-Go parser for Windows Registry hive files (REGF format) with a plugin-based architecture for forensic analysis. It includes both a command-line interface and an interactive Terminal User Interface (TUI) with dual workflow modes.

> [!NOTE]
> As a vibe-coded project, it as some flaws and bug. You are welcome to contribute to make it more production ready!


## Why vibe coding?

During my cyber-security courses, I needed a tool to analyze hives. I found the super cool projects from [@Eric Zimmerman](https://ericzimmerman.github.io), but it was in pearl.

I recently started coding in *go*, so I thought it was the right time to adapt a program into go!

But as I had to have this project working quickly, I vibe-coded it. When I'll have some time, I'll correct as much as I can so that this project has a production ready structure!

## Features

- **Pure-Go Implementation**: Cross-platform support without cgo dependencies
- **Forensic-Friendly**: Exposes raw cell bytes and offsets for analysis
- **Plugin Architecture**: Extensible design with 40+ forensic plugins
- **Dual Workflow Modes**:
  - **File-First**: Select a hive file, then choose compatible plugins
  - **Plugin-First**: Select a plugin, then choose compatible hive files
- **Interactive TUI**: Browse hives and run plugins interactively with Bubble Tea
- **Smart Plugin Filtering**: Automatically filters plugins by hive type (toggleable)
- **Best-Effort Parsing**: Handles NK (key) and VK (value) cells with graceful degradation

## Building

### CLI Version

```bash
go build -o hivedigger ./cmd/hivedigger
```

### TUI Version (Recommended)

```bash
go build -o hivedigger-tui ./cmd/hivedigger-tui
```

Or install directly:

```bash
go install ./cmd/hivedigger
go install ./cmd/hivedigger-tui
```

## Usage

### Interactive TUI (Terminal User Interface)

The TUI provides an interactive way to browse hive files and run plugins with two workflow modes:

```bash
./hivedigger-tui
```

**New: Dual Workflow System**

1. **File-First Workflow** (Traditional):
   - Browse and select a hive file
   - View compatible plugins for that hive
   - Run plugin on selected hive

2. **Plugin-First Workflow** (New):
   - Browse and select a plugin
   - View compatible hive files for that plugin
   - Run plugin on selected hive

**Features:**
- **Workflow Selection Menu**: Choose your preferred analysis workflow on startup
- Automatically scans for registry hives recursively in the current working directory
- Browse discovered hives with arrow keys
- **Smart Plugin Filtering**: Toggle with `w` key to show only compatible plugins for selected hive type
  - When enabled (default): Shows only plugins designed for the selected hive (e.g., only SYSTEM plugins for SYSTEM hive)
  - When disabled: Shows all plugins (useful for renamed hive files)
- View plugin output in a scrollable viewport
- Filter hives and plugins with `/`
- Navigate: Enter (select), b or x (back), q (quit), w (toggle filter)

### Command-Line Interface

#### List Available Plugins

```bash
./hivedigger -list
```

#### Run a Plugin

```bash
./hivedigger -hive <path-to-hive-file> -plugin <plugin-name>
```

Example:

```bash
./hivedigger -hive example/config/SYSTEM -plugin ips
./hivedigger -hive example/config/SOFTWARE -plugin listsoft
./hivedigger -hive example/config/SYSTEM -plugin services
```

## Available Plugins

HiveDigger includes 40+ plugins adapted from RegRipper for forensic analysis:

### SYSTEM Hive Plugins (16)

- **ips**: Extract IP configuration from TCP/IP interfaces
- **services**: List Windows services with start type and image path
- **compname**: Display computer name
- **timezone**: Display timezone information
- **usbdevices**: List USB devices from USBSTOR and USB keys
- **mountpoints**: Display mounted devices and volumes
- **environment**: Display system environment variables
- **shutdown**: Display shutdown information
- **prefetch**: Display prefetch configuration
- **bootexecute**: Display boot execute commands
- **sessionmgr**: Display Session Manager information
- **knowndlls**: Display KnownDLLs
- **rdp**: Display Terminal Server/RDP configuration
- **printers**: Display installed printers
- **shimcache**: Display Application Compatibility Cache (ShimCache) entries
- **bam**: Display Background Activity Moderator (BAM) entries (Windows 10+)

### SOFTWARE Hive Plugins (14)

- **listsoft**: List installed software from Uninstall keys
- **uninstall**: Comprehensive list of installed/uninstalled programs
- **run**: List programs that run at startup (Run/RunOnce keys)
- **autorun**: List autorun locations
- **networklist**: List network profiles
- **winver**: Display Windows version information
- **apppaths**: Display App Paths
- **fileassoc**: Display file associations
- **appinit**: Display AppInit DLLs
- **bho**: Display Browser Helper Objects
- **winlogon**: Display Winlogon information
- **activesetup**: Display Active Setup components
- **tasks**: Display scheduled tasks information

### NTUSER.DAT / USRCLASS.DAT Hive Plugins (9)

- **userassist**: Display UserAssist data (program execution)
- **recentdocs**: Display recently opened documents
- **typedurls**: Display typed URLs from Internet Explorer
- **runmru**: Display Run dialog history
- **typedpaths**: Display typed paths from Windows Explorer
- **wordwheel**: Display Windows search terms
- **shellbags**: Display ShellBags data (folder access history)
- **mapnetdrive**: Display mapped network drives
- **muicache**: Display MUICache entries (executed applications)
- **appcompat**: Display Application Compatibility flags

### SAM Hive Plugins (1)

- **samusers**: List local users with RIDs

### Special Hive Plugins (1)

- **amcache**: Display AmCache entries (program execution artifacts) from AmCache.hve

## Testing

### Running Tests

```bash
go test ./...
```

### Adding Test Fixtures

To add test hive fixtures for more comprehensive testing:

1. Place registry hive files in a `testdata` directory within the relevant package
2. Keep test hives small (< 1MB) to avoid repository bloat
3. Add tests in `*_test.go` files that reference these fixtures

Example test structure:
```
pkg/regf/
  ├── regf.go
  ├── regf_test.go
  └── testdata/
      └── minimal.hive
```

## Architecture

### Parser Design

The REGF parser is designed for forensic analysis:

- **Best-Effort Parsing**: Gracefully handles malformed hives
- **Raw Access**: `RawCellAt()` and `IterateCells()` for low-level analysis
- **Memory-Based**: Loads entire hive into memory for performance

### Plugin System

Plugins are compiled Go code implementing the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Description() string
    Run(hive *regf.Hive) error
}
```

To add a new plugin:

1. Create a new file in `pkg/plugins/`
2. Implement the `Plugin` interface
3. Register it in an `init()` function

## Limitations

This is an initial implementation with the following limitations:

- **Limited Cell Types**: Currently parses NK (key) and VK (value) cells
- **Basic Structure**: Does not handle all REGF edge cases
- **No Security**: Does not parse SK (security) cells
- **No Class Names**: Does not parse class name data

## Future Development

Potential enhancements:

- [ ] Expand cell type support (SK, DB, etc.)
- [ ] Add more comprehensive unit tests
- [ ] Improve Unicode handling
- [ ] Add support for deleted key recovery
- [ ] Consider JSON-based plugin DSL vs compiled Go plugins
- [ ] Add transaction log (LOG1/LOG2) support
- [ ] Implement lazy loading for large hives

## References

- [Windows Registry File Format Specification](https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md)
- [RegRipper by keydet89](https://github.com/keydet89/RegRipper3.0)
