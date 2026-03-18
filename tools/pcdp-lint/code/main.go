package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	ExitOK         = 0
	ExitError      = 1
	ExitInvocation = 2
)

type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
)

type Diagnostic struct {
	Severity Severity
	Section  string
	Message  string
	Line     int
}

type LintResult struct {
	File        string
	Diagnostics []Diagnostic
	ExitCode    int
}

type MetaField struct {
	Key   string
	Value string
	Line  int
}

type ExampleBlock struct {
	Name        string
	HasGiven    bool
	HasWhen     bool
	HasThen     bool
	EmptyGiven  bool
	EmptyWhen   bool
	EmptyThen   bool
	Line        int
}

var knownDeploymentTemplates = []string{
	"wasm", "ebpf", "kernel-module", "crypto-library",
	"cli-tool", "gui-tool", "cloud-native", "backend-service",
	"library-c-abi", "enterprise-software", "academic",
	"enhance-existing", "manual", "template",
}

var knownVerificationValues = []string{
	"none", "lean4", "fstar", "dafny", "custom",
}

var spdxLicenses = map[string]bool{
	"Apache-2.0":         true,
	"MIT":                true,
	"GPL-2.0-only":       true,
	"GPL-3.0-only":       true,
	"LGPL-2.1-or-later":  true,
	"LGPL-3.0-or-later":  true,
	"BSD-2-Clause":       true,
	"BSD-3-Clause":       true,
	"ISC":                true,
	"MPL-2.0":            true,
	"CC0-1.0":            true,
	"Unlicense":          true,
	// Add more as needed
}

func main() {
	args := os.Args[1:]
	
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: missing file argument\n")
		fmt.Fprintf(os.Stderr, "Usage: pcdp-lint [strict=true] <specfile.md>\n")
		fmt.Fprintf(os.Stderr, "       pcdp-lint list-templates\n")
		os.Exit(ExitInvocation)
	}

	// Handle list-templates command
	if len(args) == 1 && args[0] == "list-templates" {
		listTemplates()
		return
	}

	// Parse arguments
	strict := false
	var filename string

	for _, arg := range args {
		if strings.HasPrefix(arg, "strict=") {
			value := strings.TrimPrefix(arg, "strict=")
			if value == "true" {
				strict = true
			} else if value == "false" {
				strict = false
			} else {
				fmt.Fprintf(os.Stderr, "error: strict must be true or false, got: %s\n", value)
				os.Exit(ExitInvocation)
			}
		} else if strings.Contains(arg, "=") {
			key := strings.Split(arg, "=")[0]
			fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", key)
			os.Exit(ExitInvocation)
		} else {
			if filename != "" {
				fmt.Fprintf(os.Stderr, "error: multiple file arguments not supported\n")
				os.Exit(ExitInvocation)
			}
			filename = arg
		}
	}

	if filename == "" {
		fmt.Fprintf(os.Stderr, "error: missing file argument\n")
		fmt.Fprintf(os.Stderr, "Usage: pcdp-lint [strict=true] <specfile.md>\n")
		fmt.Fprintf(os.Stderr, "       pcdp-lint list-templates\n")
		os.Exit(ExitInvocation)
	}

	// Check file extension
	if !strings.HasSuffix(filename, ".md") {
		fmt.Fprintf(os.Stderr, "error: file must have .md extension: %s\n", filename)
		os.Exit(ExitInvocation)
	}

	// Check if file exists and is readable
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: cannot open file: %s\n", filename)
		os.Exit(ExitInvocation)
	}

	// Perform lint
	result := lint(filename, strict)
	
	// Output diagnostics to stderr
	for _, diag := range result.Diagnostics {
		fmt.Fprintf(os.Stderr, "%s  %s:%d  [%s]  %s\n",
			diag.Severity, result.File, diag.Line, diag.Section, diag.Message)
	}

	// Output summary to stdout
	errorCount := 0
	warningCount := 0
	for _, diag := range result.Diagnostics {
		if diag.Severity == SeverityError {
			errorCount++
		} else {
			warningCount++
		}
	}

	if errorCount == 0 && warningCount == 0 {
		fmt.Printf("✓ %s: valid\n", result.File)
	} else if errorCount == 0 && !strict {
		fmt.Printf("✓ %s: valid (%d warning(s))\n", result.File, warningCount)
	} else if strict && errorCount == 0 && warningCount > 0 {
		fmt.Printf("✗ %s: %d error(s), %d warning(s) [strict mode]\n", result.File, errorCount, warningCount)
	} else {
		if strict && warningCount > 0 {
			fmt.Printf("✗ %s: %d error(s), %d warning(s) [strict mode]\n", result.File, errorCount, warningCount)
		} else {
			fmt.Printf("✗ %s: %d error(s), %d warning(s)\n", result.File, errorCount, warningCount)
		}
	}

	os.Exit(result.ExitCode)
}

