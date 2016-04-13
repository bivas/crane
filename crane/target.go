package crane

import (
	"fmt"
	"sort"
	"strings"
)

type Target struct {
	initial      []string
	dependencies []string
	affected     []string
}

// NewTarget receives the specified target
// and determines which containers should be targeted.
// The target might be extended depending whether the
// dynamic targets "dependencies" and/or "affected"
// are included in the targetFlag.
// Additionally, the target is sorted alphabetically.
func NewTarget(dependencyMap map[string]*Dependencies, targetFlag string) (target Target, err error) {

	targetParts := strings.Split(targetFlag, "+")
	targetName := targetParts[0]
	extendDependencies := false
	extendAffected := false
	for _, v := range targetParts[1:] {
		if v == "dependencies" || v == "d" {
			extendDependencies = true
		} else if v == "affected" || v == "a" {
			extendAffected = true
		} else {
			err = fmt.Errorf("Unknown target extension %s. Available options are 'dependencies'/'d' and 'affected'/'a'", v)
			return
		}
	}

	target = Target{
		initial:      []string{},
		dependencies: []string{},
		affected:     []string{},
	}

	initialTarget := cfg.ContainersForReference(targetName)
	for _, c := range initialTarget {
		if includes(allowed, c) {
			target.initial = append(target.initial, c)
		}
	}

	if extendDependencies {
		var (
			dependenciesSet = make(map[string]struct{})
			cascadingSeeds  = []string{}
		)
		// start from the explicitly targeted target
		for _, name := range target.initial {
			dependenciesSet[name] = struct{}{}
			cascadingSeeds = append(cascadingSeeds, name)
		}

		// Cascade until the dependency map has been fully traversed
		// according to the cascading flags.
		for len(cascadingSeeds) > 0 {
			nextCascadingSeeds := []string{}
			for _, seed := range cascadingSeeds {
				if dependencies, ok := dependencyMap[seed]; ok {
					// Queue direct dependencies if we haven't already considered them
					for _, name := range dependencies.All {
						if _, alreadyIncluded := dependenciesSet[name]; !alreadyIncluded {
							dependenciesSet[name] = struct{}{}
							nextCascadingSeeds = append(nextCascadingSeeds, name)
						}
					}
				}
			}
			cascadingSeeds = nextCascadingSeeds
		}

		for name := range dependenciesSet {
			if !includes(target.initial, name) {
				target.dependencies = append(target.dependencies, name)
			}
		}

		sort.Strings(target.dependencies)
	}

	if extendAffected {
		var (
			affected       = make(map[string]bool)
			cascadingSeeds = []string{}
		)
		// start from the explicitly targeted target
		for _, name := range target.initial {
			affected[name] = true
			cascadingSeeds = append(cascadingSeeds, name)
		}

		for len(cascadingSeeds) > 0 {
			nextCascadingSeeds := []string{}
			for _, seed := range cascadingSeeds {
				for name, dependencies := range dependencyMap {
					if _, alreadyConsidered := affected[name]; !alreadyConsidered {
						if dependencies.includes(seed) {
							if cfg.Container(name).Exists() {
								affected[name] = true
								nextCascadingSeeds = append(nextCascadingSeeds, name)
							} else {
								affected[name] = false
							}
						}
					}
				}
			}
			cascadingSeeds = nextCascadingSeeds
		}

		for name, included := range affected {
			if included && !includes(target.initial, name) {
				target.affected = append(target.affected, name)
			}
		}

		sort.Strings(target.affected)
	}

	return
}

// includes checks whether the given needle is
// included in the target
func (t Target) includes(needle string) bool {
	for _, name := range t.all() {
		if name == needle {
			return true
		}
	}
	return false
}

// Return all targeted containers, sorted alphabetically
func (t Target) all() []string {
	all := t.initial
	for _, name := range t.dependencies {
		all = append(all, name)
	}
	for _, name := range t.affected {
		all = append(all, name)
	}
	sort.Strings(all)
	return all
}
