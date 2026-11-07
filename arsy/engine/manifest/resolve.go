package manifest

import (
	"fmt"
	"sort"
	"strings"
)

// ResolveOrder returns module names in deterministic dependency order
// (dependencies before dependents). Inputs may be given in any order;
// output is stable for identical inputs.
func ResolveOrder(modules []*Manifest) ([]string, error) {
	byName := make(map[string]*Manifest, len(modules))
	names := make([]string, 0, len(modules))
	for _, m := range modules {
		if m == nil {
			return nil, fmt.Errorf("%w: nil module", ErrInvalidManifest)
		}
		if _, ok := byName[m.Name]; ok {
			return nil, fmt.Errorf("%w: duplicate module %q", ErrInvalidManifest, m.Name)
		}
		byName[m.Name] = m
		names = append(names, m.Name)
	}
	sort.Strings(names)

	state := make(map[string]int, len(modules)) // 0 unvisited, 1 visiting, 2 done
	order := make([]string, 0, len(modules))

	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		switch state[name] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("%w: dependency cycle: %s -> %s", ErrInvalidDependency, strings.Join(path, " -> "), name)
		}
		state[name] = 1
		m := byName[name]
		for _, dep := range m.Depends {
			depName := dep.Name
			target, ok := byName[depName]
			if !ok {
				return fmt.Errorf("%w: module %q depends on unknown module %q", ErrInvalidDependency, name, depName)
			}
			if err := checkVersionConstraint(target.Version, dep.Constraint); err != nil {
				return fmt.Errorf("%w: module %q requires %q %s but %s is provided", ErrInvalidDependency, name, depName, dep.Constraint, target.Version)
			}
			nextPath := append(append([]string{}, path...), name)
			if err := visit(depName, nextPath); err != nil {
				return err
			}
		}
		state[name] = 2
		order = append(order, name)
		return nil
	}

	for _, name := range names {
		if err := visit(name, nil); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// checkVersionConstraint verifies a semver version satisfies a constraint.
// Constraints may be space-separated (AND). Supported operators: =, >=, >, <, <=.
// Empty constraint accepts all.
func checkVersionConstraint(version, constraint string) error {
	constraint = strings.TrimSpace(constraint)
	if constraint == "" {
		return nil
	}
	for _, clause := range strings.Fields(constraint) {
		if err := checkSingleConstraint(version, clause); err != nil {
			return err
		}
	}
	return nil
}

func checkSingleConstraint(version, clause string) error {
	op := "="
	rest := clause
	for _, candidate := range []string{">=", "<=", "=", ">", "<"} {
		if strings.HasPrefix(clause, candidate) {
			op = candidate
			rest = strings.TrimSpace(strings.TrimPrefix(clause, candidate))
			break
		}
	}
	if rest == "" {
		return fmt.Errorf("%w: invalid constraint %q", ErrInvalidDependency, clause)
	}
	cmp := compareSemver(version, rest)
	satisfied := false
	switch op {
	case "=":
		satisfied = cmp == 0
	case ">=":
		satisfied = cmp >= 0
	case ">":
		satisfied = cmp > 0
	case "<":
		satisfied = cmp < 0
	case "<=":
		satisfied = cmp <= 0
	}
	if !satisfied {
		return fmt.Errorf("%w: version %s does not satisfy %q", ErrInvalidDependency, version, clause)
	}
	return nil
}

// compareSemver compares two MAJOR.MINOR.PATCH versions. Returns -1, 0, or 1.
func compareSemver(a, b string) int {
	va := parseSemverCore(strings.SplitN(a, "-", 2)[0])
	vb := parseSemverCore(strings.SplitN(b, "-", 2)[0])
	for i := 0; i < 3; i++ {
		if va[i] < vb[i] {
			return -1
		}
		if va[i] > vb[i] {
			return 1
		}
	}
	return 0
}

func parseSemverCore(s string) [3]int {
	var out [3]int
	parts := strings.Split(s, ".")
	for i := 0; i < len(parts) && i < 3; i++ {
		n := 0
		for _, c := range parts[i] {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		out[i] = n
	}
	return out
}
