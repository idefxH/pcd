package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type WizardState struct {
	SessionID    string    `json:"session_id"`
	Component    string    `json:"component"`
	StartedAt    time.Time `json:"started_at"`
	LastUpdated  time.Time `json:"last_updated"`
	SectionsDone []string  `json:"sections_done"`
	PartialSpec  string    `json:"partial_spec"`
}

type LintResult int

const (
	LintPassed LintResult = iota
	LintFailed
	LintSkipped
)

var requiredSections = []string{
	"META", "TYPES", "BEHAVIOR", "PRECONDITIONS",
	"POSTCONDITIONS", "INVARIANTS", "EXAMPLES", "DEPLOYMENT",
}

func main() {
	args := parseArgs()

	if args["list-sessions"] == "true" {
		listSessions()
		os.Exit(0)
	}

	component := args["component"]
	output := args["output"]

	if component == "" {
		component = askString("Component name: ")
	}

	if output == "" {
		output = fmt.Sprintf("./%s.md", strings.ToLower(strings.ReplaceAll(component, " ", "-")))
	}

	state, err := startOrResume(component, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	_, lintResult := interview(state)

	switch lintResult {
	case LintPassed:
		fmt.Printf("✓ %s: written and valid\n", output)
		os.Exit(0)
	case LintFailed:
		fmt.Printf("✗ %s: written with errors — run pcdp-lint %s to review\n", output, output)
		os.Exit(1)
	case LintSkipped:
		fmt.Printf("✗ %s: written, pcdp-lint not found\n", output)
		os.Exit(0)
	}
}

func parseArgs() map[string]string {
	args := make(map[string]string)
	
	for _, arg := range os.Args[1:] {
		if arg == "list-sessions" {
			args["list-sessions"] = "true"
			continue
		}
		
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			args[parts[0]] = parts[1]
		}
	}
	
	return args
}

func startOrResume(component, output string) (*WizardState, error) {
	stateDir := filepath.Join(os.Getenv("HOME"), ".config", "pcdp", "wizard-state")
	err := os.MkdirAll(stateDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create state directory: %v", err)
	}

	// Check for existing session
	files, _ := ioutil.ReadDir(stateDir)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		data, err := ioutil.ReadFile(filepath.Join(stateDir, file.Name()))
		if err != nil {
			continue
		}
		
		var state WizardState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}
		
		if state.Component == component {
			fmt.Printf("Resuming session for '%s' (started %s)\n", 
				component, state.StartedAt.Format("2006-01-02 15:04"))
			fmt.Printf("Completed: %s\n", strings.Join(state.SectionsDone, ", "))
			
			nextSection := findNextSection(state.SectionsDone)
			if nextSection != "" {
				fmt.Printf("Continuing from: %s\n", nextSection)
			}
			
			state.LastUpdated = time.Now()
			saveState(&state, stateDir)
			return &state, nil
		}
	}

	// Create new session
	sessionID := uuid.New().String()
	state := &WizardState{
		SessionID:    sessionID,
		Component:    component,
		StartedAt:    time.Now(),
		LastUpdated:  time.Now(),
		SectionsDone: []string{},
		PartialSpec:  output,
	}

	saveState(state, stateDir)
	return state, nil
}

func findNextSection(done []string) string {
	doneMap := make(map[string]bool)
	for _, section := range done {
		doneMap[section] = true
	}
	
	for _, section := range requiredSections {
		if !doneMap[section] {
			return section
		}
	}
	return ""
}

func saveState(state *WizardState, stateDir string) error {
	state.LastUpdated = time.Now()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	
	stateFile := filepath.Join(stateDir, state.SessionID+".json")
	tempFile := stateFile + ".tmp"
	
	err = ioutil.WriteFile(tempFile, data, 0644)
	if err != nil {
		return err
	}
	
	return os.Rename(tempFile, stateFile)
}

func deleteState(state *WizardState) {
	stateDir := filepath.Join(os.Getenv("HOME"), ".config", "pcdp", "wizard-state")
	stateFile := filepath.Join(stateDir, state.SessionID+".json")
	os.Remove(stateFile)
}

