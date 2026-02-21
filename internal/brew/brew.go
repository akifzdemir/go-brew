package brew

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Package represents an installed Homebrew package.
type Package struct {
	Name             string
	InstalledVersion string
	LatestVersion    string
	Outdated         bool
	InstalledSize    string // e.g. "54M", "1.2G", "—" if unknown
}

// InfoResult holds the output of `brew info --json=v2 <name>`.
type InfoResult struct {
	Name          string
	FullName      string
	Tap           string
	Version       string
	Desc          string
	Homepage      string
	License       string
	Dependencies  []string
	Conflicts     []string
	Installed     []InstalledVersion
	InstalledSize string // disk usage of the cellar directory
}

type InstalledVersion struct {
	Version string `json:"version"`
}

// outdatedJSON is the structure returned by `brew outdated --json=v2`.
type outdatedJSON struct {
	Formulae []struct {
		Name              string   `json:"name"`
		InstalledVersions []string `json:"installed_versions"`
		CurrentVersion    string   `json:"current_version"`
	} `json:"formulae"`
	Casks []struct {
		Name              string   `json:"name"`
		InstalledVersions []string `json:"installed_versions"`
		CurrentVersion    string   `json:"current_version"`
	} `json:"casks"`
}

// brewInfoJSON mirrors the relevant parts of `brew info --json=v2`.
type brewInfoJSON struct {
	Formulae []struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Tap      string `json:"tap"`
		Desc     string `json:"desc"`
		Homepage string `json:"homepage"`
		License  string `json:"license"`
		Versions struct {
			Stable string `json:"stable"`
		} `json:"versions"`
		Dependencies []string `json:"dependencies"`
		Conflicts    []struct {
			Name string `json:"name"`
		} `json:"conflicts_with"`
		Installed []struct {
			Version string `json:"version"`
		} `json:"installed"`
	} `json:"formulae"`
}

// brewSearchJSON mirrors `brew search --json=v2`.
type brewSearchJSON struct {
	Formulae []struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	} `json:"formulae"`
}

// runBrewCommand runs a brew subcommand and returns combined stdout+stderr.
func runBrewCommand(args ...string) (string, error) {
	cmd := exec.Command("brew", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// runBrewStdout runs a brew subcommand and returns only stdout.
// Use this for JSON-producing commands where stderr noise would break parsing.
func runBrewStdout(args ...string) (string, error) {
	cmd := exec.Command("brew", args...)
	out, err := cmd.Output()
	return string(out), err
}

// ListInstalled returns all top-level installed packages (brew leaves) with
// their installed versions. Outdated information is merged in separately.
func ListInstalled() ([]Package, error) {
	// Get top-level packages via brew leaves
	leavesOut, err := runBrewCommand("leaves")
	if err != nil {
		// Fall back to full list if leaves fails
		leavesOut, err = runBrewCommand("list", "--formula")
		if err != nil {
			return nil, fmt.Errorf("brew list failed: %w", err)
		}
	}

	names := parseLines(leavesOut)
	if len(names) == 0 {
		return []Package{}, nil
	}

	// Get versions for all packages at once
	versionsOut, err := runBrewCommand(append([]string{"list", "--versions"}, names...)...)
	if err != nil {
		return nil, fmt.Errorf("brew list --versions failed: %w", err)
	}

	// Build version map: name -> installed version
	versionMap := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(versionsOut), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// brew list --versions can return multiple versions; take the last (most recent)
			versionMap[parts[0]] = parts[len(parts)-1]
		}
	}

	// Get outdated map
	outdatedMap, err := getOutdatedMap()
	if err != nil {
		// Non-fatal — just show no outdated info
		outdatedMap = map[string]string{}
	}

	packages := make([]Package, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		ver := versionMap[name]
		latest, outdated := outdatedMap[name]
		pkg := Package{
			Name:             name,
			InstalledVersion: ver,
			Outdated:         outdated,
		}
		if outdated {
			pkg.LatestVersion = latest
		} else {
			pkg.LatestVersion = ver
		}
		packages = append(packages, pkg)
	}

	// Fetch disk sizes concurrently
	cellar := cellarPrefix()
	var wg sync.WaitGroup
	for i := range packages {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pkg := &packages[i]
			dir := filepath.Join(cellar, pkg.Name, pkg.InstalledVersion)
			pkg.InstalledSize = diskUsage(dir)
		}(i)
	}
	wg.Wait()

	return packages, nil
}

// getOutdatedMap returns a map of package name -> latest available version
// for all outdated packages.
func getOutdatedMap() (map[string]string, error) {
	out, err := runBrewStdout("outdated", "--json=v2")
	if err != nil {
		// brew outdated exits 1 when there ARE outdated packages — that is normal
		if out == "" {
			return nil, fmt.Errorf("brew outdated failed: %w", err)
		}
	}

	var result outdatedJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil, fmt.Errorf("parse outdated json: %w", err)
	}

	m := make(map[string]string)
	for _, f := range result.Formulae {
		m[f.Name] = f.CurrentVersion
	}
	return m, nil
}

