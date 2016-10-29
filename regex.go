package karmabot

import (
	"fmt"
	"regexp"
	"strings"
)

type karmaRegex struct {
	user, autocomplete, explicitAutocomplete, points, reason string
}

var karmaReg = &karmaRegex{
	user:                 `@??((?:<@)??\w[A-Za-z0-9_\-@<>]+?)`,
	autocomplete:         `:?? ??`,
	explicitAutocomplete: `(?:: )??`,
	points:               `([\+]{2,}|[\-]{2,})`,
	reason:               `(?:(?: for)? +(.*))?`,
}

func (r *karmaRegex) GetGive() *regexp.Regexp {
	expression := fmt.Sprintf(
		"(?:%s)|(?:%s)",
		strings.Join(
			[]string{
				"^",
				r.user,
				r.autocomplete,
				r.points,
				r.reason,
				"$",
			},
			"",
		),
		strings.Join(
			[]string{
				`\s+`,
				r.user,
				r.explicitAutocomplete,
				r.points,
				r.reason,
				"$",
			},
			"",
		),
	)

	return regexp.MustCompile(expression)
}

func (r *karmaRegex) GetMotivate() *regexp.Regexp {
	expression := strings.Join(
		[]string{
			`^(?:\?|!)m +`,
			r.user,
			r.autocomplete,
			"$",
		},
		"",
	)

	return regexp.MustCompile(expression)
}