func interview(state *WizardState) (string, LintResult) {
	var spec strings.Builder
	
	doneMap := make(map[string]bool)
	for _, section := range state.SectionsDone {
		doneMap[section] = true
	}

	// Process sections in order
	for _, section := range requiredSections {
		if doneMap[section] {
			continue
		}
		
		switch section {
		case "META":
			interviewMeta(&spec, state)
		case "TYPES":
			interviewTypes(&spec, state)
		case "BEHAVIOR":
			interviewBehavior(&spec, state)
		case "PRECONDITIONS":
			interviewConditions(&spec, state, "PRECONDITIONS")
		case "POSTCONDITIONS":
			interviewConditions(&spec, state, "POSTCONDITIONS")
		case "INVARIANTS":
			interviewInvariants(&spec, state)
		case "EXAMPLES":
			interviewExamples(&spec, state)
		case "DEPLOYMENT":
			interviewDeployment(&spec, state)
		}
		
		state.SectionsDone = append(state.SectionsDone, section)
		saveState(state, filepath.Join(os.Getenv("HOME"), ".config", "pcdp", "wizard-state"))
	}

	// Write complete spec
	finalSpec := fmt.Sprintf("# %s\n\n%s", state.Component, spec.String())
	
	// Check if file exists and ask for confirmation
	if _, err := os.Stat(state.PartialSpec); err == nil {
		if !askConfirm(fmt.Sprintf("%s already exists. Overwrite? (y/N)", state.PartialSpec)) {
			os.Exit(0)
		}
	}
	
	err := ioutil.WriteFile(state.PartialSpec, []byte(finalSpec), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing spec file: %v\n", err)
		os.Exit(2)
	}

	// Run pcdp-lint
	lintResult := runLint(state.PartialSpec)
	
	if lintResult == LintPassed {
		deleteState(state)
	}
	
	return finalSpec, lintResult
}

func interviewMeta(spec *strings.Builder, state *WizardState) {
	spec.WriteString("## META\n")
	
	// Get available templates
	templates := getAvailableTemplates()
	if len(templates) == 0 {
		fmt.Fprintf(os.Stderr, "error: no deployment templates found\n")
		os.Exit(2)
	}
	
	fmt.Println("Available deployment templates:")
	for i, template := range templates {
		fmt.Printf("%d. %s\n", i+1, template)
	}
	
	templateIdx := askInt("Select deployment template (number): ", 1, len(templates))
	deployment := templates[templateIdx-1]
	
	version := askStringWithDefault("Version", "0.1.0")
	author := askStringWithDefault("Author", getDefaultAuthor())
	license := askLicense()
	verification := askChoice("Verification", []string{"none", "lean4", "fstar", "dafny", "custom"}, "none")
	safetyLevel := askChoice("Safety-Level", []string{"QM", "ASIL-A", "ASIL-B", "ASIL-C", "ASIL-D", "DAL-A", "DAL-B", "DAL-C", "DAL-D", "DAL-E"}, "QM")
	
	spec.WriteString(fmt.Sprintf("Deployment:  %s\n", deployment))
	spec.WriteString(fmt.Sprintf("Version:     %s\n", version))
	spec.WriteString("Spec-Schema: 0.3.7\n")
	spec.WriteString(fmt.Sprintf("Author:      %s\n", author))
	spec.WriteString(fmt.Sprintf("License:     %s\n", license))
	spec.WriteString(fmt.Sprintf("Verification: %s\n", verification))
	spec.WriteString(fmt.Sprintf("Safety-Level: %s\n", safetyLevel))
	spec.WriteString("\n---\n\n")
}

func interviewTypes(spec *strings.Builder, state *WizardState) {
	spec.WriteString("## TYPES\n\n")
	
	if !askConfirm("Does your component work with custom data types? (Y/n)") {
		spec.WriteString("```\n// No custom types required for this component\n```\n\n---\n\n")
		return
	}
	
	spec.WriteString("```\n")
	
	for {
		typeName := askString("Type name: ")
		definition := askString("Definition: ")
		constraints := askString("Constraints (optional): ")
		
		spec.WriteString(fmt.Sprintf("%s := %s", typeName, definition))
		if constraints != "" {
			spec.WriteString(fmt.Sprintf(" where %s", constraints))
		}
		spec.WriteString("\n\n")
		
		if !askConfirm("Add another type? (Y/n)") {
			break
		}
	}
	
	spec.WriteString("```\n\n---\n\n")
}