// GetOutdated returns only the outdated packages.
func GetOutdated() ([]Package, error) {
	out, err := runBrewStdout("outdated", "--json=v2")
	if err != nil && out == "" {
		return nil, fmt.Errorf("brew outdated failed: %w", err)
	}

	var result outdatedJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		return nil, fmt.Errorf("parse outdated json: %w", err)
	}

	packages := make([]Package, 0, len(result.Formulae))
	for _, f := range result.Formulae {
		installed := ""
		if len(f.InstalledVersions) > 0 {
			installed = f.InstalledVersions[len(f.InstalledVersions)-1]
		}
		packages = append(packages, Package{
			Name:             f.Name,
			InstalledVersion: installed,
			LatestVersion:    f.CurrentVersion,
			Outdated:         true,
		})
	}
	return packages, nil
}

// GetInfo returns detailed information about a single package.
func GetInfo(name string) (*InfoResult, error) {
	out, err := runBrewStdout("info", "--json=v2", name)
	if err != nil && out == "" {
		return nil, fmt.Errorf("brew info failed: %w", err)
	}

	var raw brewInfoJSON
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, fmt.Errorf("parse info json: %w", err)
	}

	if len(raw.Formulae) == 0 {
		return nil, fmt.Errorf("package %q not found", name)
	}

	f := raw.Formulae[0]
	conflicts := make([]string, 0, len(f.Conflicts))
	for _, c := range f.Conflicts {
		conflicts = append(conflicts, c.Name)
	}

	installed := make([]InstalledVersion, 0, len(f.Installed))
	for _, iv := range f.Installed {
		installed = append(installed, InstalledVersion{Version: iv.Version})
	}

	// Disk size from cellar
	installedVer := ""
	if len(installed) > 0 {
		installedVer = installed[len(installed)-1].Version
	}
	size := diskUsage(filepath.Join(cellarPrefix(), f.Name, installedVer))

	return &InfoResult{
		Name:          f.Name,
		FullName:      f.FullName,
		Tap:           f.Tap,
		Version:       f.Versions.Stable,
		Desc:          f.Desc,
		Homepage:      f.Homepage,
		License:       f.License,
		Dependencies:  f.Dependencies,
		Conflicts:     conflicts,
		Installed:     installed,
		InstalledSize: size,
	}, nil
}

// Upgrade runs `brew upgrade <name>` and returns the combined output.
func Upgrade(name string) (string, error) {
	return runBrewCommand("upgrade", name)
}

// UpgradeAll runs `brew upgrade` for all outdated packages.
func UpgradeAll() (string, error) {
	return runBrewCommand("upgrade")
}

// Uninstall runs `brew uninstall <name>`.
func Uninstall(name string) (string, error) {
	return runBrewCommand("uninstall", name)
}

// Search runs `brew search --formula --json=v2 <query>` and returns matching names+descriptions.
func Search(query string) ([]Package, error) {
	out, err := runBrewStdout("search", "--formula", "--json=v2", query)
	if err != nil && out == "" {
		return nil, fmt.Errorf("brew search failed: %w", err)
	}

	var raw brewSearchJSON
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		// Fallback: plain-text search output
		return parsePlainSearch(out), nil
	}

	results := make([]Package, 0, len(raw.Formulae))
	for _, f := range raw.Formulae {
		results = append(results, Package{
			Name:          f.Name,
			LatestVersion: f.Desc, // repurposing LatestVersion as description in search results
		})
	}
	return results, nil
}

// Install runs `brew install <name>`.
func Install(name string) (string, error) {
	return runBrewCommand("install", name)
}

// Doctor runs `brew doctor` and returns the output.
func Doctor() (string, error) {
	return runBrewCommand("doctor")
}

// --- helpers ---

var (
	cellarOnce   sync.Once
	cellarCached string
)

// cellarPrefix returns the path to the Homebrew Cellar directory (cached).
func cellarPrefix() string {
	cellarOnce.Do(func() {
		out, err := exec.Command("brew", "--cellar").Output()
		if err != nil {
			cellarCached = "/opt/homebrew/Cellar"
			return
		}
		cellarCached = strings.TrimSpace(string(out))
	})
	return cellarCached
}

func parsePlainSearch(out string) []Package {
	var pkgs []Package
	for _, line := range parseLines(out) {
		pkgs = append(pkgs, Package{Name: line})
	}
	return pkgs
}

func parseLines(s string) []string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// diskUsage returns a human-readable disk usage string for the given path,
// e.g. "54M", "1.2G". Returns "—" if the path doesn't exist or du fails.
func diskUsage(dir string) string {
	if dir == "" {
		return "—"
	}
	out, err := exec.Command("du", "-sh", dir).Output()
	if err != nil {
		return "—"
	}
	// du output format: "<size>\t<path>"
	parts := strings.SplitN(strings.TrimSpace(string(out)), "\t", 2)
	if len(parts) == 0 {
		return "—"
	}
	return strings.TrimSpace(parts[0])
}
