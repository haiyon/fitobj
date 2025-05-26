package i18n

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractKeysFromFile(t *testing.T) {
	// Create a temporary test file
	content := `
	import React from 'react';

	function Component() {
		const message = t('hello.world');
		return (
			<div>
				<h1>{t("header.title")}</h1>
				<p>{t('content.description')}</p>
				<button>{t('buttons.submit')}</button>
			</div>
		);
	}
	`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsx")

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test extraction
	keys, err := ExtractKeysFromFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		"hello.world":         true,
		"header.title":        true,
		"content.description": true,
		"buttons.submit":      true,
	}

	if !reflect.DeepEqual(keys, expected) {
		t.Fatalf("Expected %v, got %v", expected, keys)
	}
}

func TestExtractKeysFromJSON(t *testing.T) {
	// Create a temporary test JSON file
	content := `{
		"hello": {
			"world": "Hello World"
		},
		"header": {
			"title": "Welcome"
		},
		"content": {
			"description": "This is a test"
		},
		"buttons": {
			"submit": "Submit"
		}
	}`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test extraction
	keys, err := ExtractKeysFromJSON(testFile)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		"hello.world":         true,
		"header.title":        true,
		"content.description": true,
		"buttons.submit":      true,
	}

	if !reflect.DeepEqual(keys, expected) {
		t.Fatalf("Expected %v, got %v", expected, keys)
	}
}

func TestCompareKeys(t *testing.T) {
	sourceKeys := map[string]bool{
		"hello.world":         true,
		"header.title":        true,
		"content.description": true,
		"buttons.submit":      true,
		"missing.key":         true,
	}

	jsonKeys := map[string]bool{
		"hello.world":         true,
		"header.title":        true,
		"content.description": true,
		"buttons.submit":      true,
		"unused.key":          true,
	}

	missingInJSON, unusedInSource := CompareKeys(sourceKeys, jsonKeys)

	expectedMissing := []string{"missing.key"}
	expectedUnused := []string{"unused.key"}

	// Compare slices regardless of order
	if len(missingInJSON) != len(expectedMissing) {
		t.Fatalf("Expected %v missing keys, got %v", expectedMissing, missingInJSON)
	}

	if len(unusedInSource) != len(expectedUnused) {
		t.Fatalf("Expected %v unused keys, got %v", expectedUnused, unusedInSource)
	}

	// Check if all expected keys are in the result
	missingMap := make(map[string]bool)
	for _, key := range missingInJSON {
		missingMap[key] = true
	}

	unusedMap := make(map[string]bool)
	for _, key := range unusedInSource {
		unusedMap[key] = true
	}

	for _, key := range expectedMissing {
		if !missingMap[key] {
			t.Fatalf("Expected missing key %s not found", key)
		}
	}

	for _, key := range expectedUnused {
		if !unusedMap[key] {
			t.Fatalf("Expected unused key %s not found", key)
		}
	}
}
