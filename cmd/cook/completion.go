package main

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

// completeCookFiles provides shell completion for .cook files
func completeCookFiles(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Look for .cook files matching the partial input
	pattern := toComplete + "*.cook"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Also try directory completion
	if toComplete != "" {
		dirPattern := toComplete + "*"
		dirMatches, _ := filepath.Glob(dirPattern)
		for _, m := range dirMatches {
			// Check if it's a directory
			if info, err := filepath.Glob(m + "/*.cook"); err == nil && len(info) > 0 {
				matches = append(matches, m+"/")
			}
		}
	}

	// If no specific prefix, show all .cook files in current directory
	if len(matches) == 0 && toComplete == "" {
		matches, _ = filepath.Glob("*.cook")
	}

	return matches, cobra.ShellCompDirectiveNoSpace
}

// completeUnitFlag provides shell completion for the --unit flag (unit systems only)
func completeUnitFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	units := []string{
		"metric\tConvert to metric units (g, kg, ml, l)",
		"imperial\tConvert to imperial units (oz, lb, fl oz)",
		"us\tConvert to US customary units (cup, tbsp, tsp)",
	}
	return units, cobra.ShellCompDirectiveNoFileComp
}

// completeFormatFlag provides shell completion for format flags
func completeFormatFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	formats := []string{
		"cooklang\tOriginal Cooklang format",
		"markdown\tMarkdown format",
		"html\tHTML format",
		"print\tPrint-optimized HTML",
		"json\tJSON format",
	}
	return formats, cobra.ShellCompDirectiveNoFileComp
}

// completeServingsFlag provides shell completion for servings flags
func completeServingsFlag(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Suggest common serving sizes
	servings := []string{
		"1\tSingle serving",
		"2\tTwo servings",
		"4\tFour servings (typical family)",
		"6\tSix servings",
		"8\tEight servings (dinner party)",
		"12\tTwelve servings (large gathering)",
	}
	return servings, cobra.ShellCompDirectiveNoFileComp
}