func listTemplates() {
	templates := map[string]string{
		"wasm":                "Go",
		"ebpf":                "C",
		"kernel-module":       "C",
		"crypto-library":      "C",
		"cli-tool":            "Go",
		"gui-tool":            "Go",
		"cloud-native":        "Go",
		"backend-service":     "Go",
		"library-c-abi":       "C",
		"enterprise-software": "Go",
		"academic":            "Go",
		"enhance-existing":    "(declare Language: in META)",
		"manual":              "(declare Target: in META)",
		"template":            "(template definition file, not translatable)",
	}

	for _, template := range knownDeploymentTemplates {
		defaultLang := templates[template]
		fmt.Printf("%s  →  %s\n", template, defaultLang)
	}
}

func lint(filename string, strict bool) LintResult {
	result := LintResult{
		File:        filename,
		Diagnostics: []Diagnostic{},
		ExitCode:    ExitOK,
	}

	file, err := os.Open(filename)
	if err != nil {
		// This should not happen as we already checked file existence
		result.ExitCode = ExitInvocation
		return result
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Run all validation rules
	validateRequiredSections(lines, &result)
	metaFields := extractMetaFields(lines)
	validateMetaFields(metaFields, &result)
	validateExamplesSection(lines, &result)

	// Determine exit code
	hasErrors := false
	hasWarnings := false
	for _, diag := range result.Diagnostics {
		if diag.Severity == SeverityError {
			hasErrors = true
		} else {
			hasWarnings = true
		}
	}

	if hasErrors {
		result.ExitCode = ExitError
	} else if strict && hasWarnings {
		result.ExitCode = ExitError
	} else {
		result.ExitCode = ExitOK
	}

	return result
}

func validateRequiredSections(lines []string, result *LintResult) {
	requiredSections := []string{
		"## META",
		"## TYPES", 
		"## BEHAVIOR",
		"## PRECONDITIONS",
		"## POSTCONDITIONS",
		"## INVARIANTS",
		"## EXAMPLES",
	}

	foundSections := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, section := range requiredSections {
			if trimmed == section {
				foundSections[section] = true
			}
		}
		// Check for BEHAVIOR variants
		if strings.HasPrefix(trimmed, "## BEHAVIOR:") || strings.HasPrefix(trimmed, "## BEHAVIOR/INTERNAL:") {
			foundSections["## BEHAVIOR"] = true
		}
	}

	for _, section := range requiredSections {
		if !foundSections[section] {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "structure",
				Message:  fmt.Sprintf("Missing required section: %s", section),
				Line:     1,
			})
		}
	}
}

func extractMetaFields(lines []string) []MetaField {
	var fields []MetaField
	inMeta := false
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "## META" {
			inMeta = true
			continue
		}
		
		if inMeta && strings.HasPrefix(trimmed, "##") {
			break
		}
		
		if inMeta && strings.Contains(line, ":") && !strings.HasPrefix(trimmed, "---") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key != "" {
					fields = append(fields, MetaField{
						Key:   key,
						Value: value,
						Line:  i + 1,
					})
				}
			}
		}
	}
	
	return fields
}

