package plugins

import (
	"testing"
)

func TestPluginRegistry(t *testing.T) {
	// Get list of all plugins
	plugins := List()

	if len(plugins) == 0 {
		t.Error("no plugins registered")
	}

	t.Logf("Found %d registered plugins", len(plugins))

	// Try to get each plugin
	for _, name := range plugins {
		plugin, err := Get(name)
		if err != nil {
			t.Errorf("failed to get plugin %q: %v", name, err)
			continue
		}

		if plugin == nil {
			t.Errorf("plugin %q is nil", name)
			continue
		}

		// Verify plugin has name and description
		if plugin.Name() == "" {
			t.Errorf("plugin %q has empty name", name)
		}

		if plugin.Description() == "" {
			t.Errorf("plugin %q has empty description", name)
		}

		t.Logf("  - %s: %s", plugin.Name(), plugin.Description())
	}
}

func TestGetValueString(t *testing.T) {
	// This is a basic test - more comprehensive tests would need actual registry data
	if GetValueString(nil) != "" {
		t.Error("GetValueString(nil) should return empty string")
	}
}

func TestPluginNotFound(t *testing.T) {
	_, err := Get("nonexistent-plugin")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestPluginCount(t *testing.T) {
	plugins := List()
	if len(plugins) < 30 {
		t.Errorf("expected at least 30 plugins, got %d", len(plugins))
	}
}
