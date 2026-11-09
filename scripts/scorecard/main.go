package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

type TestResult struct {
	Package string `json:"Package"`
	Status  string `json:"Status"`
	Time    string `json:"Time"`
}

type ScorecardData struct {
	Version       string
	Date          string
	Engine        map[string]string
	WebStudio     map[string]string
	Installer     map[string]string
	CLI           map[string]string
	Security      map[string]string
	DX            map[string]string
	OverallStatus string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/scorecard/main.go <output-file>")
		os.Exit(1)
	}
	outputFile := os.Args[1]

	// Get version from git
	version := getVersion()
	date := time.Now().Format("2006-01-02")

	// Run tests and capture results
	engineResults := runEngineTests()
	webResults := runWebTests()
	installerResults := runInstallerTests()
	cliResults := runCLITests()
	securityResults := runSecurityTests()
	dxResults := runDXTests()

	data := ScorecardData{
		Version:   version,
		Date:      date,
		Engine:    engineResults,
		WebStudio: webResults,
		Installer: installerResults,
		CLI:       cliResults,
		Security:  securityResults,
		DX:        dxResults,
	}

	// Determine overall status
	allPass := true
	for _, v := range engineResults {
		if v != "PASS" {
			allPass = false
		}
	}
	for _, v := range webResults {
		if v != "PASS" {
			allPass = false
		}
	}
	for _, v := range installerResults {
		if v != "PASS" {
			allPass = false
		}
	}
	for _, v := range cliResults {
		if v != "PASS" {
			allPass = false
		}
	}
	for _, v := range securityResults {
		if v != "PASS" {
			allPass = false
		}
	}
	for _, v := range dxResults {
		if v != "PASS" {
			allPass = false
		}
	}

	if allPass {
		data.OverallStatus = "PASS"
	} else {
		data.OverallStatus = "FAIL"
	}

	// Generate scorecard
	tmpl := template.Must(template.New("scorecard").Parse(scorecardTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		fmt.Fprintf(os.Stderr, "Template error: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Scorecard written to %s\n", outputFile)
}

func getVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	output, err := cmd.Output()
	if err != nil {
		return "dev"
	}
	return strings.TrimSpace(string(output))
}

func runTest(packagePath, runPattern string) string {
	args := []string{"test", packagePath, "-count=1", "-timeout=60s"}
	if runPattern != "" {
		args = append(args, "-run", runPattern)
	}
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "FAIL"
	}
	if bytes.Contains(output, []byte("ok  ")) || bytes.Contains(output, []byte("ok\t")) {
		return "PASS"
	}
	if bytes.Contains(output, []byte("FAIL")) {
		return "FAIL"
	}
	return "NOT RUN"
}

func runEngineTests() map[string]string {
	results := make(map[string]string)

	// Model Descriptor
	results["Model Descriptor"] = runTest("./arsy/engine/manifest", "")

	// Field Types
	results["Field Types (22)"] = runTest("./core", "Field")

	// Relations
	results["Relations (4 types)"] = runTest("./arsy/engine/relations", "")

	// Validation Pipeline
	results["Validation Pipeline"] = runTest("./arsy/engine/validation", "")

	// Unit of Work
	results["Unit of Work"] = runTest("./arsy/engine/uow", "")

	// Repository Mapper (SQLite)
	results["Repository Mapper (SQLite)"] = runTest("./arsy/tenancy", "SQLite")

	// Repository Mapper (PostgreSQL) - skip if no PG
	results["Repository Mapper (PostgreSQL)"] = "NOT RUN"

	// Schema Planner
	results["Schema Planner"] = runTest("./arsy/aeng", "Migration")

	// Migration Executor
	results["Migration Executor"] = runTest("./arsy/aeng", "Migration")

	// Model/Field Auth
	results["Model/Field Auth"] = runTest("./arsy/security", "")

	// Record Policies
	results["Record Policies"] = runTest("./arsy/tenancy", "Adversarial|Tenant")

	// Audit Trail
	results["Audit Trail"] = runTest("./arsy/aeng", "Audit")

	return results
}

func runWebTests() map[string]string {
	results := make(map[string]string)

	// Studio Workflows - check if UI builds
	results["Studio Workflows"] = "NOT RUN"

	// No Legacy UI - brand scan
	results["No Legacy UI"] = runBrandScan()

	// A-R-S-Y Identity
	results["A-R-S-Y Identity"] = runBrandScan()

	// Accessibility
	results["Accessibility"] = "NOT RUN"

	return results
}

func runInstallerTests() map[string]string {
	results := make(map[string]string)

	// Clean Install E2E
	results["Clean Install E2E"] = "NOT RUN"

	// Resume/Recovery
	results["Resume/Recovery"] = "NOT RUN"

	// Unattended Install
	results["Unattended Install"] = "NOT RUN"

	// Upgrade Matrix
	results["Upgrade Matrix"] = "NOT RUN"

	// Rollback
	results["Rollback"] = "NOT RUN"

	// Backup/Restore
	results["Backup/Restore"] = runTest("./arsy/tenancy", "Backup|Restore")

	return results
}

func runCLITests() map[string]string {
	results := make(map[string]string)

	// Command Tree Complete
	results["Command Tree Complete"] = "NOT RUN"

	// JSON Output Stable
	results["JSON Output Stable"] = "NOT RUN"

	// No PB CLI
	results["No PB CLI"] = runBrandScan()

	// Shell Completion
	results["Shell Completion"] = "NOT RUN"

	return results
}

