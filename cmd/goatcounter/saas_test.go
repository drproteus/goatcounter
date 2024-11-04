// Copyright © Martin Tournoij – This file is part of GoatCounter and published
// under the terms of a slightly modified EUPL v1.2 license, which can be found
// in the LICENSE file or at https://license.goatcounter.com

package main

import (
	"io"
	"net/http"
	"testing"
)

func TestSaas(t *testing.T) {
	exit, _, _, _, dbc := startTest(t)

	ready := make(chan struct{}, 1)
	stop := make(chan struct{})
	go func() {
		runCmdStop(t, exit, ready, stop, "saas",
			"-db="+dbc,
			"-debug=all",
			"-domain=goatcounter.com,a.a",
			"-listen=localhost:31874",
			"-tls=http")
	}()
	<-ready

	resp, err := http.Get("http://localhost:31874/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Errorf("status %d: %s", resp.StatusCode, b)
	}
	if len(b) < 100 {
		t.Errorf("%s", b)
	}

	stop <- struct{}{}
	mainDone.Wait()
}