func validateMetaFields(metaFields []MetaField, result *LintResult) {
	requiredFields := []string{
		"Deployment", "Verification", "Safety-Level",
		"Version", "Spec-Schema", "License",
	}

	fieldMap := make(map[string]MetaField)
	authors := []MetaField{}
	
	for _, field := range metaFields {
		if field.Key == "Author" {
			authors = append(authors, field)
		} else {
			fieldMap[field.Key] = field
		}
	}

	// Check required fields
	for _, field := range requiredFields {
		if meta, exists := fieldMap[field]; !exists {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Message:  fmt.Sprintf("Missing required META field: %s", field),
				Line:     1,
			})
		} else if meta.Value == "" {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Message:  fmt.Sprintf("META field %s has empty value", field),
				Line:     meta.Line,
			})
		}
	}

	// Check Author field
	if len(authors) == 0 {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: SeverityError,
			Section:  "META",
			Message:  "Missing required META field: Author (at least one Author: line required)",
			Line:     1,
		})
	} else {
		for _, author := range authors {
			if author.Value == "" {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Message:  "Author: field has empty value",
					Line:     author.Line,
				})
			}
		}
	}

	// Validate Version format
	if version, exists := fieldMap["Version"]; exists && version.Value != "" {
		if !isValidSemanticVersion(version.Value) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Message:  fmt.Sprintf("Version '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", version.Value),
				Line:     version.Line,
			})
		}
	}

	// Validate Spec-Schema format
	if schema, exists := fieldMap["Spec-Schema"]; exists && schema.Value != "" {
		if !isValidSemanticVersion(schema.Value) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Message:  fmt.Sprintf("Spec-Schema '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", schema.Value),
				Line:     schema.Line,
			})
		}
	}

	// Validate License SPDX
	if license, exists := fieldMap["License"]; exists && license.Value != "" {
		if !isValidSPDXLicense(license.Value) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Message:  fmt.Sprintf("License '%s' is not a valid SPDX identifier. See https://spdx.org/licenses/ for valid identifiers. Compound expressions permitted (e.g. Apache-2.0 OR MIT).", license.Value),
				Line:     license.Line,
			})
		}
	}

	// Validate Deployment template
	if deployment, exists := fieldMap["Deployment"]; exists && deployment.Value != "" {
		if !contains(knownDeploymentTemplates, deployment.Value) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Message:  fmt.Sprintf("Unknown deployment template: '%s'. Run 'pcdp-lint list-templates' to see valid values.", deployment.Value),
				Line:     deployment.Line,
			})
		}

		// Special validation for enhance-existing
		if deployment.Value == "enhance-existing" {
			if _, exists := fieldMap["Language"]; !exists {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Message:  "Deployment 'enhance-existing' requires META field 'Language'",
					Line:     deployment.Line,
				})
			} else if fieldMap["Language"].Value == "" {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Message:  "META field 'Language' has empty value",
					Line:     fieldMap["Language"].Line,
				})
			}
		}

		// Special validation for manual
		if deployment.Value == "manual" {
			if _, exists := fieldMap["Target"]; !exists {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Message:  "Deployment 'manual' requires META field 'Target' (no template available for language resolution)",
					Line:     deployment.Line,
				})
			}
		}

		// Check for deprecated Target field
		if target, exists := fieldMap["Target"]; exists && deployment.Value != "manual" {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityWarning,
				Section:  "META",
				Message:  "META field 'Target' is deprecated since v0.3.0. Target language is derived from the deployment template. Remove 'Target', or switch to Deployment: manual if explicit language control is required.",
				Line:     target.Line,
			})
		}
	}

	// Check for deprecated Domain field
	if domain, exists := fieldMap["Domain"]; exists {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: SeverityWarning,
			Section:  "META",
			Message:  "META field 'Domain' is deprecated since v0.3.0. Use 'Deployment' instead.",
			Line:     domain.Line,
		})
	}

	// Validate Verification field
	if verification, exists := fieldMap["Verification"]; exists && verification.Value != "" {
		if !contains(knownVerificationValues, verification.Value) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityWarning,
				Section:  "META",
				Message:  fmt.Sprintf("Unknown verification value: '%s'. Known values: none, lean4, fstar, dafny, custom. Custom verification backends are permitted; verify the value is intentional.", verification.Value),
				Line:     verification.Line,
			})
		}
	}
}

