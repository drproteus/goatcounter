// Copyright © Martin Tournoij – This file is part of GoatCounter and published
// under the terms of a slightly modified EUPL v1.2 license, which can be found
// in the LICENSE file or at https://license.goatcounter.com

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"zgo.at/goatcounter/v2"
	"zgo.at/goatcounter/v2/cron"
	"zgo.at/zdb"
	"zgo.at/zli"
	"zgo.at/zstd/zslice"
	"zgo.at/zstd/ztest"
)

func startServer(ctx context.Context, t *testing.T, exit *zli.TestExit, dbc string) chan struct{} {
	var site goatcounter.Site
	err := site.ByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	cn := "test.example.com"
	site.Cname = &cn
	err = site.Update(ctx)
	if err != nil {
		t.Fatal(err)
	}

	key := goatcounter.APIToken{SiteID: 1, UserID: 1, Name: "test", Permissions: goatcounter.APIPermCount}
	err = key.Insert(ctx)
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("GOATCOUNTER_API_KEY", key.Token)

	ready := make(chan struct{}, 1)
	stop := make(chan struct{})
	go runCmdStop(t, exit, ready, stop, "serve",
		"-tls=http",
		"-db="+dbc,
		"-listen=localhost:9876",
		"-debug=all")
	<-ready

	err = goatcounter.Memstore.TestInit(zdb.MustGetDB(ctx))
	if err != nil {
		t.Fatal(err)
	}
	return stop
}

func runImport(ctx context.Context, t *testing.T, exit *zli.TestExit, args ...string) func() {
	runCmd(t, exit, "import", append([]string{
		"-site=http://test.localhost:9876",
		"-debug=all"}, args...)...)

	err := cron.TaskPersistAndStat()
	if err != nil {
		t.Fatal(err)
	}

	return runImportClean(ctx, t)
}

func runImportBg(ctx context.Context, t *testing.T, exit *zli.TestExit, args ...string) (chan struct{}, func()) {
	ready := make(chan struct{}, 1)
	stop := make(chan struct{})
	go runCmdStop(t, exit, ready, stop, "import", append([]string{
		"-site=http://test.localhost:9876",
		"-debug=all"}, args...)...)
	<-ready
	time.Sleep(500 * time.Millisecond) // Tiny sleep for delay between "ready" and start of loop.

	return stop, runImportClean(ctx, t)
}

