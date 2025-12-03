package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/plugins"
	"github.com/Robin-Van-de-Merghel/HiveDigger/pkg/regf"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Hive represents a discovered hive file
type Hive struct {
	Path     string
	Name     string
	Type     string
	Size     int64
	ModTime  time.Time
	hiveData *regf.Hive
}

// Implement list.Item interface
func (h Hive) FilterValue() string { return h.Name }
func (h Hive) Title() string       { return h.Name }
func (h Hive) Description() string {
	return fmt.Sprintf("%s | %s | %.2f MB", h.Type, h.Path, float64(h.Size)/1024/1024)
}

type viewMode int

const (
	workflowSelectionMode viewMode = iota
	hiveBrowserMode
	pluginBrowserMode          // New: Browse plugins first
	hiveSelectionForPluginMode // New: Select hive(s) after plugin selection
	pluginSelectorMode
	resultViewMode
)

type workflowType int

const (
	fileFirstWorkflow workflowType = iota
	pluginFirstWorkflow
)

type model struct {
	mode             viewMode
	workflow         workflowType
	hives            []Hive
	hiveList         list.Model
	pluginList       list.Model
	workflowList     list.Model  // New: workflow selection list
	allPluginItems   []list.Item // Store all plugins
	viewport         viewport.Model
	searchInput      textinput.Model
	selectedHive     *Hive
	selectedPlugin   string
	result           string
	scanning         bool
	scanPath         string
	ready            bool
	width            int
	height           int
	err              error
	filterByHiveType bool // Toggle for filtering plugins by hive type
}

type scanCompleteMsg struct {
	hives []Hive
	err   error
}

type pluginResultMsg struct {
	result string
	err    error
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			PaddingTop(1).
			PaddingBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
)

func initialModel() model {
	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search path (e.g., ./example)"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// Initialize hive list
	hiveDelegate := list.NewDefaultDelegate()
	hiveList := list.New([]list.Item{}, hiveDelegate, 0, 0)
	hiveList.Title = "Registry Hives"
	hiveList.SetShowStatusBar(true)
	hiveList.SetFilteringEnabled(true)

	// Initialize plugin list
	pluginItems := make([]list.Item, 0)
	for _, name := range plugins.List() {
		p, _ := plugins.Get(name)
		if p != nil {
			pluginItems = append(pluginItems, pluginItem{name: name, desc: p.Description()})
		}
	}

	pluginDelegate := list.NewDefaultDelegate()
	pluginList := list.New(pluginItems, pluginDelegate, 0, 0)
	pluginList.Title = "Available Plugins"
	pluginList.SetFilteringEnabled(true)

	// Initialize workflow selection list
	workflowItems := []list.Item{
		workflowItem{name: "File-First Workflow", desc: "Select a hive file first, then choose compatible plugins"},
		workflowItem{name: "Plugin-First Workflow", desc: "Select a plugin first, then choose compatible hive files"},
	}
	workflowDelegate := list.NewDefaultDelegate()
	workflowList := list.New(workflowItems, workflowDelegate, 0, 0)
	workflowList.Title = "Select Analysis Workflow"
	workflowList.SetShowStatusBar(false)

	// Initialize viewport for results
	vp := viewport.New(80, 20)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "." // Fallback to current directory
	}

	return model{
		mode:             workflowSelectionMode,
		hiveList:         hiveList,
		pluginList:       pluginList,
		workflowList:     workflowList,
		allPluginItems:   pluginItems, // Store all plugins
		viewport:         vp,
		searchInput:      ti,
		scanPath:         cwd,
		filterByHiveType: true, // Enable filtering by default
	}
}

type workflowItem struct {
	name string
	desc string
}

func (w workflowItem) FilterValue() string { return w.name }
func (w workflowItem) Title() string       { return w.name }
func (w workflowItem) Description() string { return w.desc }

type pluginItem struct {
	name string
	desc string
}

func (p pluginItem) FilterValue() string { return p.name }
func (p pluginItem) Title() string       { return p.name }
func (p pluginItem) Description() string { return p.desc }

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		scanForHives(m.scanPath),
	)
}

