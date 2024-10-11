package errs

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

var NoFifo = errors.New("no fifo path provided")

func IsHelpFlag(err error) bool {
	return errors.Is(err, flag.ErrHelp)
}

func Trace(err error, msg ...string) error {
	if len(msg) > 0 {
		return errors.Wrap(err, strings.Join(msg, "\n"))
	}
	return errors.WithStack(err)
}

func Check(err error, result int, onError ...func(error)) {
	if err != nil {
		for _, check := range onError {
			check(err)
		}
		if result >= 0 {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(result)
		}
	}
}
