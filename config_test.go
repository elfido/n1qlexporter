package main

import "testing"

func TestReadConfigFromFile(t *testing.T) {
	monitorDefinitions := getServers()
	if len(monitorDefinitions) != 2 {
		t.Errorf("Expected 2 monitor definitions, found %d", len(monitorDefinitions))
	}
}
