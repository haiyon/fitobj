package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractKeysFromFile(t *testing.T) {
	content := `
	import React from 'react';

	function Component() {
		const message = t('hello.world');
		return (
			<div>
				<h1>{t("header.title")}</h1>
				<p>{t('content.description')}</p>
				<button>{t('buttons.submit')}</button>
				<span>{t('nested.deep.key')}</span>
			</div>
		);
	}
	`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.jsx")

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	keys, err := ExtractKeysFromFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		"hello.world":         true,
		"header.title":        true,
		"content.description": true,
		"buttons.submit":      true,
		"nested.deep.key":     true,
	}

	if !reflect.DeepEqual(keys, expected) {
		t.Fatalf("Expected %v, got %v", expected, keys)
	}
}

func TestExtractKeysFromJSON(t *testing.T) {
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
		},
		"nested": {
			"deep": {
				"key": "Deep nested value"
			}
		}
	}`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	keys, err := ExtractKeysFromJSON(testFile)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		"hello.world":         true,
		"header.title":        true,
		"content.description": true,
		"buttons.submit":      true,
		"nested.deep.key":     true,
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
		"another.missing":     true,
	}

	jsonKeys := map[string]bool{
		"hello.world":         true,
		"header.title":        true,
		"content.description": true,
		"buttons.submit":      true,
		"unused.key":          true,
		"another.unused":      true,
	}

	missingInJSON, unusedInSource := CompareKeys(sourceKeys, jsonKeys)

	expectedMissing := []string{"missing.key", "another.missing"}
	expectedUnused := []string{"unused.key", "another.unused"}

	if len(missingInJSON) != len(expectedMissing) {
		t.Fatalf("Expected %d missing keys, got %d", len(expectedMissing), len(missingInJSON))
	}

	if len(unusedInSource) != len(expectedUnused) {
		t.Fatalf("Expected %d unused keys, got %d", len(expectedUnused), len(unusedInSource))
	}

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

func TestRemoveKeysFromPath(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]any
		keyPath  string
		expected map[string]any
		removed  bool
	}{
		{
			name: "Remove nested key",
			initial: map[string]any{
				"hello": map[string]any{
					"world": "Hello World",
				},
				"unused": map[string]any{
					"key": "Unused Value",
				},
			},
			keyPath: "unused.key",
			expected: map[string]any{
				"hello": map[string]any{
					"world": "Hello World",
				},
			},
			removed: true,
		},
		{
			name: "Remove top-level key",
			initial: map[string]any{
				"hello":  "world",
				"unused": "value",
			},
			keyPath: "unused",
			expected: map[string]any{
				"hello": "world",
			},
			removed: true,
		},
		{
			name: "Remove one key from multiple siblings",
			initial: map[string]any{
				"buttons": map[string]any{
					"submit": "Submit",
					"cancel": "Cancel",
					"unused": "Unused",
				},
			},
			keyPath: "buttons.unused",
			expected: map[string]any{
				"buttons": map[string]any{
					"submit": "Submit",
					"cancel": "Cancel",
				},
			},
			removed: true,
		},
		{
			name: "Remove deeply nested key with empty parent cleanup",
			initial: map[string]any{
				"deep": map[string]any{
					"nested": map[string]any{
						"object": map[string]any{
							"key": "value",
						},
					},
				},
				"keep": "this",
			},
			keyPath: "deep.nested.object.key",
			expected: map[string]any{
				"keep": "this",
			},
			removed: true,
		},
		{
			name: "Try to remove non-existent key",
			initial: map[string]any{
				"hello": "world",
			},
			keyPath: "nonexistent.key",
			expected: map[string]any{
				"hello": "world",
			},
			removed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveKeysFromPath(tt.initial, tt.keyPath, ".")

			if result != tt.removed {
				t.Fatalf("Expected removal result %v, got %v", tt.removed, result)
			}

			if !reflect.DeepEqual(tt.initial, tt.expected) {
				t.Fatalf("Expected %v, got %v", tt.expected, tt.initial)
			}
		})
	}
}

func TestSplitKeyPath(t *testing.T) {
	tests := []struct {
		input     string
		separator string
		expected  []string
	}{
		{"hello.world", ".", []string{"hello", "world"}},
		{"a.b.c.d", ".", []string{"a", "b", "c", "d"}},
		{"single", ".", []string{"single"}},
		{"", ".", []string{}},
		{"hello__world", "__", []string{"hello", "world"}},
		{"a__b__c", "__", []string{"a", "b", "c"}},
		{"deep.nested.object.key", ".", []string{"deep", "nested", "object", "key"}},
	}

	for _, test := range tests {
		result := splitKeyPath(test.input, test.separator)
		if !reflect.DeepEqual(result, test.expected) {
			t.Fatalf("For input '%s' with separator '%s', expected %v, got %v",
				test.input, test.separator, test.expected, result)
		}
	}
}

func TestCleanupUnusedKeys(t *testing.T) {
	initialContent := map[string]any{
		"hello": map[string]any{
			"world": "Hello World",
		},
		"header": map[string]any{
			"title": "Welcome",
		},
		"unused": map[string]any{
			"key1": "Unused Value 1",
			"key2": "Unused Value 2",
		},
		"buttons": map[string]any{
			"submit": "Submit",
			"unused": "Unused Button",
		},
		"completely": map[string]any{
			"unused": map[string]any{
				"deeply": map[string]any{
					"nested": "Should be removed",
				},
			},
		},
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	jsonData, err := json.MarshalIndent(initialContent, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(testFile, jsonData, 0644); err != nil {
		t.Fatal(err)
	}

	unusedKeys := []string{
		"unused.key1",
		"unused.key2",
		"buttons.unused",
		"completely.unused.deeply.nested",
	}

	err = CleanupUnusedKeys(testFile, unusedKeys, ".")
	if err != nil {
		t.Fatal(err)
	}

	updatedData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	var updatedObj map[string]any
	if err := json.Unmarshal(updatedData, &updatedObj); err != nil {
		t.Fatal(err)
	}

	// Verify unused keys and their empty parents were removed
	if _, exists := updatedObj["unused"]; exists {
		t.Fatal("unused object should have been completely removed")
	}

	if _, exists := updatedObj["completely"]; exists {
		t.Fatal("completely object should have been completely removed")
	}

	if buttonsMap, exists := updatedObj["buttons"].(map[string]any); exists {
		if _, exists := buttonsMap["unused"]; exists {
			t.Fatal("buttons.unused should have been removed")
		}
		if _, exists := buttonsMap["submit"]; !exists {
			t.Fatal("buttons.submit should still exist")
		}
	}

	// Verify used keys are still there
	if helloMap, exists := updatedObj["hello"].(map[string]any); exists {
		if _, exists := helloMap["world"]; !exists {
			t.Fatal("hello.world should still exist")
		}
	} else {
		t.Fatal("hello object should still exist")
	}

	if headerMap, exists := updatedObj["header"].(map[string]any); exists {
		if _, exists := headerMap["title"]; !exists {
			t.Fatal("header.title should still exist")
		}
	} else {
		t.Fatal("header object should still exist")
	}
}

func TestCleanupUnusedKeysDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple JSON files
	files := map[string]map[string]any{
		"en.json": {
			"common": map[string]any{
				"save":   "Save",
				"cancel": "Cancel",
				"unused": "Unused",
			},
			"unused_section": map[string]any{
				"key": "value",
			},
		},
		"zh.json": {
			"common": map[string]any{
				"save":   "保存",
				"cancel": "取消",
				"unused": "未使用",
			},
			"unused_section": map[string]any{
				"key": "值",
			},
		},
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		jsonData, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
			t.Fatal(err)
		}
	}

	unusedKeys := []string{"common.unused", "unused_section.key"}

	err := CleanupUnusedKeys(tmpDir, unusedKeys, ".")
	if err != nil {
		t.Fatal(err)
	}

	// Verify both files were cleaned up
	for filename := range files {
		filePath := filepath.Join(tmpDir, filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}

		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			t.Fatal(err)
		}

		// Check that unused_section was completely removed
		if _, exists := obj["unused_section"]; exists {
			t.Fatalf("unused_section should have been removed from %s", filename)
		}

		// Check that common.unused was removed but common still exists
		if commonMap, exists := obj["common"].(map[string]any); exists {
			if _, exists := commonMap["unused"]; exists {
				t.Fatalf("common.unused should have been removed from %s", filename)
			}
			if _, exists := commonMap["save"]; !exists {
				t.Fatalf("common.save should still exist in %s", filename)
			}
		} else {
			t.Fatalf("common section should still exist in %s", filename)
		}
	}
}

func TestCleanupEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.json")

	// Create file with only unused keys
	content := map[string]any{
		"unused": map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
	}

	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(testFile, jsonData, 0644); err != nil {
		t.Fatal(err)
	}

	unusedKeys := []string{"unused.key1", "unused.key2"}

	err = CleanupUnusedKeys(testFile, unusedKeys, ".")
	if err != nil {
		t.Fatal(err)
	}

	// Verify file becomes empty object
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}

	if len(obj) != 0 {
		t.Fatalf("Expected empty object, got %v", obj)
	}
}
