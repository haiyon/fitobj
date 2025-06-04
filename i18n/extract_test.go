package i18n

import (
    "encoding/json"
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

func TestRemoveKeysFromPath(t *testing.T) {
    // Test data
    jsonObj := map[string]any{
        "hello": map[string]any{
            "world": "Hello World",
        },
        "header": map[string]any{
            "title": "Welcome",
        },
        "unused": map[string]any{
            "key": "Unused Value",
        },
        "buttons": map[string]any{
            "submit": "Submit",
            "cancel": "Cancel",
        },
    }

    // Test removing existing key
    result := RemoveKeysFromPath(jsonObj, "unused.key", ".")
    if !result {
        t.Fatal("Expected to successfully remove unused.key")
    }

    // Verify key was removed
    if _, exists := jsonObj["unused"].(map[string]any)["key"]; exists {
        t.Fatal("Key should have been removed")
    }

    // Test removing non-existent key
    result = RemoveKeysFromPath(jsonObj, "nonexistent.key", ".")
    if result {
        t.Fatal("Expected false when removing non-existent key")
    }

    // Test removing top-level key
    result = RemoveKeysFromPath(jsonObj, "unused", ".")
    if !result {
        t.Fatal("Expected to successfully remove unused")
    }

    // Verify top-level key was removed
    if _, exists := jsonObj["unused"]; exists {
        t.Fatal("Top-level key should have been removed")
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
    // Create a temporary test JSON file
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
    }

    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.json")

    // Write initial content
    jsonData, err := json.MarshalIndent(initialContent, "", "  ")
    if err != nil {
        t.Fatal(err)
    }

    if err := os.WriteFile(testFile, jsonData, 0644); err != nil {
        t.Fatal(err)
    }

    // Test cleanup
    unusedKeys := []string{"unused.key1", "unused.key2", "buttons.unused"}
    err = CleanupUnusedKeys(testFile, unusedKeys, ".")
    if err != nil {
        t.Fatal(err)
    }

    // Read and verify the updated file
    updatedData, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatal(err)
    }

    var updatedObj map[string]any
    if err := json.Unmarshal(updatedData, &updatedObj); err != nil {
        t.Fatal(err)
    }

    // Verify unused keys were removed
    if unusedMap, exists := updatedObj["unused"].(map[string]any); exists {
        if _, exists := unusedMap["key1"]; exists {
            t.Fatal("unused.key1 should have been removed")
        }
        if _, exists := unusedMap["key2"]; exists {
            t.Fatal("unused.key2 should have been removed")
        }
    }

    if buttonsMap, exists := updatedObj["buttons"].(map[string]any); exists {
        if _, exists := buttonsMap["unused"]; exists {
            t.Fatal("buttons.unused should have been removed")
        }
        // Verify that used keys are still there
        if _, exists := buttonsMap["submit"]; !exists {
            t.Fatal("buttons.submit should still exist")
        }
    }

    // Verify used keys are still there
    if helloMap, exists := updatedObj["hello"].(map[string]any); exists {
        if _, exists := helloMap["world"]; !exists {
            t.Fatal("hello.world should still exist")
        }
    }
}
