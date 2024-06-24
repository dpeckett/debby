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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/adrg/xdg"
	"github.com/dpeckett/debby/internal/config"
	latestconfig "github.com/dpeckett/debby/internal/config/v1alpha1"
	"github.com/dpeckett/debby/internal/packages"
	"github.com/dpeckett/debby/internal/source"
	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/types/version"
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

			cache, err := util.NewDiskCache(logger, cacheDir, "http")
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

			packageCollection, err := loadPackages(c.Context, logger, httpClient, conf, targetArch)
			if err != nil {
				return err
			}

			bash, ok := packageCollection.ExactlyEqual("bash", version.MustParse("5.2.15-2+b2"))
			if !ok {
				return errors.New("bash not found")
			}

			fmt.Printf("Bash results: %+v\n", bash)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error("Error", slog.Any("error", err))
		os.Exit(1)
	}
}

func loadPackages(ctx context.Context, logger *slog.Logger, httpClient *http.Client, conf *latestconfig.Config, targetArch arch.Arch) (*packages.PackageCollection, error) {
	p := mpb.NewWithContext(ctx)
	defer p.Shutdown()

	var componentsMu sync.Mutex
	var components []source.Component

	{
		g, ctx := errgroup.WithContext(ctx)

		bar := p.AddBar(int64(len(conf.Contents.Sources)),
			mpb.PrependDecorators(
				decor.Name("Source: "),
				decor.CountersNoUnit("%d / %d"),
			),
			mpb.AppendDecorators(
				decor.Percentage(),
			),
		)

		for _, sourceConf := range conf.Contents.Sources {
			sourceConf := sourceConf

			g.Go(func() error {
				defer bar.Increment()

				keyring, err := loadKeyring(logger, httpClient, sourceConf.SignedBy)
				if err != nil {
					return fmt.Errorf("failed to load keyring for source: %w", err)
				}

				s, err := source.NewSource(logger, httpClient, keyring, sourceConf)
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
		bar.SetTotal(bar.Current(), true)
		if err != nil {
			return nil, fmt.Errorf("failed to get components: %w", err)
		}
	}

	packageCollection := packages.NewPackageCollection()

	{
		g, ctx := errgroup.WithContext(ctx)

		bar := p.AddBar(int64(len(components)),
			mpb.PrependDecorators(
				decor.Name("Component: "),
				decor.CountersNoUnit("%d / %d"),
			),
			mpb.AppendDecorators(
				decor.Percentage(),
			),
		)

		for _, component := range components {
			component := component

			g.Go(func() error {
				defer bar.Increment()

				componentPackages, err := component.Packages(ctx)
				if err != nil {
					return fmt.Errorf("failed to get packages: %w", err)
				}

				packageCollection.AddAll(componentPackages)

				return nil
			})
		}

		err := g.Wait()
		bar.SetTotal(bar.Current(), true)
		if err != nil {
			return nil, fmt.Errorf("failed to get packages: %w", err)
		}
	}

	return packageCollection, nil
}

func loadKeyring(logger *slog.Logger, httpClient *http.Client, key string) (openpgp.EntityList, error) {
	if len(key) == 0 {
		return openpgp.EntityList{}, nil
	}

	// If the key is a URL, download it.
	if strings.Contains(key, "://") {
		logger.Debug("Downloading key", slog.String("url", key))

		keyURL, err := url.Parse(key)
		if err != nil {
			return nil, err
		}

		if keyURL.Scheme != "https" {
			return nil, errors.New("key URL must be HTTPS")
		}

		resp, err := httpClient.Get(keyURL.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// ReadArmoredKeyRing() doesn't read the entire response body, so we need
		// to do it ourselves (so that response caching will work as expected).
		keyringData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return openpgp.ReadArmoredKeyRing(bytes.NewReader(keyringData))
	} else { // If the key is a file, open it.
		logger.Debug("Reading key file", slog.String("path", key))

		f, err := os.Open(key)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return openpgp.ReadArmoredKeyRing(f)
	}
}