func scanForHives(searchPath string) tea.Cmd {
	return func() tea.Msg {
		var hives []Hive

		err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Continue on error
			}

			if d.IsDir() {
				return nil
			}

			// Check if filename looks like a registry hive
			name := strings.ToUpper(d.Name())
			isHive := false
			hiveType := "Unknown"

			switch {
			case name == "SYSTEM" || strings.HasPrefix(name, "SYSTEM."):
				isHive = true
				hiveType = "SYSTEM"
			case name == "SOFTWARE" || strings.HasPrefix(name, "SOFTWARE."):
				isHive = true
				hiveType = "SOFTWARE"
			case name == "SAM" || strings.HasPrefix(name, "SAM."):
				isHive = true
				hiveType = "SAM"
			case name == "SECURITY" || strings.HasPrefix(name, "SECURITY."):
				isHive = true
				hiveType = "SECURITY"
			case name == "NTUSER.DAT" || strings.HasPrefix(name, "NTUSER.DAT"):
				isHive = true
				hiveType = "NTUSER.DAT"
			case name == "USRCLASS.DAT" || strings.HasPrefix(name, "USRCLASS.DAT"):
				isHive = true
				hiveType = "USRCLASS.DAT"
			case strings.HasSuffix(name, ".HIVE"):
				isHive = true
				hiveType = "Custom"
			}

			if isHive {
				info, err := d.Info()
				if err != nil {
					return nil
				}

				// Quick validation - check for regf signature
				f, err := os.Open(path)
				if err != nil {
					return nil
				}
				defer func() {
					if err := f.Close(); err != nil {
						fmt.Printf("failed to close file %q: %v", path, err)
					}
				}()

				sig := make([]byte, 4)
				if n, err := f.Read(sig); err != nil || n != 4 || string(sig) != "regf" {
					return nil // Not a valid hive
				}

				hives = append(hives, Hive{
					Path:    path,
					Name:    filepath.Base(path),
					Type:    hiveType,
					Size:    info.Size(),
					ModTime: info.ModTime(),
				})
			}

			return nil
		})

		return scanCompleteMsg{hives: hives, err: err}
	}
}

