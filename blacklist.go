package karmabot

import (
	"flag"
	"strings"
)

type UserBlacklist map[string]struct{}

var _ flag.Value = new(UserBlacklist)

func (ub *UserBlacklist) String() string {
	var (
		keys = make([]string, len(*ub))
		i    = 0
	)

	for k := range *ub {
		keys[i] = k
		i++
	}

	return strings.Join(keys, ", ")
}

func (ub *UserBlacklist) Set(value string) error {
	(*ub)[value] = struct{}{}
	return nil
}
