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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/adrg/xdg"
	"github.com/dpeckett/debby/internal/config"
	latestconfig "github.com/dpeckett/debby/internal/config/v1alpha1"
	"github.com/dpeckett/debby/internal/packages"
	"github.com/dpeckett/debby/internal/source"
	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/util"
	"github.com/gregjones/httpcache"
	"github.com/urfave/cli/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/sync/errgroup"
)

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
				Value: runtime.GOARCH,
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

			cache, err := util.NewDiskCache(cacheDir, "http")
			if err != nil {
				return fmt.Errorf("failed to create disk cache: %w", err)
			}

			httpClient := &http.Client{
				Transport: httpcache.NewTransport(cache),
			}

			targetArch, err := arch.Parse(c.String("arch"))
			if err != nil {
				return fmt.Errorf("failed to parse target architecture: %w", err)
			}

			logger.Info("Loading packages")

			packageDB, err := loadPackages(c.Context, httpClient, conf, targetArch)
			if err != nil {
				return err
			}

			res := packages.NewResolver(packageDB)

			selectedDB, err := res.Resolve(conf.Contents.Packages)
			if err != nil {
				return err
			}

			logger.Info("Selected", slog.Int("count", selectedDB.Len()))

			selectedDB.ForEach(func(pkg types.Package) error {
				if !pkg.IsVirtual {
					fmt.Println(pkg.Package, pkg.Version)
				}
				return nil
			})

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("Error", slog.Any("error", err))
		os.Exit(1)
	}
}

func loadPackages(ctx context.Context, httpClient *http.Client, conf *latestconfig.Config, targetArch arch.Arch) (*packages.PackageDB, error) {
	var p *mpb.Progress

	// Is the logger debug level?
	if !slog.Default().Enabled(ctx, slog.LevelDebug) {
		p = mpb.NewWithContext(ctx)
		defer p.Shutdown()
	}

	var componentsMu sync.Mutex
	var components []source.Component

	{
		g, ctx := errgroup.WithContext(ctx)

		var bar *mpb.Bar
		if p != nil {
			bar = p.AddBar(int64(len(conf.Contents.Sources)),
				mpb.PrependDecorators(
					decor.Name("Source: "),
					decor.CountersNoUnit("%d / %d"),
				),
				mpb.AppendDecorators(
					decor.Percentage(),
				),
			)
		}

		for _, sourceConf := range conf.Contents.Sources {
			sourceConf := sourceConf

			g.Go(func() error {
				defer func() {
					if bar != nil {
						bar.Increment()
					}
				}()

				s, err := source.NewSource(ctx, httpClient, sourceConf)
				if err != nil {
					return fmt.Errorf("failed to create source: %w", err)
				}

				sourceComponents, err := s.Components(ctx, targetArch)
				if err != nil {
					return fmt.Errorf("failed to get components: %w", err)
				}

				componentsMu.Lock()
				components = append(components, sourceComponents...)
				componentsMu.Unlock()

				return nil
			})
		}

		err := g.Wait()

		if bar != nil {
			bar.SetTotal(bar.Current(), true)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get components: %w", err)
		}
	}

	packageDB := packages.NewPackageDB()

	{
		g, ctx := errgroup.WithContext(ctx)

		var bar *mpb.Bar
		if p != nil {
			bar = p.AddBar(int64(len(components)),
				mpb.PrependDecorators(
					decor.Name("Component: "),
					decor.CountersNoUnit("%d / %d"),
				),
				mpb.AppendDecorators(
					decor.Percentage(),
				),
			)
		}

		for _, component := range components {
			component := component

			g.Go(func() error {
				defer func() {
					if bar != nil {
						bar.Increment()
					}
				}()

				componentPackages, err := component.Packages(ctx)
				if err != nil {
					return fmt.Errorf("failed to get packages: %w", err)
				}

				packageDB.AddAll(componentPackages)

				return nil
			})
		}

		err := g.Wait()

		if bar != nil {
			bar.SetTotal(bar.Current(), true)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get packages: %w", err)
		}
	}

	return packageDB, nil
}
