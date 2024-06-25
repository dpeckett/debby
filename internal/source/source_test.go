// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package source_test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"testing"

	latestconfig "github.com/dpeckett/debby/internal/config/v1alpha1"
	"github.com/dpeckett/debby/internal/source"
	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

func TestSource(t *testing.T) {
	slog.SetDefault(slogt.New(t))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	resultCh := make(chan runMirrorResult, 1)
	t.Cleanup(func() {
		close(resultCh)
	})

	go runDebianMirror(ctx, resultCh)

	mirrorResult := <-resultCh
	require.NoError(t, mirrorResult.err)

	s, err := source.NewSource(ctx, http.DefaultClient, latestconfig.SourceConfig{
		URL:      fmt.Sprintf("http://%s/debian", mirrorResult.addr.String()),
		SignedBy: "../../testdata/archive-key-12.asc",
	})
	require.NoError(t, err)

	components, err := s.Components(ctx, arch.MustParse("amd64"))
	require.NoError(t, err)

	require.Len(t, components, 2)
	require.Equal(t, "main", components[0].Name)
	require.Equal(t, "all", components[0].Arch.String())
	require.Equal(t, "main", components[1].Name)
	require.Equal(t, "amd64", components[1].Arch.String())

	packageList, err := components[1].Packages(ctx)
	require.NoError(t, err)

	require.Len(t, packageList, 63408)
}

type runMirrorResult struct {
	err  error
	addr net.Addr
}

func runDebianMirror(ctx context.Context, result chan runMirrorResult) {
	mux := http.NewServeMux()

	mux.HandleFunc("/debian/dists/stable/InRelease", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../testdata/InRelease")
	})

	mux.HandleFunc("/debian/dists/stable/main/binary-amd64/Packages.gz", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../testdata/Packages.gz")
	})

	srv := &http.Server{
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		result <- runMirrorResult{err: err}
		return
	}

	result <- runMirrorResult{addr: lis.Addr()}

	go func() {
		<-ctx.Done()

		srv.Shutdown(context.Background())
	}()

	if err := srv.Serve(lis); err != nil {
		return
	}
}
