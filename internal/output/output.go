package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Format int

const (
	FormatPlain Format = iota
	FormatJSON
)

type Printer struct {
	Format Format
	Out    io.Writer
	Err    io.Writer
}

func NewPrinter(format Format) *Printer {
	return &Printer{
		Format: format,
		Out:    os.Stdout,
		Err:    os.Stderr,
	}
}

func (p *Printer) Print(data any) error {
	switch p.Format {
	case FormatJSON:
		return p.printJSON(data)
	default:
		return p.printPlain(data)
	}
}

func (p *Printer) printJSON(data any) error {
	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (p *Printer) printPlain(data any) error {
	switch v := data.(type) {
	case string:
		fmt.Fprintln(p.Out, v)
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(p.Out, "%-30s %v\n", k, v[k])
		}
	case []any:
		for i, item := range v {
			if m, ok := item.(map[string]any); ok {
				if i > 0 {
					fmt.Fprintln(p.Out, strings.Repeat("-", 60))
				}
				keys := make([]string, 0, len(m))
				for k := range m {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Fprintf(p.Out, "%-30s %v\n", k, m[k])
				}
			} else {
				fmt.Fprintln(p.Out, item)
			}
		}
	default:
		fmt.Fprintln(p.Out, data)
	}
	return nil
}

func (p *Printer) Hint(msg string) {
	fmt.Fprintln(p.Err, msg)
}
