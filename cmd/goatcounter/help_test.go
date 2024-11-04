// Copyright © Martin Tournoij – This file is part of GoatCounter and published
// under the terms of a slightly modified EUPL v1.2 license, which can be found
// in the LICENSE file or at https://license.goatcounter.com

package main

import (
	"testing"

	"zgo.at/zli"
)

func TestHelp(t *testing.T) {
	exit, _, out := zli.Test(t)

	{
		runCmd(t, exit, "help", "db")
		wantExit(t, exit, out, 0)
		if len(out.String()) < 1_000 {
			t.Error()
		}
		out.Reset()
	}

	{
		runCmd(t, exit, "help", "all")
		wantExit(t, exit, out, 0)
		if len(out.String()) < 20_000 {
			t.Error()
		}
		out.Reset()
	}
}
