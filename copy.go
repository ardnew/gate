package main

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

var (
	ErrReadCanceled = errors.WithMessage(io.ErrClosedPipe,
		"timeout occurred before finishing read(s) from fifo")
	ErrWriteCanceled = errors.WithMessage(io.ErrClosedPipe,
		"timeout occurred before finishing write(s) to fifo")
)

type Pattern []Matcher

type Copier struct {
	rwc io.ReadWriteCloser
	ctx context.Context
	pat Pattern
}

type Matcher interface {
	Match([]byte) (bool, string)
}

type (
	Glob   struct{ glob.Glob }
	Regexp struct{ *regexp.Regexp }
)

func (g *Glob) Match(p []byte) (bool, string) {
	s := string(p)
	return g.Glob.Match(s), s
}

func (r *Regexp) Match(p []byte) (bool, string) {
	m := r.Regexp.Find(p)
	return m != nil, string(m)
}

func NewCopier(ctx context.Context, rwc io.ReadWriteCloser, pat ...string) *Copier {
	match := []Matcher{}
	for _, p := range pat {
		if len(p) >= 2 && p[0] == '/' && p[len(p)-1] == '/' {
			match = append(match, &Regexp{regexp.MustCompile(p[1 : len(p)-1])})
		} else {
			match = append(match, &Glob{glob.MustCompile(p)})
		}
	}
	return &Copier{rwc, ctx, match}
}

func (c *Copier) IsPatternDefined() bool {
	for _, p := range c.pat {
		if p != nil {
			return true
		}
	}
	return false
}

func (c *Copier) Match(p []byte) (bool, string) {
	for _, m := range c.pat {
		if ok, match := m.Match(p); ok {
			return true, match
		}
	}
	return false, ""
}

func (c *Copier) Read(p []byte) (n int, err error) {
	select {
	case <-c.ctx.Done():
		return 0, fmt.Errorf("%w: %w", ErrReadCanceled, c.ctx.Err())
	default:
		return c.rwc.Read(p)
	}
}

func (c *Copier) Write(p []byte) (n int, err error) {
	select {
	case <-c.ctx.Done():
		return 0, fmt.Errorf("%w: %w", ErrWriteCanceled, c.ctx.Err())
	default:
		return c.rwc.Write(p)
	}
}

func (c *Copier) Close() error { return c.rwc.Close() }
