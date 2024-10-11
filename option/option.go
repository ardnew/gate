package option

import (
	"fmt"
	"strings"

	"github.com/ardnew/gateproc/errs"
	flag "github.com/spf13/pflag"
)

type Global struct {
	*flag.FlagSet
	Fifo    *EnvVar
	Timeout *EnvVar
	Match   *EnvVar
}

func New(name string, usage func(string) string) *Global {
	glo := &Global{
		FlagSet: flag.NewFlagSet(name, flag.ContinueOnError),
		Fifo:    NewEnvVar(name, "fifo", "f", "`file` path to named fifo", "string"),
		Timeout: NewEnvVar(name, "timeout", "t", "`duration` to wait for named fifo", "string"),
		Match:   NewEnvVar(name, "match", "m", "(recv-only) close after `pattern`", "string"),
	}
	return glo.register(name, usage)
}

func (g *Global) Parse(args []string) (err error) {
	if err = g.FlagSet.Parse(args); err == nil {
		if g.Fifo.string == "" {
			err = errs.NoFifo
		}
	}
	return err
}

func (g *Global) register(name string, usage func(string) string) *Global {
	if g == nil || g.FlagSet == nil {
		return g
	}

	usageOpt := UsageRows{}
	usageEnv := UsageRows{}

	func(vs ...*EnvVar) {
		for _, v := range vs {
			if v != nil {
				g.StringVarP(&v.string, v.Syms.Long, v.Syms.Short, v.string, v.Desc)
				usageOpt = append(usageOpt, v.OptionUsage()...)
				usageEnv = append(usageEnv, v.EnvironmentUsage()...)
			}
		}
	}(g.Fifo, g.Timeout, g.Match)

	wide := maxWidth(usageOpt, usageEnv)
	g.Usage = func() {
		fmt.Printf(
			usage(name),
			strings.Join(formatUsage(usageOpt, wide), "\n"),
			strings.Join(formatUsage(usageEnv, wide), "\n"),
		)
	}

	return g
}
