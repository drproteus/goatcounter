// Copyright © Martin Tournoij – This file is part of GoatCounter and published
// under the terms of a slightly modified EUPL v1.2 license, which can be found
// in the LICENSE file or at https://license.goatcounter.com

package widgets

import (
	"context"
	"html/template"
	"strconv"
	"sync"

	"zgo.at/errors"
	"zgo.at/goatcounter/v2"
	"zgo.at/z18n"
	"zgo.at/zlog"
	"zgo.at/zstd/ztime"
)

type Pages struct {
	id     int
	loaded bool
	err    error
	html   template.HTML
	s      goatcounter.WidgetSettings

	RefsForPath      int64
	Style            string
	Limit, LimitRefs int
	Display          int
	More             bool
	Pages            goatcounter.HitLists
	Refs             goatcounter.HitStats
	Max              int
	Exclude          []int64
	Diff             []float64
}

func (w Pages) Name() string                         { return "pages" }
func (w Pages) Type() string                         { return "full-width" }
func (w Pages) Label(ctx context.Context) string     { return z18n.T(ctx, "label/paths|Paths overview") }
func (w *Pages) SetHTML(h template.HTML)             { w.html = h }
func (w Pages) HTML() template.HTML                  { return w.html }
func (w *Pages) SetErr(h error)                      { w.err = h }
func (w Pages) Err() error                           { return w.err }
func (w Pages) ID() int                              { return w.id }
func (w Pages) Settings() goatcounter.WidgetSettings { return w.s }

func (w *Pages) SetSettings(s goatcounter.WidgetSettings) {
	w.s = s
	if x := s["limit_pages"].Value; x != nil {
		w.Limit = int(x.(float64))
	}
	if x := s["limit_refs"].Value; x != nil {
		w.LimitRefs = int(x.(float64))
	}
	if x := s["key"].Value; x != nil {
		w.RefsForPath, _ = strconv.ParseInt(x.(string), 10, 64)
	}
	if x := s["style"].Value; x != nil {
		w.Style = x.(string)
	}
}

func (w *Pages) GetData(ctx context.Context, a Args) (bool, error) {
	if w.RefsForPath > 0 {
		err := w.Refs.ListRefsByPathID(ctx, w.RefsForPath, a.Rng, w.LimitRefs, a.Offset)
		return w.Refs.More, err
	}

	var (
		wg   sync.WaitGroup
		errs = errors.NewGroup(2)
	)
	if a.ShowRefs > 0 {
		wg.Add(1)
		go func() {
			defer zlog.Recover()
			defer wg.Done()
			errs.Append(w.Refs.ListRefsByPathID(ctx, a.ShowRefs, a.Rng, w.LimitRefs, a.Offset))
		}()
	}

	var err error
	w.Display, w.More, err = w.Pages.List(ctx, a.Rng, a.PathFilter, w.Exclude, w.Limit, a.Daily)
	errs.Append(err)

	if !goatcounter.MustGetUser(ctx).Settings.FewerNumbers {
		w.Diff, err = w.Pages.Diff(ctx, a.Rng, a.Rng)
		errs.Append(err)
	}

	wg.Wait()

	for _, p := range w.Pages {
		if p.Max > w.Max {
			w.Max = p.Max
		}
	}

	w.loaded = true
	return w.More, errs.ErrorOrNil()
}

func (w Pages) RenderHTML(ctx context.Context, shared SharedData) (string, any) {
	if w.RefsForPath > 0 {
		return "_dashboard_pages_refs.gohtml", struct {
			Context context.Context
			Site    *goatcounter.Site
			User    *goatcounter.User
			ID      int
			Loaded  bool
			Err     error

			Refs  goatcounter.HitStats
			Count int
		}{ctx, shared.Site, shared.User, w.id, w.loaded, w.err,
			w.Refs, shared.Total}
	}

	t := "_dashboard_pages"
	if w.Style == "text" {
		t += "_text"
	}
	if shared.RowsOnly {
		t += "_rows"
	}
	t += ".gohtml"

	// Correct max for chunked data in text view.
	if w.Style == "text" {
		w.Max = 0
		for _, p := range w.Pages {
			m, _ := goatcounter.ChunkStat(p.Stats)
			if m > w.Max {
				w.Max = m
			}
		}
	} else if len(w.Pages) > 0 && len(w.Pages[0].Stats) > 0 {
		// Set days in the future to -1; we filter this in the JS when rendering
		// the chart.
		// It's easier to do this here because JavaScript Date() has piss-poor
		// support for timezones.
		//
		// Only remove them if the last day is today: for everything else we
		// want to display the future as "greyed out".
		var (
			now   = ztime.Now().In(goatcounter.MustGetUser(ctx).Settings.Timezone.Loc())
			today = now.Format("2006-01-02")
			hour  = now.Hour()
		)
		if w.Pages[0].Stats[len(w.Pages[0].Stats)-1].Day == today {
			for i := range w.Pages {
				j := len(w.Pages[i].Stats) - 1
				w.Pages[i].Stats[j].Hourly = w.Pages[i].Stats[j].Hourly[:hour+1]
			}
		}
	}

	return t, struct {
		Context context.Context
		Site    *goatcounter.Site
		User    *goatcounter.User

		ID          int
		Loaded      bool
		Err         error
		Pages       goatcounter.HitLists
		Period      ztime.Range
		Daily       bool
		ForcedDaily bool
		Offset      int
		Max         int

		TotalDisplay int
		Total        int
		TotalEvents  int
		MorePages    bool

		Style    string
		Refs     goatcounter.HitStats
		ShowRefs int64
		Diff     []float64
	}{
		ctx, shared.Site, shared.User,
		w.id, w.loaded, w.err, w.Pages, shared.Args.Rng, shared.Args.Daily,
		shared.Args.ForcedDaily, 1, w.Max,
		w.Display, shared.Total, shared.TotalEvents, w.More,
		w.Style, w.Refs, shared.Args.ShowRefs,
		w.Diff,
	}
}
