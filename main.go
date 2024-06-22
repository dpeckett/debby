// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/adrg/xdg"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dpeckett/debby/internal/config"
	latestconfig "github.com/dpeckett/debby/internal/config/v1alpha1"
	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/util"
	"github.com/gregjones/httpcache"
	"github.com/ulikunitz/xz"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	defaultDistribution   = "stable"
	maxConcurrentRequests = 16
)

var defaultComponents = []string{"main"}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cacheDir, err := xdg.CacheFile("debby")
	if err != nil {
		logger.Error("Failed to get cache directory", slog.Any("error", err))
		os.Exit(1)
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		logger.Error("Failed to create cache directory", slog.Any("error", err))
		os.Exit(1)
	}

	sharedFlags := []cli.Flag{
		&cli.GenericFlag{
			Name:  "log-level",
			Usage: "Set the log verbosity level",
			Value: util.FromSlogLevel(slog.LevelInfo),
		},
	}

	initLogger := func(c *cli.Context) error {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: (*slog.Level)(c.Generic("log-level").(*util.LevelFlag)),
		}))

		return nil
	}

	app := &cli.App{
		Name:  "debby",
		Usage: "A declarative Debian base system builder",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Configuration file",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "arch",
				Usage: "Architecture to build for",
				Value: "amd64",
			},
		}, sharedFlags...),
		Before: initLogger,
		Action: func(c *cli.Context) error {
			confFile, err := os.Open(c.String("config"))
			if err != nil {
				return fmt.Errorf("failed to open config file: %w", err)
			}
			defer confFile.Close()

			conf, err := config.FromYAML(confFile)
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}

			cache, err := util.NewDiskCache(logger, cacheDir, "http")
			if err != nil {
				return fmt.Errorf("failed to create disk cache: %w", err)
			}

			httpClient := &http.Client{
				Transport: httpcache.NewTransport(cache),
			}

			keyring, err := newKeyring(logger, httpClient, conf.Contents.Keyring)
			if err != nil {
				return fmt.Errorf("failed to load keyring: %w", err)
			}

			availablePackages, err := getAvailablePackages(c.Context, logger,
				httpClient, keyring, conf, c.String("arch"))
			if err != nil {
				return fmt.Errorf("failed to get available packages: %w", err)
			}

			logger.Info("Available packages", slog.Int("count", len(availablePackages)))

			ids := make([]string, 0, len(availablePackages))
			for key := range availablePackages {
				ids = append(ids, key)
			}

			sort.Strings(ids)

			availablePackageList := make([]types.Package, 0, len(availablePackages))
			for _, key := range ids {
				availablePackageList = append(availablePackageList, availablePackages[key])
			}

			f, err := os.Create("packages.json")
			if err != nil {
				return fmt.Errorf("failed to create Packages file: %w", err)
			}
			defer f.Close()

			if err := json.NewEncoder(f).Encode(availablePackageList); err != nil {
				return fmt.Errorf("failed to encode Packages file: %w", err)
			}

			// TODO: go through the list of user selected packages and check if they are available.

			// TODO: then recursively evaluate the dependency tree to find all transitive dependencies.

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("Error", slog.Any("error", err))
		os.Exit(1)
	}
}

