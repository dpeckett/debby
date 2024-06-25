// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package packages

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/dependency"
	"github.com/dpeckett/debby/internal/types/version"
)

// Resolver resolves package dependencies.
type Resolver struct {
	logger    *slog.Logger
	packageDB *PackageDB
}

// NewResolver creates a new resolver.
func NewResolver(logger *slog.Logger, packageDB *PackageDB) *Resolver {
	return &Resolver{
		logger:    logger,
		packageDB: packageDB,
	}
}

// Resolve resolves the dependencies of a list of packages, specified as a list
// of package name and optional version strings.
func (r *Resolver) Resolve(packageNameVersions []string) (*PackageDB, error) {
	requestedPackages := map[string]*version.Version{}
	candidateDB := NewPackageDB()

	for _, packageNameVersion := range packageNameVersions {
		parts := strings.SplitN(packageNameVersion, "=", 2)
		name := parts[0]

		var packageVersion *version.Version
		if len(parts) > 1 {
			v, err := version.Parse(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid version: %s: %w", parts[1], err)
			}

			packageVersion = &v
		}
		requestedPackages[name] = packageVersion

		if packageVersion != nil {
			pkg, exists := r.packageDB.ExactlyEqual(name, *packageVersion)
			if !exists {
				return nil, fmt.Errorf("unable to locate package: %s", packageNameVersion)
			}

			candidateDB.Add(*pkg)
		} else {
			packageList := r.packageDB.Get(name)
			if len(packageList) == 0 {
				return nil, fmt.Errorf("unable to locate package: %s", packageNameVersion)
			}

			candidateDB.AddAll(packageList)
		}
	}

	r.logger.Debug("Found initial candidates", slog.Int("count", candidateDB.Len()))

	r.logger.Debug("Building dependency tree")

	var queue []types.Package
	candidateDB.ForEach(func(pkg types.Package) error {
		queue = append(queue, pkg)
		return nil
	})

	visited := map[string]bool{}
	for len(queue) > 0 {
		pkg := queue[0]
		queue = queue[1:]

		id := pkg.ID()
		if visited[id] {
			continue
		}
		visited[id] = true

		deps, err := r.getDependencies(r.packageDB, candidateDB, pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to get dependencies for package %s: %w", pkg.Package, err)
		}

		for _, pkg := range deps {
			if !visited[pkg.ID()] {
				candidateDB.Add(pkg)
				queue = append(queue, pkg)
			}
		}
	}

	r.logger.Debug("Pruning candidates with unsatisfiable dependencies")

	r.pruneUnsatisfied(candidateDB)

	// If there are multiple versions of the same package, select the newest
	// version.
	// TODO: shell out to a SAT solver to find the optimal solution.
	// TODO: handle conflicts etc.
	r.logger.Debug("Selecting newest version of each package")

	var selectedDB = NewPackageDB()
	_ = candidateDB.ForEach(func(pkg types.Package) error {
		// If the package is requested with an explicit version, only select it if the version matches.
		if packageVersion, ok := requestedPackages[pkg.Package]; ok && packageVersion != nil {
			if pkg.Version.Compare(*packageVersion) == 0 {
				selectedDB.Add(pkg)
			}
			return nil
		}

		// If the package is already selected, only replace it if the new version
		// is higher.
		if existing := selectedDB.Get(pkg.Package); len(existing) > 0 {
			if pkg.Version.Compare(existing[0].Version) > 0 {
				selectedDB.Remove(existing[0])
				selectedDB.Add(pkg)
			}
		} else {
			selectedDB.Add(pkg)
		}

		return nil
	})

	r.pruneUnsatisfied(selectedDB)

	r.logger.Debug("Confirming requested packages are still selected")

	// Confirm all the requested packages are still selected.
	for name, version := range requestedPackages {
		if version != nil {
			if _, exists := selectedDB.ExactlyEqual(name, *version); !exists {
				return nil, fmt.Errorf("requested package %s=%s is not selected", name, version)
			}
		} else {
			if len(selectedDB.Get(name)) == 0 {
				return nil, fmt.Errorf("requested package %s is not selected", name)
			}
		}
	}

	return selectedDB, nil
}

