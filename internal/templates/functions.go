package templates

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
	"strconv"
	"slices"
)

var tmplFuncs = template.FuncMap{
	"now": time.Now,
	"formatSecs": formatSecs,
	"parseInt": func(s string) (int, error) {
		return strconv.Atoi(s)
	},
	"iterate": func(n int) []int {
		var result []int
		for i := 1; i <= n; i++ {
			result = append(result, i)
		}

		return result
	},
	"contains": func(slice []string, item string) bool {
		return slices.Contains(slice, item)
	},
	"simpleTime": func(t *time.Time) string {
		if t == nil {
			return "Not set"
		}

		return t.Format("2006-01-02 15:04")
	},
	"add": func(values ...int) int {
		sum := 0
		for _, val := range values {
			sum += val
		}

		return sum
	},
	"addFloats": func(values ...float64) float64 {
		sum := 0.0
		for _, val := range values {
			sum += val
		}

		return sum
	},
	"min": func(a, b int) int {
		if a < b {
			return a
		}

		return b
	},
	"isEven": func(i int) bool {
		return i%2 == 0
	},
	"trimQuotes": func(s string) string {
		return strings.Trim(s, `"`)
	},
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"map": func(pairs ...any) (map[string]any, error) {
		if len(pairs)%2 != 0 {
			return nil, os.ErrInvalid
		}

		m := make(map[string]any)
		for i := 0; i < len(pairs); i += 2 {
			key, ok := pairs[i].(string)
			if !ok {
				return nil, os.ErrInvalid
			}

			m[key] = pairs[i+1]
		}

		return m, nil
	},
}

func formatSecs(secs int64) string {
	formattedHours := secs / 3600
	formattedMins := (secs % 3600) / 60
	formattedSecs := secs % 60

	var b strings.Builder

	if formattedHours > 0 {
		fmt.Fprintf(&b, "%dh", formattedHours)
	}
	if formattedMins > 0 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%dm", formattedMins)
	}
	if formattedSecs > 0 || b.Len() == 0 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%ds", formattedSecs)
	}

	return b.String()
}