func getAvailablePackages(ctx context.Context, logger *slog.Logger,
	httpClient *http.Client, keyring openpgp.EntityList,
	conf *latestconfig.Config, arch string) (map[string]types.Package, error) {
	var availablePackagesMu sync.Mutex
	var availablePackages = make(map[string]types.Package)

	sem := semaphore.NewWeighted(int64(maxConcurrentRequests))
	g, ctx := errgroup.WithContext(ctx)

	for _, source := range conf.Contents.Sources {
		source := source

		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			distribution := defaultDistribution
			if source.Distribution != "" {
				distribution = source.Distribution
			}

			logger := logger.With(
				slog.String("sourceURL", source.URL),
				slog.String("distribution", distribution))

			sourceURL, err := url.Parse(source.URL)
			if err != nil {
				return fmt.Errorf("failed to parse source URL: %w", err)
			}

			inReleaseURL, err := url.Parse(sourceURL.String())
			if err != nil {
				return fmt.Errorf("failed to parse source URL: %w", err)
			}

			inReleaseURL.Path = path.Join(inReleaseURL.Path, "dists", distribution, "InRelease")

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, inReleaseURL.String(), nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %w", err)
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to download InRelease file: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to download InRelease file: %s", resp.Status)
			}

			decoder, err := deb822.NewDecoder(resp.Body, keyring)
			if err != nil {
				return fmt.Errorf("failed to create decoder: %w", err)
			}

			var release types.Release
			if err := decoder.Decode(&release); err != nil {
				return fmt.Errorf("failed to decode InRelease file: %w", err)
			}

			// Extract the SHA256 sums for files in the release.
			releaseSHA256Sums, err := release.SHA256Sums()
			if err != nil {
				return fmt.Errorf("failed to extract SHA256 sums: %w", err)
			}

			availableArchitectures := mapset.NewSet(release.Architectures...).
				Intersect(mapset.NewSet("all", arch))
			if availableArchitectures.Cardinality() == 0 {
				logger.Warn("No architectures available")
				return nil
			}

			components := mapset.NewSet(defaultComponents...)
			if len(source.Components) > 0 {
				components = mapset.NewSet(source.Components...)
			}

			availableComponents := components.Intersect(mapset.NewSet(release.Components...))
			if availableComponents.Cardinality() == 0 {
				logger.Warn("No components available")
				return nil
			}

			for _, component := range availableComponents.ToSlice() {
				component := component

				g.Go(func() error {
					if err := sem.Acquire(ctx, 1); err != nil {
						return err
					}
					defer sem.Release(1)

					logger := logger.With(slog.String("component", component))

					for _, arch := range availableArchitectures.ToSlice() {
						arch := arch

						g.Go(func() error {
							if err := sem.Acquire(ctx, 1); err != nil {
								return err
							}
							defer sem.Release(1)

							logger := logger.With(slog.String("arch", arch))

							logger.Info("Downloading package list")

							packagesURL, err := url.Parse(sourceURL.String())
							if err != nil {
								return fmt.Errorf("failed to parse source URL: %w", err)
							}

							packagesURL.Path = path.Join(packagesURL.Path, "dists", distribution, path.Join(component, "binary-"+arch, "Packages.xz"))

							req, err := http.NewRequestWithContext(ctx, http.MethodGet, packagesURL.String(), nil)
							if err != nil {
								return fmt.Errorf("failed to create request: %w", err)
							}

							resp, err := httpClient.Do(req)
							if err != nil {
								return fmt.Errorf("failed to download Packages file: %w", err)
							}
							defer resp.Body.Close()

							if resp.StatusCode != http.StatusOK {
								return fmt.Errorf("failed to download Packages file: %s", resp.Status)
							}

							hr := util.NewHashReader(resp.Body)

							xzReader, err := xz.NewReader(hr)
							if err != nil {
								return fmt.Errorf("failed to create xz reader: %w", err)
							}

							decoder, err := deb822.NewDecoder(xzReader, keyring)
							if err != nil {
								return fmt.Errorf("failed to create decoder: %w", err)
							}

							var packages []types.Package
							if err := decoder.Decode(&packages); err != nil {
								return fmt.Errorf("failed to decode Packages file: %w", err)
							}

							relativePackagesPath := path.Join(path.Base(component), "binary-"+arch, "Packages.xz")
							expectedSHA256Sum, ok := releaseSHA256Sums[relativePackagesPath]
							if !ok {
								return fmt.Errorf("no SHA256 sum for Packages file: %s", relativePackagesPath)
							}

							if subtle.ConstantTimeCompare(expectedSHA256Sum, hr.Sum()) != 1 {
								return fmt.Errorf("checksum mismatch for Packages file: %s", relativePackagesPath)
							}

							logger.Debug("Package list checksum verified", slog.String("url", packagesURL.String()))

							logger.Debug("Found packages in package list", slog.Int("count", len(packages)))

							availablePackagesMu.Lock()
							defer availablePackagesMu.Unlock()

							for _, pkg := range packages {
								id := pkg.ID()
								if _, exist := availablePackages[id]; !exist {
									pkgURL, err := url.Parse(sourceURL.String())
									if err != nil {
										return fmt.Errorf("failed to parse source URL: %w", err)
									}

									pkgURL.Path = path.Join(pkgURL.Path, pkg.Filename)

									pkg := pkg
									pkg.URL = pkgURL.String()

									availablePackages[id] = pkg
								}
							}

							return nil
						})
					}

					return nil
				})
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return availablePackages, nil
}

func newKeyring(logger *slog.Logger, httpClient *http.Client, keyring []string) (openpgp.EntityList, error) {
	if len(keyring) == 0 {
		return openpgp.EntityList{}, nil
	}

	var entities openpgp.EntityList
	for _, key := range keyring {
		var entity openpgp.EntityList

		// If the key is a URL, download it.
		if strings.Contains(key, "://") && !strings.HasPrefix(key, "file://") {
			logger.Debug("Downloading key", slog.String("url", key))

			resp, err := httpClient.Get(key)
			if err != nil {
				return nil, err
			}

			// ReadArmoredKeyRing() doesn't read the entire response body, so we need
			// to do it ourselves (so that response caching will work as expected).
			keyringData, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if err != nil {
				return nil, err
			}

			entity, err = openpgp.ReadArmoredKeyRing(bytes.NewReader(keyringData))
			_ = resp.Body.Close()
			if err != nil {
				return nil, err
			}
		} else { // If the key is a file, open it.
			logger.Debug("Reading key file", slog.String("path", key))

			f, err := os.Open(strings.TrimPrefix(key, "file://"))
			if err != nil {
				return nil, err
			}

			entity, err = openpgp.ReadArmoredKeyRing(f)
			_ = f.Close()
			if err != nil {
				return nil, err
			}
		}

		entities = append(entities, entity...)
	}

	return entities, nil
}