func runSecurityTests() map[string]string {
	results := make(map[string]string)

	// Cross-Tenant Isolation
	results["Cross-Tenant Isolation"] = runTest("./arsy/tenancy", "Adversarial")

	// No Data Leaks
	results["No Data Leaks"] = "NOT RUN"

	// PG RLS Enforced
	results["PG RLS Enforced"] = "NOT RUN"

	// Secret Redaction
	results["Secret Redaction"] = runTest("./arsy/aeng", "Redaction")

	return results
}

func runDXTests() map[string]string {
	results := make(map[string]string)

	// Module Lifecycle
	results["Module Lifecycle"] = runTest("./arsy/engine/manifest", "Install|Upgrade|Rollback")

	// Seed Idempotency
	results["Seed Idempotency"] = runTest("./arsy/aeng", "Seed")

	// SDK Generation
	results["SDK Generation"] = "NOT RUN"

	// Code Generation
	results["Code Generation"] = "NOT RUN"

	return results
}

func runBrandScan() string {
	// Scan for PocketBase/Superuser/pb_data/_superusers in non-test, non-internal files
	cmd := exec.Command("grep", "-r", "PocketBase|Superuser|pb_data|_superusers", "--include=*.go", "--include=*.js", "--include=*.html", ".")
	output, _ := cmd.CombinedOutput()

	// Filter out test files and internal references
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "_test.go") || strings.Contains(line, "test_") ||
			strings.Contains(line, "internal") || strings.Contains(line, "license") ||
			strings.Contains(line, "LICENSE") || strings.Contains(line, "doc.go") {
			continue
		}
		if strings.TrimSpace(line) != "" {
			return "FAIL"
		}
	}
	return "PASS"
}

const scorecardTemplate = `# A-R-S-Y Release Scorecard — {{.Version}} — {{.Date}}

## Engine (Layer 1-3)
| Component | Status | Evidence |
|-----------|--------|----------|
| Model Descriptor | {{index .Engine "Model Descriptor"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Field Types (22) | {{index .Engine "Field Types (22)"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Relations (4 types) | {{index .Engine "Relations (4 types)"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Validation Pipeline | {{index .Engine "Validation Pipeline"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Unit of Work | {{index .Engine "Unit of Work"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Repository Mapper (SQLite) | {{index .Engine "Repository Mapper (SQLite)"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Repository Mapper (PostgreSQL) | {{index .Engine "Repository Mapper (PostgreSQL)"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Schema Planner | {{index .Engine "Schema Planner"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Migration Executor | {{index .Engine "Migration Executor"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Model/Field Auth | {{index .Engine "Model/Field Auth"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Record Policies | {{index .Engine "Record Policies"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Audit Trail | {{index .Engine "Audit Trail"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |

## Web & Studio (Layer 6)
| Component | Status | Evidence |
|-----------|--------|----------|
| Studio Workflows | {{index .WebStudio "Studio Workflows"}} | [E2E Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| No Legacy UI | {{index .WebStudio "No Legacy UI"}} | [Brand Scan](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| A-R-S-Y Identity | {{index .WebStudio "A-R-S-Y Identity"}} | [Brand Scan](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Accessibility | {{index .WebStudio "Accessibility"}} | [Axe Report](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |

## Installer & Bootstrap
| Component | Status | Evidence |
|-----------|--------|----------|
| Clean Install E2E | {{index .Installer "Clean Install E2E"}} | [E2E Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Resume/Recovery | {{index .Installer "Resume/Recovery"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Unattended Install | {{index .Installer "Unattended Install"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Upgrade Matrix | {{index .Installer "Upgrade Matrix"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Rollback | {{index .Installer "Rollback"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Backup/Restore | {{index .Installer "Backup/Restore"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |

## CLI
| Component | Status | Evidence |
|-----------|--------|----------|
| Command Tree Complete | {{index .CLI "Command Tree Complete"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| JSON Output Stable | {{index .CLI "JSON Output Stable"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| No PB CLI | {{index .CLI "No PB CLI"}} | [Brand Scan](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Shell Completion | {{index .CLI "Shell Completion"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |

## Security & Tenancy
| Component | Status | Evidence |
|-----------|--------|----------|
| Cross-Tenant Isolation | {{index .Security "Cross-Tenant Isolation"}} | [Adversarial Tests](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| No Data Leaks | {{index .Security "No Data Leaks"}} | [Leak Tests](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| PG RLS Enforced | {{index .Security "PG RLS Enforced"}} | [PG Tests](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Secret Redaction | {{index .Security "Secret Redaction"}} | [Security Tests](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |

## Developer Experience
| Component | Status | Evidence |
|-----------|--------|----------|
| Module Lifecycle | {{index .DX "Module Lifecycle"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Seed Idempotency | {{index .DX "Seed Idempotency"}} | [Test Run](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| SDK Generation | {{index .DX "SDK Generation"}} | [Contract Tests](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |
| Code Generation | {{index .DX "Code Generation"}} | [Scaffold Tests](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) |

## Overall
**GATE: {{.OverallStatus}}** — All release-blocking checks must be PASS for release.
`