func runImportClean(ctx context.Context, t *testing.T) func() {
	return func() {
		err := goatcounter.Memstore.TestInit(zdb.MustGetDB(ctx))
		if err != nil {
			t.Fatal(err)
		}

		var paths []int64
		err = zdb.Select(ctx, &paths, `select path_id from paths`)
		if err != nil {
			t.Fatal(err)
		}
		if len(paths) == 0 {
			return
		}
		err = (&goatcounter.Hits{}).Purge(ctx, paths)
		if err != nil {
			t.Fatal(err)
		}

		if zdb.SQLDialect(ctx) == zdb.DialectSQLite {
			err = zdb.Exec(ctx, `update sqlite_sequence set seq = 0 where name in ('hits', 'paths')`)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			err = zdb.Exec(ctx, `truncate hits, paths`)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func tmpFile(t *testing.T) *os.File {
	tmp := filepath.Join(t.TempDir(), "access_log")
	fp, err := os.Create(tmp)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { fp.Close() })
	return fp
}

func writeLines(t *testing.T, fp *os.File, lines ...string) {
	for _, line := range lines {
		_, err := fp.WriteString(line + "\n")
		if err != nil {
			t.Fatal(err)
		}
	}
	err := fp.Sync()
	if err != nil {
		t.Fatal(err)
	}
	// Give import some time to make sure it's processed.
	time.Sleep(1000 * time.Millisecond)
}

func TestImport(t *testing.T) {
	t.Skip()

	exit, _, out, ctx, dbc := startTest(t)
	_ = out

	stopServer := startServer(ctx, t, exit, dbc)

	t.Run("csv", func(t *testing.T) {
		defer runImport(ctx, t, exit, "./testdata/export.csv")()

		got := zdb.DumpString(ctx, `select * from hits`)
		want := `
			hit_id  site_id  path_id  session                           bot  ref             ref_scheme  size         location  first_visit  created_at
			1       1        1        00112233445566778899aabbccddef03  0                    NULL        1280,768,1   AR        1            2020-12-01 00:07:10
			2       1        2        00112233445566778899aabbccddef03  0                    NULL        1280,768,1   AR        1            2020-12-01 00:07:44
			3       1        3        00112233445566778899aabbccddef04  0    www.reddit.com  o           1680,1050,2  RO        1            2020-12-27 00:37:37`
		if d := ztest.Diff(got, want, ztest.DiffNormalizeWhitespace); d != "" {
			t.Error(d)
		}
	})

	t.Run("log", func(t *testing.T) {
		defer runImport(ctx, t, exit, "-format=combined", "./testdata/access_log")()

		got := zdb.DumpString(ctx, `select * from hits`)
		want := `
			hit_id  site_id  path_id  session                           bot  ref                         ref_scheme  size  location  first_visit  created_at
			1       1        1        00112233445566778899aabbccddef01  0    www.example.com/start.html  h                           1            2000-10-10 20:55:36
			2       1        1        00112233445566778899aabbccddef01  0                                NULL                        0            2000-10-10 20:55:36`
		if d := ztest.Diff(got, want, ztest.DiffNormalizeWhitespace); d != "" {
			t.Error(d)
		}
	})

	t.Run("log-follow-4", func(t *testing.T) {
		fp := tmpFile(t)
		stop, clean := runImportBg(ctx, t, exit, "-format=combined", "-follow", fp.Name())
		defer clean()

		writeLines(t, fp, zslice.Repeat(
			`127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /test.html HTTP/1.1" 200 2326 "http://www.example.com/start.html" "Mozilla/5.0"`,
			4)...)
		stop <- struct{}{}
		time.Sleep(1000 * time.Millisecond)
		err := cron.TaskPersistAndStat()
		if err != nil {
			t.Fatal(err)
		}

		got := zdb.DumpString(ctx, `select * from hits`)

		want := "hit_id  site_id  path_id  session                           bot  ref                         ref_scheme  size  location  first_visit  created_at\n"
		for i := 1; i < 5; i++ {
			want += fmt.Sprintf(
				"%-3d     1        1        00112233445566778899aabbccddef01  0    www.example.com/start.html  h                           0            2000-10-10 20:55:36\n",
				i)

			if i == 1 { // first_visit
				want = strings.Replace(want, "0            ", "1            ", 1)
			}
		}
		if d := ztest.Diff(got, want, ztest.DiffNormalizeWhitespace); d != "" {
			t.Error(d)
		}
	})

	t.Run("log-follow-100", func(t *testing.T) {
		fp := tmpFile(t)
		stop, clean := runImportBg(ctx, t, exit, "-format=combined", "-follow", fp.Name())
		defer clean()

		writeLines(t, fp, zslice.Repeat(
			`127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /test.html HTTP/1.1" 200 2326 "http://www.example.com/start.html" "Mozilla/5.0"`,
			100)...)
		stop <- struct{}{}
		time.Sleep(1000 * time.Millisecond)
		err := cron.TaskPersistAndStat()
		if err != nil {
			t.Fatal(err)
		}

		got := zdb.DumpString(ctx, `select * from hits`)
		want := "hit_id  site_id  path_id  session                           bot  ref                         ref_scheme  size  location  first_visit  created_at\n"
		for i := 1; i < 101; i++ {
			want += fmt.Sprintf(
				"%-3d     1        1        00112233445566778899aabbccddef01  0    www.example.com/start.html  h                           0            2000-10-10 20:55:36\n",
				i)

			if i == 1 { // first_visit
				want = strings.Replace(want, "0            ", "1            ", 1)
			}
		}
		if d := ztest.Diff(got, want, ztest.DiffNormalizeWhitespace); d != "" {
			t.Error(d)
		}
	})

	stopServer <- struct{}{}
	mainDone.Wait()
}