func interviewBehavior(spec *strings.Builder, state *WizardState) {
	for {
		behaviorName := askString("Behavior name: ")
		
		spec.WriteString(fmt.Sprintf("## BEHAVIOR: %s\n\n", behaviorName))
		spec.WriteString("INPUTS:\n```\n")
		
		fmt.Println("Enter inputs (name: type pairs, one per line, empty to finish):")
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}
			spec.WriteString(line + "\n")
		}
		
		spec.WriteString("```\n\n")
		spec.WriteString("OUTPUTS:\n```\n")
		spec.WriteString("// TODO: Define outputs\n")
		spec.WriteString("```\n\n")
		spec.WriteString("PRECONDITIONS:\n- TODO: Define preconditions\n\n")
		spec.WriteString("POSTCONDITIONS:\n- TODO: Define postconditions\n\n")
		spec.WriteString("SIDE-EFFECTS:\n- TODO: Define side effects\n\n")
		
		if !askConfirm("Add another BEHAVIOR section? (Y/n)") {
			break
		}
	}
	
	spec.WriteString("---\n\n")
}

func interviewConditions(spec *strings.Builder, state *WizardState, sectionType string) {
	spec.WriteString(fmt.Sprintf("## %s\n\n", sectionType))
	
	var prompt string
	if sectionType == "PRECONDITIONS" {
		prompt = "List the conditions that must be true before your component runs.\nOne condition per line. Empty line to finish.\nExample: from.balance >= amount"
	} else {
		prompt = "List the conditions that must be true after your component runs.\nOne condition per line. Empty line to finish.\nExample: from.balance' = from.balance - amount"
	}
	
	fmt.Println(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("- ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}
		spec.WriteString(fmt.Sprintf("- %s\n", line))
	}
	
	spec.WriteString("\n---\n\n")
}

func interviewInvariants(spec *strings.Builder, state *WizardState) {
	spec.WriteString("## INVARIANTS\n\n")
	
	fmt.Println("List conditions that must always hold, regardless of which operation runs.")
	fmt.Println("Prefix with GLOBAL: for system-wide invariants.")
	fmt.Println("One invariant per line. Empty line to finish.")
	fmt.Println("Example: GLOBAL: ∀ a: Account. a.balance >= 0")
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("- ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}
		spec.WriteString(fmt.Sprintf("- %s\n", line))
	}
	
	spec.WriteString("\n---\n\n")
}

func interviewExamples(spec *strings.Builder, state *WizardState) {
	spec.WriteString("## EXAMPLES\n\n")
	
	exampleCount := 0
	for {
		exampleName := askString("Example name (identifier, no spaces): ")
		
		spec.WriteString(fmt.Sprintf("EXAMPLE: %s\n", exampleName))
		
		spec.WriteString("GIVEN:\n")
		fmt.Println("Describe the starting state. One item per line. Empty to finish.")
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("  ")
			if !scanner.Scan() {
				break
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}
			spec.WriteString(fmt.Sprintf("  %s\n", line))
		}
		
		when := askString("WHEN (describe the operation being performed): ")
		spec.WriteString(fmt.Sprintf("WHEN:\n  %s\n", when))
		
		spec.WriteString("THEN:\n")
		fmt.Println("Describe the expected outcome. One item per line. Empty to finish.")
		for {
			fmt.Print("  ")
			if !scanner.Scan() {
				break
			}
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}
			spec.WriteString(fmt.Sprintf("  %s\n", line))
		}
		
		spec.WriteString("\n")
		exampleCount++
		
		if !askConfirm("Add another example? (Y/n)") {
			break
		}
	}
	
	if exampleCount < 2 {
		fmt.Println("Consider adding a failure/error case example for completeness.")
	}
	
	spec.WriteString("---\n\n")
}

func interviewDeployment(spec *strings.Builder, state *WizardState) {
	spec.WriteString("## DEPLOYMENT\n\n")
	
	runtime := askStringWithDefault("Runtime description", "command-line tool, single static binary")
	installation := askStringWithDefault("Installation notes", "OBS package")
	platform := askStringWithDefault("Platform", "Linux (primary)")
	additional := askString("Any additional deployment notes (optional): ")
	
	spec.WriteString(fmt.Sprintf("Runtime: %s\n", runtime))
	spec.WriteString(fmt.Sprintf("Installation: %s\n", installation))
	spec.WriteString(fmt.Sprintf("Platform: %s\n", platform))
	
	if additional != "" {
		spec.WriteString(fmt.Sprintf("\n%s\n", additional))
	}
}

