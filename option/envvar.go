package option

import (
	"fmt"
	"os"
	"strings"
)

type EnvVar struct {
	string
	Ident string
	Syms  struct{ Long, Short string }
	Desc  string
	VType string
}

func NewEnvVar(owner, long, short, desc, vtype string) *EnvVar {
	ident := strings.ToUpper(owner + "_" + long)
	return &EnvVar{
		string: os.Getenv(ident),
		Ident:  ident,
		Syms: struct{ Long, Short string }{
			Long: long, Short: short,
		},
		Desc:  desc,
		VType: vtype,
	}
}

func (e *EnvVar) String() string {
	return e.string
}

func (e *EnvVar) Syntax(param string) string {
	return fmt.Sprintf("-%s|--%s %s", e.Syms.Short, e.Syms.Long, strings.ToUpper(param))
}

func (e *EnvVar) Default() string {
	const indent = 2
	if def := e.String(); def != "" {
		return fmt.Sprintf("%*s(default %q)", indent, "", def)
	}
	return ""
}

func (e *EnvVar) OptionUsage() UsageRows {
	arg, use := e.UnquoteUsage()
	cols := UsageRows{{e.Syntax(arg), use}}
	if def := e.Default(); def != "" {
		cols = append(cols, UsageCols{"", def})
	}
	return cols
}

func (e *EnvVar) EnvironmentUsage() UsageRows {
	_, use := e.UnquoteUsage()
	return UsageRows{{e.Ident, use}}
}

func (e *EnvVar) UnquoteUsage() (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = e.Desc
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	name = e.VType
	return
}

type (
	UsageCols  [2]string
	UsageRows  []UsageCols
	usageWidth [len(UsageCols{})]int
)

func maxWidth(rows ...UsageRows) usageWidth {
	widths := usageWidth{}
	for _, row := range rows {
		for _, cols := range row {
			for i, col := range cols {
				if n := len(col); n > widths[i] {
					widths[i] = n
				}
			}
		}
	}
	return widths
}

func formatUsage(rows UsageRows, width usageWidth) []string {
	formatted := []string{}
	for _, row := range rows {
		formatted = append(formatted,
			fmt.Sprintf("\t%-*s\t\t%-*s", width[0], row[0], width[1], row[1]),
		)
	}
	return formatted
}
