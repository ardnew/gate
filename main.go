package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ardnew/gateproc/errs"
	"github.com/ardnew/gateproc/option"
	"github.com/containerd/fifo"
)

const version = "0.1.0"

const badExe = "%!s(BADEXE)"

func exeName() string {
	e, err := os.Executable()
	if err != nil {
		return badExe
	}
	return filepath.Base(e)
}

func usage(exe string) string {
	desc := `%[1]s version %[2]s usage:
NAME
	%[1]s

DESCRIPTION
	%[1]s uses a named fifo to enable one command to gate execution of another.

	The named fifo path is intialized with either the environment variable
	GATEPROC_FIFO or the command-line flag [-f|--fifo].
	It is removed after both the sender and receiver have closed the fifo.

	The sender closes the fifo after writing all data and signals EOF.

	By default, the receiver closes the fifo after reading all data and
	receiving EOF. The receiver can instead keep the fifo open indefinitely
	until it matches a user-provided pattern ("glob" or "/regexp/").
	Matching the pattern will close the fifo after receiving EOF so that
	the sender can finish writing gracefully.

	Currently, only line-buffered matching is supported.

	[NOTE] Command-line flags take precedence over environment variables.

SYNOPSIS
	%[1]s -              (send) stdin => named fifo
	%[1]s                (recv) named fifo => stdout

	%[1]s < file         (send) file => named fifo
	%[1]s > file         (recv) named fifo => file

	cmd | %[1]s          (send) cmd stdout => named fifo
	%[1]s | cmd          (recv) named fifo => cmd stdin

OPTIONS
%%s

ENVIRONMENT
%%s

`
	return fmt.Sprintf(desc, exeName(), version)
}

func main() {
	var cancel context.CancelFunc
	ctx := context.Background()

	opts := option.New(exeName(), usage)

	errs.Check(
		errs.Trace(opts.Parse(os.Args[1:]), "parse command-line options"), 1,
		func(err error) {
			if errs.IsHelpFlag(err) {
				os.Exit(0)
			}
		})

	if opts.Timeout.String() != "" {
		timeout, err := time.ParseDuration(opts.Timeout.String())
		errs.Check(errs.Trace(err, "parse timeout duration"), 1)
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if !isInteractive() || opts.Arg(0) == "-" {
		// send mode (write): stdin ⭆ fifo
		flag := syscall.O_WRONLY | syscall.O_CREAT | syscall.O_NONBLOCK
		w, err := fifo.OpenFifo(ctx, opts.Fifo.String(), flag, 0o600)
		errs.Check(errs.Trace(err, "open fifo for write"), 1)
		defer cleanup(w, opts.Fifo.String())

		_, err = io.Copy(NewCopier(ctx, w), os.Stdin)
		errs.Check(errs.Trace(err, "copy from stdin to fifo"), 1)
	} else {
		// recv mode (read): stdout ⭅ fifo
		flag := syscall.O_RDONLY | syscall.O_CREAT | syscall.O_NONBLOCK
		r, err := fifo.OpenFifo(ctx, opts.Fifo.String(), flag, 0o600)
		errs.Check(errs.Trace(err, "open fifo for read"), 1)
		defer cleanup(r, opts.Fifo.String())

		reader := NewCopier(ctx, r, opts.Match.String())
		if reader.IsPatternDefined() {
			scan := bufio.NewScanner(reader)
			for err == nil {
				if scan.Scan() {
					if ok, match := reader.Match(scan.Bytes()); ok {
						_, err = io.WriteString(os.Stdout, match)
						io.Copy(io.Discard, reader)
						break
					}
				} else if err = scan.Err(); nil == err {
					// EOF, restart scanner
					scan = bufio.NewScanner(reader)
				}
			}
		} else {
			_, err = io.Copy(os.Stdout, reader)
		}
		errs.Check(errs.Trace(err, "copy from fifo to stdout"), 1)
	}
}

func cleanup(rwc io.ReadWriteCloser, path string) {
	if rwc == nil || path == "" {
		return
	}
	_ = rwc.Close()
	if ok, err := fifo.IsFifo(path); err == nil && ok {
		_ = os.Remove(path)
	}
}