func getAvailableTemplates() []string {
	templateDir := "/usr/share/pcdp/templates"
	
	// Fallback for development/testing
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return []string{"cli-tool", "library", "service"}
	}
	
	files, err := ioutil.ReadDir(templateDir)
	if err != nil {
		return []string{"cli-tool"}
	}
	
	var templates []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".template.md") {
			name := strings.TrimSuffix(file.Name(), ".template.md")
			templates = append(templates, name)
		}
	}
	
	if len(templates) == 0 {
		return []string{"cli-tool"}
	}
	
	return templates
}

func getDefaultAuthor() string {
	// Try to get from git config
	if cmd := exec.Command("git", "config", "user.name"); cmd.Err == nil {
		if name, err := cmd.Output(); err == nil {
			if cmd := exec.Command("git", "config", "user.email"); cmd.Err == nil {
				if email, err := cmd.Output(); err == nil {
					return fmt.Sprintf("%s <%s>", 
						strings.TrimSpace(string(name)), 
						strings.TrimSpace(string(email)))
				}
			}
		}
	}
	return ""
}

func askString(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func askStringWithDefault(prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	value := strings.TrimSpace(scanner.Text())
	
	if value == "" && defaultValue != "" {
		return defaultValue
	}
	return value
}

func askInt(prompt string, min, max int) int {
	for {
		fmt.Print(prompt)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		
		var num int
		if _, err := fmt.Sscanf(scanner.Text(), "%d", &num); err == nil {
			if num >= min && num <= max {
				return num
			}
		}
		fmt.Printf("Please enter a number between %d and %d\n", min, max)
	}
}

func askConfirm(prompt string) bool {
	fmt.Print(prompt + " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	
	// Default to yes for Y/n prompts
	if strings.HasSuffix(prompt, "(Y/n)") && response == "" {
		return true
	}
	
	return response == "y" || response == "yes"
}

func askChoice(prompt string, choices []string, defaultChoice string) string {
	fmt.Printf("%s [%s] (options: %s): ", prompt, defaultChoice, strings.Join(choices, ", "))
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())
	
	if choice == "" {
		return defaultChoice
	}
	
	for _, validChoice := range choices {
		if choice == validChoice {
			return choice
		}
	}
	
	return defaultChoice
}

func askLicense() string {
	fmt.Println("Common licenses: Apache-2.0, MIT, GPL-2.0-only, GPL-3.0-only, CC-BY-4.0")
	return askStringWithDefault("License (SPDX identifier)", "Apache-2.0")
}

func listSessions() {
	stateDir := filepath.Join(os.Getenv("HOME"), ".config", "pcdp", "wizard-state")
	files, err := ioutil.ReadDir(stateDir)
	if err != nil || len(files) == 0 {
		fmt.Println("No resumable sessions found.")
		return
	}
	
	found := false
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		data, err := ioutil.ReadFile(filepath.Join(stateDir, file.Name()))
		if err != nil {
			continue
		}
		
		var state WizardState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}
		
		found = true
		fmt.Printf("Session: %s\n", state.SessionID)
		fmt.Printf("Component: %s\n", state.Component)
		fmt.Printf("Started: %s\n", state.StartedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Last Updated: %s\n", state.LastUpdated.Format("2006-01-02 15:04:05"))
		fmt.Printf("Sections Completed: %s\n", strings.Join(state.SectionsDone, ", "))
		fmt.Printf("Partial Spec: %s\n", state.PartialSpec)
		fmt.Println()
	}
	
	if !found {
		fmt.Println("No resumable sessions found.")
	}
}

func runLint(specFile string) LintResult {
	cmd := exec.Command("pcdp-lint", specFile)
	err := cmd.Run()
	
	if err != nil {
		// Check if command not found
		if strings.Contains(err.Error(), "executable file not found") {
			fmt.Fprintf(os.Stderr, "warning: pcdp-lint not found in PATH\n")
			return LintSkipped
		}
		
		// pcdp-lint found errors
		fmt.Fprintf(os.Stderr, "pcdp-lint reported errors in %s\n", specFile)
		return LintFailed
	}
	
	return LintPassed
}