// pruneUnsatisfied iteratively removes candidates with unsatisfiable dependencies.
func (r *Resolver) pruneUnsatisfied(candidateDB *PackageDB) {
	for {
		var pruneList []types.Package
		_ = candidateDB.ForEach(func(pkg types.Package) error {
			if _, err := r.getDependencies(candidateDB, candidateDB, pkg); err != nil {
				r.logger.Debug("Pruning unsatisfiable candidate",
					slog.String("name", pkg.Package), slog.String("version", pkg.Version.String()),
					slog.Any("error", err))

				pruneList = append(pruneList, pkg)
			}

			return nil
		})

		for _, pkg := range pruneList {
			candidateDB.Remove(pkg)
		}

		if len(pruneList) == 0 {
			break
		}
	}
}

func (r *Resolver) getDependencies(packageDB, candidateDB *PackageDB, pkg types.Package) ([]types.Package, error) {
	var dependencies []types.Package

	var relations []dependency.Relation
	relations = append(relations, pkg.PreDepends.Relations...)
	relations = append(relations, pkg.Depends.Relations...)

	for _, rel := range relations {
		var resolved bool
		for _, possi := range rel.Possibilities {
			// TODO: implement all of the remainder of the debian relation constraints.

			var packageList []types.Package
			if possi.Version != nil {
				switch possi.Version.Operator {
				case "<<":
					packageList = packageDB.EarlierOrEqual(possi.Name, possi.Version.Version)
				case "<=":
					packageList = packageDB.EarlierOrEqual(possi.Name, possi.Version.Version)
				case "=":
					pkg, exists := packageDB.ExactlyEqual(possi.Name, possi.Version.Version)
					if exists {
						packageList = []types.Package{*pkg}
					}
				case ">=":
					packageList = packageDB.LaterOrEqual(possi.Name, possi.Version.Version)
				case ">>":
					packageList = packageDB.LaterOrEqual(possi.Name, possi.Version.Version)
				default:
					return nil, fmt.Errorf("unknown version relation operator: %s", possi.Version.Operator)
				}
			} else {
				packageList = packageDB.Get(possi.Name)
			}

			// Resolve virtual packages.
			var resolvedPackages []types.Package
			for _, pkg := range packageList {
				if pkg.IsVirtual {
					if resolvedPkg, err := r.resolveVirtualPackage(packageDB, candidateDB, pkg); err == nil {
						resolvedPackages = append(resolvedPackages, resolvedPkg)
					} else {
						r.logger.Debug("Failed to resolve virtual package",
							slog.String("name", pkg.Package), slog.String("version", pkg.Version.String()),
							slog.Any("error", err))
					}
				} else {
					resolvedPackages = append(resolvedPackages, pkg)
				}
			}

			if len(resolvedPackages) > 0 {
				dependencies = append(dependencies, resolvedPackages...)
				resolved = true
				break
			}
		}

		if !resolved {
			return nil, fmt.Errorf("unsatisfiable dependency: %s", rel.String())
		}
	}

	return dependencies, nil
}

func (r *Resolver) resolveVirtualPackage(packageDB, candidateDB *PackageDB, virtualPkg types.Package) (types.Package, error) {
	var virtualProviders []types.Package
	for _, provider := range virtualPkg.Providers {
		if pkg, exists := packageDB.ExactlyEqual(provider.Package, provider.Version); exists {
			virtualProviders = append(virtualProviders, *pkg)
		}
	}

	if len(virtualProviders) == 0 {
		return types.Package{}, fmt.Errorf("unsatisfiable dependency: %s", virtualPkg.Package)
	} else if len(virtualProviders) == 1 {
		return virtualProviders[0], nil
	} else {
		// Has a provider already been selected? Eg. its part of the candidate list.
		for _, pkg := range virtualProviders {
			if _, exists := candidateDB.ExactlyEqual(pkg.Package, pkg.Version); exists {
				return pkg, nil
			}
		}

		// Is one of the providers marked as required priority?
		for _, pkg := range virtualProviders {
			if pkg.Priority == "required" {
				return pkg, nil
			}
		}

		return types.Package{}, fmt.Errorf("virtual package with multiple installation candidates: %s", virtualPkg.Package)
	}
}
