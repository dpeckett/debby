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
	"sync"

	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/version"

	"github.com/google/btree"
)

// PackageDB is a package database.
type PackageDB struct {
	mu   sync.RWMutex
	tree *btree.BTreeG[types.Package]
}

// NewPackageDB creates a new package database.
func NewPackageDB() *PackageDB {
	return &PackageDB{
		tree: btree.NewG(2, func(a, b types.Package) bool {
			return a.Compare(b) < 0
		}),
	}
}

// Len returns the total number of packages in the database.
func (db *PackageDB) Len() int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.tree.Len()
}

// Add adds a package to the database.
func (db *PackageDB) Add(pkg types.Package) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.addPackage(pkg)
}

// AddAll adds multiple packages to the database.
func (db *PackageDB) AddAll(packageList []types.Package) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, pkg := range packageList {
		db.addPackage(pkg)
	}
}

func (db *PackageDB) addPackage(pkg types.Package) {
	db.tree.ReplaceOrInsert(pkg)

	// Does this package provide any virtual packages?
	if len(pkg.Provides.Relations) > 0 {
		for _, rel := range pkg.Provides.Relations {
			for _, possi := range rel.Possibilities {
				virtualPkg := types.Package{
					Package:   possi.Name,
					IsVirtual: true,
				}

				if possi.Version != nil {
					virtualPkg.Version = possi.Version.Version
				}

				// Do we already have a virtual package?
				if existing, ok := db.tree.Get(virtualPkg); ok {
					virtualPkg = existing
				}

				// Make sure the package is not already in the providers list.
				var found bool
				for _, provider := range virtualPkg.Providers {
					if provider.Compare(pkg) == 0 {
						found = true
						break
					}
				}

				// Add the package to the providers list (if it is not already there).
				if !found {
					virtualPkg.Providers = append(virtualPkg.Providers, pkg)
					db.tree.ReplaceOrInsert(virtualPkg)
				}
			}
		}
	}
}

// Remove removes a package from the database.
func (db *PackageDB) Remove(pkg types.Package) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.tree.Delete(pkg)

	// If the package provides any virtual packages, update the providers.
	if len(pkg.Provides.Relations) > 0 {
		for _, rel := range pkg.Provides.Relations {
			for _, possi := range rel.Possibilities {
				virtualPkg := types.Package{
					Package:   possi.Name,
					IsVirtual: true,
				}

				if possi.Version != nil {
					virtualPkg.Version = possi.Version.Version
				}

				if virtualPkg, ok := db.tree.Get(virtualPkg); ok {

					// Remove the package from the providers list.
					var updatedProviders []types.Package
					for _, provider := range virtualPkg.Providers {
						if provider.Compare(pkg) != 0 {
							updatedProviders = append(updatedProviders, provider)
						}
					}
					virtualPkg.Providers = updatedProviders

					// If there are no more providers, remove the virtual package.
					if len(virtualPkg.Providers) == 0 {
						db.tree.Delete(virtualPkg)
					} else {
						db.tree.ReplaceOrInsert(virtualPkg)
					}
				}
			}
		}
	}
}

// ForEach iterates over each package in the database.
// If the provided function returns an error, the iteration will stop.
func (db *PackageDB) ForEach(fn func(pkg types.Package) error) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var err error
	db.tree.Ascend(func(pkg types.Package) bool {
		err = fn(pkg)
		return err == nil
	})
	return err
}

// Get returns all packages that match the provided name.
func (db *PackageDB) Get(name string) (packageList []types.Package) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		packageList = append(packageList, pkg)

		return true
	})
	return
}

// StrictlyEarlier returns all packages that match the provided name and are
// strictly earlier than the provided version.
func (db *PackageDB) StrictlyEarlier(name string, version version.Version) (packageList []types.Package) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.tree.DescendLessOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		// Skip the package if it is the same version (since we want strictly earlier)
		if pkg.Version.Compare(version) == 0 {
			return true
		}

		packageList = append(packageList, pkg)

		return true
	})
	return
}

// EarlierOrEqual returns all packages that match the provided name and are
// earlier or equal to the provided version.
func (db *PackageDB) EarlierOrEqual(name string, version version.Version) (packageList []types.Package) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.tree.DescendLessOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		packageList = append(packageList, pkg)

		return true
	})
	return
}

// ExactlyEqual returns the package that matches the provided name and version.
func (db *PackageDB) ExactlyEqual(name string, version version.Version) (*types.Package, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var foundPackage *types.Package
	db.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		if pkg.Version.Compare(version) == 0 {
			foundPackage = &pkg
		}

		return false
	})
	return foundPackage, foundPackage != nil
}

// LaterOrEqual returns all packages that match the provided name and are
// later or equal to the provided version.
func (db *PackageDB) LaterOrEqual(name string, version version.Version) (packageList []types.Package) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		packageList = append(packageList, pkg)

		return true
	})
	return
}

// StrictlyLater returns all packages that match the provided name and are
// strictly later than the provided version.
func (db *PackageDB) StrictlyLater(name string, version version.Version) (packageList []types.Package) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	db.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		// Skip the package if it is the same version (since we want strictly later)
		if pkg.Version.Compare(version) == 0 {
			return true
		}

		packageList = append(packageList, pkg)

		return true
	})
	return
}