func validateExamplesSection(lines []string, result *LintResult) {
	inExamples := false
	exampleBlocks := []ExampleBlock{}
	currentExample := ExampleBlock{}
	currentExampleName := ""
	
	givenStart := -1
	whenStart := -1
	thenStart := -1
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "## EXAMPLES" {
			inExamples = true
			continue
		}
		
		if inExamples && strings.HasPrefix(trimmed, "##") {
			// End of EXAMPLES section
			if currentExampleName != "" {
				// Finalize current example
				currentExample.Name = currentExampleName
				currentExample.EmptyGiven = isBlockEmpty(lines, givenStart, whenStart)
				currentExample.EmptyWhen = isBlockEmpty(lines, whenStart, thenStart)
				currentExample.EmptyThen = isBlockEmpty(lines, thenStart, len(lines))
				exampleBlocks = append(exampleBlocks, currentExample)
			}
			break
		}
		
		if inExamples {
			if strings.HasPrefix(trimmed, "EXAMPLE:") {
				// Finalize previous example if exists
				if currentExampleName != "" {
					currentExample.Name = currentExampleName
					currentExample.EmptyGiven = isBlockEmpty(lines, givenStart, whenStart)
					currentExample.EmptyWhen = isBlockEmpty(lines, whenStart, thenStart)
					currentExample.EmptyThen = isBlockEmpty(lines, thenStart, i)
					exampleBlocks = append(exampleBlocks, currentExample)
				}
				
				// Start new example
				currentExampleName = strings.TrimSpace(strings.TrimPrefix(trimmed, "EXAMPLE:"))
				currentExample = ExampleBlock{Line: i + 1}
				givenStart = -1
				whenStart = -1
				thenStart = -1
			} else if trimmed == "GIVEN:" {
				currentExample.HasGiven = true
				givenStart = i + 1
			} else if trimmed == "WHEN:" {
				currentExample.HasWhen = true
				whenStart = i + 1
			} else if trimmed == "THEN:" {
				currentExample.HasThen = true
				thenStart = i + 1
			}
		}
	}
	
	// Finalize last example
	if inExamples && currentExampleName != "" {
		currentExample.Name = currentExampleName
		currentExample.EmptyGiven = isBlockEmpty(lines, givenStart, whenStart)
		currentExample.EmptyWhen = isBlockEmpty(lines, whenStart, thenStart)
		currentExample.EmptyThen = isBlockEmpty(lines, thenStart, len(lines))
		exampleBlocks = append(exampleBlocks, currentExample)
	}
	
	// Validate examples
	if len(exampleBlocks) == 0 && inExamples {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: SeverityError,
			Section:  "EXAMPLES",
			Message:  "EXAMPLES section contains no example blocks. Each example requires EXAMPLE:, GIVEN:, WHEN:, THEN: markers.",
			Line:     1,
		})
	}
	
	for _, example := range exampleBlocks {
		if !example.HasGiven {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' missing GIVEN: marker", example.Name),
				Line:     example.Line,
			})
		}
		if !example.HasWhen {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' missing WHEN: marker", example.Name),
				Line:     example.Line,
			})
		}
		if !example.HasThen {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' missing THEN: marker", example.Name),
				Line:     example.Line,
			})
		}
		
		// Check for empty blocks (warnings)
		if example.EmptyGiven {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityWarning,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' has empty GIVEN block", example.Name),
				Line:     example.Line,
			})
		}
		if example.EmptyWhen {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityWarning,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' has empty WHEN block", example.Name),
				Line:     example.Line,
			})
		}
		if example.EmptyThen {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: SeverityWarning,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' has empty THEN block", example.Name),
				Line:     example.Line,
			})
		}
	}
}

func isBlockEmpty(lines []string, start, end int) bool {
	if start == -1 || end == -1 || start >= end {
		return true
	}
	
	for i := start; i < end && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "EXAMPLE:") && 
		   !strings.HasPrefix(line, "GIVEN:") && 
		   !strings.HasPrefix(line, "WHEN:") && 
		   !strings.HasPrefix(line, "THEN:") &&
		   !strings.HasPrefix(line, "##") {
			return false
		}
	}
	
	return true
}

func isValidSemanticVersion(version string) bool {
	pattern := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	return pattern.MatchString(version)
}

func isValidSPDXLicense(license string) bool {
	// Simple validation - check if it's a known SPDX license or contains OR/AND
	license = strings.TrimSpace(license)
	
	// Handle compound expressions
	if strings.Contains(license, " OR ") || strings.Contains(license, " AND ") {
		parts := regexp.MustCompile(` OR | AND `).Split(license, -1)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if !spdxLicenses[part] {
				return false
			}
		}
		return true
	}
	
	return spdxLicenses[license]
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
