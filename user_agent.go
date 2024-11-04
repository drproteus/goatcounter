// Copyright © Martin Tournoij – This file is part of GoatCounter and published
// under the terms of a slightly modified EUPL v1.2 license, which can be found
// in the LICENSE file or at https://license.goatcounter.com

package goatcounter

import (
	"context"

	"zgo.at/errors"
	"zgo.at/gadget"
	"zgo.at/isbot"
	"zgo.at/zcache"
	"zgo.at/zdb"
)

type UserAgent struct {
	UserAgent string
	Isbot     uint8
	BrowserID int64
	SystemID  int64
}

func (p *UserAgent) GetOrInsert(ctx context.Context) error {
	shortUA := gadget.ShortenUA(p.UserAgent)
	c, ok := cacheUA(ctx).Get(p.UserAgent)
	if ok {
		*p = c.(UserAgent)
		cacheUA(ctx).Touch(shortUA, zcache.DefaultExpiration)
		return nil
	}

	var (
		ua      = gadget.ParseUA(p.UserAgent)
		browser Browser
		system  System
	)

	err := browser.GetOrInsert(ctx, ua.BrowserName, ua.BrowserVersion)
	if err != nil {
		return errors.Wrap(err, "UserAgent.GetOrInsert")
	}
	p.BrowserID = browser.ID

	err = system.GetOrInsert(ctx, ua.OSName, ua.OSVersion)
	if err != nil {
		return errors.Wrap(err, "UserAgent.GetOrInsert")
	}
	p.SystemID = system.ID

	p.Isbot = uint8(isbot.UserAgent(p.UserAgent))

	cacheUA(ctx).SetDefault(shortUA, *p)
	return nil
}

type Browser struct {
	ID      int64  `db:"browser_id"`
	Name    string `db:"name"`
	Version string `db:"version"`
}

func (b *Browser) GetOrInsert(ctx context.Context, name, version string) error {
	k := name + version
	c, ok := cacheBrowsers(ctx).Get(k)
	if ok {
		*b = c.(Browser)
		cacheBrowsers(ctx).Touch(k, zcache.DefaultExpiration)
		return nil
	}

	b.Name = name
	b.Version = version

	err := zdb.Get(ctx, &b.ID,
		`select browser_id from browsers where name=$1 and version=$2`,
		name, version)
	if zdb.ErrNoRows(err) {
		b.ID, err = zdb.InsertID(ctx, "browser_id",
			`insert into browsers (name, version) values ($1, $2)`,
			name, version)
	}
	if err != nil {
		return errors.Wrapf(err, "Browser.GetOrInsert %q %q", name, version)
	}
	cacheBrowsers(ctx).SetDefault(k, *b)
	return nil
}

type System struct {
	ID      int64  `db:"system_id"`
	Name    string `db:"name"`
	Version string `db:"version"`
}

func (s *System) GetOrInsert(ctx context.Context, name, version string) error {
	k := name + version
	c, ok := cacheSystems(ctx).Get(k)
	if ok {
		*s = c.(System)
		cacheSystems(ctx).Touch(k, zcache.DefaultExpiration)
		return nil
	}

	s.Name = name
	s.Version = version

	err := zdb.Get(ctx, &s.ID,
		`select system_id from systems where name=$1 and version=$2`,
		name, version)
	if zdb.ErrNoRows(err) {
		s.ID, err = zdb.InsertID(ctx, "system_id",
			`insert into systems (name, version) values ($1, $2)`,
			name, version)
	}
	if err != nil {
		return errors.Wrapf(err, "System.GetOrInsert %q %q", name, version)
	}
	cacheSystems(ctx).SetDefault(k, *s)
	return nil
}
