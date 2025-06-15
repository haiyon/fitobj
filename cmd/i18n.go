package cmd

import (
	"fmt"

	"github.com/haiyon/fitobj/i18n"
	"github.com/spf13/cobra"
)

var i18nCmd = &cobra.Command{
	Use:   "i18n",
	Short: "i18n key management utilities",
	Long: `Manage internationalization keys by extracting, comparing, and cleaning up
translation keys between source code and JSON files.`,
}

var i18nCheckCmd = &cobra.Command{
	Use:   "check [source-dir] [json-path]",
	Short: "Check for missing and unused i18n keys",
	Long: `Extract and compare i18n keys between source code and JSON files.
Reports missing keys in JSON and unused keys in source code.

Example:
  fitobj i18n check ./src ./translations
  fitobj i18n check ./app ./locales/en.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceDir := args[0]
		jsonPath := args[1]

		fmt.Printf("Extracting and comparing i18n keys...\n")
		fmt.Printf("Source directory: %s\n", sourceDir)
		fmt.Printf("JSON path: %s\n", jsonPath)

		return runI18nCheck(sourceDir, jsonPath, false)
	},
}

var i18nCleanCmd = &cobra.Command{
	Use:   "clean [source-dir] [json-path]",
	Short: "Remove unused keys from JSON files",
	Long: `Extract, compare, and automatically remove unused i18n keys from JSON files.

Example:
  fitobj i18n clean ./src ./translations
  fitobj i18n clean ./app ./locales --separator="__"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceDir := args[0]
		jsonPath := args[1]

		fmt.Printf("Extracting and comparing i18n keys...\n")
		fmt.Printf("Source directory: %s\n", sourceDir)
		fmt.Printf("JSON path: %s\n", jsonPath)
		fmt.Printf("Cleanup mode: Enabled (unused keys will be removed)\n")

		return runI18nCheck(sourceDir, jsonPath, true)
	},
}

func init() {
	i18nCmd.AddCommand(i18nCheckCmd)
	i18nCmd.AddCommand(i18nCleanCmd)
	rootCmd.AddCommand(i18nCmd)
}

func runI18nCheck(sourceDir, jsonPath string, cleanup bool) error {
	// Extract keys from source files
	sourceKeys, err := i18n.ExtractKeysFromDir(sourceDir)
	if err != nil {
		return fmt.Errorf("extracting keys from source: %v", err)
	}

	// Extract keys from JSON files
	jsonKeys, err := i18n.ExtractKeysFromJSONDir(jsonPath)
	if err != nil {
		return fmt.Errorf("extracting keys from JSON: %v", err)
	}

	// Compare and report
	missingInJSON, unusedInSource := i18n.CompareKeys(sourceKeys, jsonKeys)

	fmt.Printf("\n🔍 Total keys in source: %d\n", len(sourceKeys))
	fmt.Printf("📚 Total keys in JSON: %d\n", len(jsonKeys))

	fmt.Printf("\n❌ Missing in JSON (%d):\n", len(missingInJSON))
	for _, key := range missingInJSON {
		fmt.Println(key)
	}

	fmt.Printf("\n🟡 Unused in Source (%d):\n", len(unusedInSource))
	for _, key := range unusedInSource {
		fmt.Println(key)
	}

	// Cleanup if requested
	if cleanup && len(unusedInSource) > 0 {
		fmt.Println("\n🧹 Cleaning up unused keys...")
		separator := getSeparator()
		if err := i18n.CleanupUnusedKeys(jsonPath, unusedInSource, separator); err != nil {
			return fmt.Errorf("cleanup failed: %v", err)
		}
		fmt.Println("✅ Cleanup completed!")
	} else if cleanup && len(unusedInSource) == 0 {
		fmt.Println("\n✅ No unused keys to cleanup!")
	}

	return nil
}
