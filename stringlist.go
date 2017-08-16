package karmabot

import (
	"flag"
	"strings"
)

// StringList is an object that accepts multiple strings and implements flag.Value
type StringList map[string]struct{}

var _ flag.Value = new(StringList)

func (sl *StringList) String() string {
	var (
		keys = make([]string, len(*sl))
		i    = 0
	)

	for k := range *sl {
		keys[i] = k
		i++
	}

	return strings.Join(keys, ", ")
}

// Set receives a string and appends it to the internal map
func (sl *StringList) Set(value string) error {
	(*sl)[value] = struct{}{}
	return nil
}