func runPlugin(hive *Hive, pluginName string) tea.Cmd {
	return func() tea.Msg {
		// Open hive if not already open
		if hive.hiveData == nil {
			h, err := regf.OpenFile(hive.Path)
			if err != nil {
				return pluginResultMsg{err: fmt.Errorf("failed to open hive: %w", err)}
			}
			hive.hiveData = h
		}

		// Get plugin
		plugin, err := plugins.Get(pluginName)
		if err != nil {
			return pluginResultMsg{err: err}
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run plugin
		err = plugin.Run(hive.hiveData)

		writeError := w.Close()
		if writeError != nil {
			return pluginResultMsg{result: "", err: writeError}
		}

		os.Stdout = oldStdout

		var result strings.Builder
		buf := make([]byte, 1024)
		for {
			n, readErr := r.Read(buf)
			if n > 0 {
				result.Write(buf[:n])
			}
			if readErr != nil {
				break
			}
		}

		if err != nil {
			return pluginResultMsg{result: result.String(), err: err}
		}

		return pluginResultMsg{result: result.String()}
	}
}

// updatePluginList updates the plugin list based on current filter settings
func (m model) updatePluginList() model {
	var filteredItems []list.Item

	if m.filterByHiveType && m.selectedHive != nil {
		// Filter plugins by hive type
		compatibleNames := plugins.ListForHiveType(m.selectedHive.Type)
		compatibleMap := make(map[string]bool)
		for _, name := range compatibleNames {
			compatibleMap[name] = true
		}

		for _, item := range m.allPluginItems {
			if pItem, ok := item.(pluginItem); ok {
				if compatibleMap[pItem.name] {
					filteredItems = append(filteredItems, item)
				}
			}
		}
	} else {
		// Show all plugins
		filteredItems = m.allPluginItems
	}

	m.pluginList.SetItems(filteredItems)

	// Update title to show filter status
	if m.filterByHiveType && m.selectedHive != nil {
		m.pluginList.Title = fmt.Sprintf("Plugins for %s (Filtered - %d/%d)",
			m.selectedHive.Type, len(filteredItems), len(m.allPluginItems))
	} else {
		m.pluginList.Title = fmt.Sprintf("Available Plugins (All - %d)", len(filteredItems))
	}

	return m
}

// updateHiveListForPlugin updates the hive list to show only compatible hives for selected plugin
func (m model) updateHiveListForPlugin() model {
	if m.selectedPlugin == "" {
		return m
	}

	// Find compatible hives for this plugin
	compatibleHives := make([]list.Item, 0)
	for _, hive := range m.hives {
		if plugins.IsCompatibleWithHiveType(m.selectedPlugin, hive.Type) {
			compatibleHives = append(compatibleHives, hive)
		}
	}

	m.hiveList.SetItems(compatibleHives)
	m.hiveList.Title = fmt.Sprintf("Compatible Hives for '%s' (%d/%d)",
		m.selectedPlugin, len(compatibleHives), len(m.hives))

	return m
}

// resetLists resets hive and plugin lists to their original state
func (m model) resetLists() model {
	// Reset hive list to show all hives
	items := make([]list.Item, len(m.hives))
	for i, h := range m.hives {
		items[i] = h
	}
	m.hiveList.SetItems(items)
	m.hiveList.Title = "Registry Hives"

	// Reset plugin list to show all plugins
	m.pluginList.SetItems(m.allPluginItems)
	m.pluginList.Title = "Available Plugins"

	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.hiveList.SetSize(msg.Width, msg.Height-10)
			m.pluginList.SetSize(msg.Width, msg.Height-10)
			m.workflowList.SetSize(msg.Width, msg.Height-10)
			m.viewport = viewport.New(msg.Width-4, msg.Height-10)
			m.ready = true
		} else {
			m.hiveList.SetSize(msg.Width, msg.Height-10)
			m.pluginList.SetSize(msg.Width, msg.Height-10)
			m.workflowList.SetSize(msg.Width, msg.Height-10)
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 10
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Close all open hives
			for i := range m.hives {
				if m.hives[i].hiveData != nil {
					err := m.hives[i].hiveData.Close()
					if err != nil {
						fmt.Printf("Error closing hive %d: %v\n", i, err)
					}
				}
			}
			return m, tea.Quit

		case "b", "x":
			switch m.mode {
			case hiveBrowserMode:
				if m.workflow == fileFirstWorkflow {
					m.mode = workflowSelectionMode
					// Reset state when going back to workflow selection
					m.selectedHive = nil
					m.selectedPlugin = ""
					m = m.resetLists()
				}
			case pluginBrowserMode:
				if m.workflow == pluginFirstWorkflow {
					m.mode = workflowSelectionMode
					// Reset state when going back to workflow selection
					m.selectedHive = nil
					m.selectedPlugin = ""
					m = m.resetLists()
				}
			case pluginSelectorMode:
				if m.workflow == fileFirstWorkflow {
					m.mode = hiveBrowserMode
					m.selectedPlugin = ""
				}
			case hiveSelectionForPluginMode:
				if m.workflow == pluginFirstWorkflow {
					m.mode = pluginBrowserMode
					m.selectedHive = nil
				}
			case resultViewMode:
				if m.workflow == fileFirstWorkflow {
					m.mode = pluginSelectorMode
				} else {
					m.mode = hiveSelectionForPluginMode
				}
			}

		case "w":
			// Toggle whitelist filter
			if m.mode == pluginSelectorMode {
				m.filterByHiveType = !m.filterByHiveType
				m = m.updatePluginList()
			}

		case "enter":
			switch m.mode {
			case workflowSelectionMode:
				if i, ok := m.workflowList.SelectedItem().(workflowItem); ok {
					if i.name == "File-First Workflow" {
						m.workflow = fileFirstWorkflow
						m.mode = hiveBrowserMode
					} else {
						m.workflow = pluginFirstWorkflow
						m.mode = pluginBrowserMode
					}
				}
			case hiveBrowserMode:
				if i, ok := m.hiveList.SelectedItem().(Hive); ok {
					m.selectedHive = &i
					m.mode = pluginSelectorMode
					m = m.updatePluginList() // Update plugin list when hive is selected
				}
			case pluginBrowserMode:
				if i, ok := m.pluginList.SelectedItem().(pluginItem); ok {
					m.selectedPlugin = i.name
					m.mode = hiveSelectionForPluginMode
					m = m.updateHiveListForPlugin() // Update hive list when plugin is selected
				}
			case pluginSelectorMode:
				if i, ok := m.pluginList.SelectedItem().(pluginItem); ok {
					m.selectedPlugin = i.name
					m.mode = resultViewMode
					return m, runPlugin(m.selectedHive, m.selectedPlugin)
				}
			case hiveSelectionForPluginMode:
				if i, ok := m.hiveList.SelectedItem().(Hive); ok {
					m.selectedHive = &i
					m.mode = resultViewMode
					return m, runPlugin(m.selectedHive, m.selectedPlugin)
				}
			}
		}

	case scanCompleteMsg:
		m.scanning = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.hives = msg.hives
			items := make([]list.Item, len(msg.hives))
			for i, h := range msg.hives {
				items[i] = h
			}
			m.hiveList.SetItems(items)
		}

	case pluginResultMsg:
		if msg.err != nil {
			m.result = errorStyle.Render(fmt.Sprintf("Error: %v", msg.err))
		} else {
			m.result = msg.result
		}
		m.viewport.SetContent(m.result)
	}

	// Update the appropriate component based on mode
	switch m.mode {
	case workflowSelectionMode:
		m.workflowList, cmd = m.workflowList.Update(msg)
		cmds = append(cmds, cmd)
	case hiveBrowserMode, hiveSelectionForPluginMode:
		m.hiveList, cmd = m.hiveList.Update(msg)
		cmds = append(cmds, cmd)
	case pluginSelectorMode, pluginBrowserMode:
		m.pluginList, cmd = m.pluginList.Update(msg)
		cmds = append(cmds, cmd)
	case resultViewMode:
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var content string

	switch m.mode {
	case workflowSelectionMode:
		title := titleStyle.Render("HiveDigger - Select Analysis Workflow")
		help := helpStyle.Render("Enter: Select Workflow | q: Quit")

		content = fmt.Sprintf("%s\n\n%s\n\n%s",
			title,
			m.workflowList.View(),
			help,
		)

	case hiveBrowserMode:
		title := titleStyle.Render("HiveDigger - Registry Hive Browser")
		help := helpStyle.Render("Enter: Select | q: Quit | /: Filter | b: Go Back")
		status := ""
		if m.scanning {
			status = fmt.Sprintf("Scanning %s for hives...", m.scanPath)
		} else {
			status = fmt.Sprintf("Found %d hive(s) in %s", len(m.hives), m.scanPath)
		}

		content = fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
			title,
			m.hiveList.View(),
			successStyle.Render(status),
			help,
		)

	case pluginBrowserMode:
		title := titleStyle.Render("HiveDigger - Plugin Browser")
		help := helpStyle.Render("Enter: Select Plugin | b: Go Back | q: Quit | /: Search")

		content = fmt.Sprintf("%s\n\n%s\n\n%s",
			title,
			m.pluginList.View(),
			help,
		)

	case hiveSelectionForPluginMode:
		title := titleStyle.Render(fmt.Sprintf("HiveDigger - Select Hive for Plugin: %s", m.selectedPlugin))
		help := helpStyle.Render("Enter: Run Plugin | b: Go Back | q: Quit | /: Filter")

		content = fmt.Sprintf("%s\n\n%s\n\n%s",
			title,
			m.hiveList.View(),
			help,
		)

	case pluginSelectorMode:
		title := titleStyle.Render(fmt.Sprintf("HiveDigger - Select Plugin for: %s", m.selectedHive.Name))

		// Build help text with filter status
		filterStatus := ""
		if m.filterByHiveType {
			filterStatus = " | Filter: ON"
		} else {
			filterStatus = " | Filter: OFF"
		}
		help := helpStyle.Render("Enter: Run | b: Back | q: Quit | /: Search | w: Toggle Filter" + filterStatus)

		content = fmt.Sprintf("%s\n\n%s\n\n%s",
			title,
			m.pluginList.View(),
			help,
		)

	case resultViewMode:
		title := titleStyle.Render(fmt.Sprintf("Results: %s on %s", m.selectedPlugin, m.selectedHive.Name))
		help := helpStyle.Render("b: Back | q: Quit | ↑↓: Scroll")

		content = fmt.Sprintf("%s\n\n%s\n\n%s",
			title,
			m.viewport.View(),
			help,
		)
	}

	return content
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
