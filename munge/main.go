package munge

var charMap = map[byte]rune{
	'a': '\u00e4',
	'b': '\u0411',
	'c': '\u010b',
	'd': '\u0111',
	'e': '\u00eb',
	'f': '\u0192',
	'g': '\u0121',
	'h': '\u0127',
	'i': '\u00ed',
	'j': '\u0135',
	'k': '\u0137',
	'l': '\u013a',
	'm': '\u1e41',
	'n': '\u00f1',
	'o': '\u00f6',
	'p': '\u03c1',
	'q': '\u02a0',
	'r': '\u0157',
	's': '\u0161',
	't': '\u0163',
	'u': '\u00fc',
	'v': '\u03bd',
	'w': '\u03c9',
	'x': '\u03c7',
	'y': '\u00ff',
	'z': '\u017a',
	'A': '\u00c5',
	'B': '\u0392',
	'C': '\u00c7',
	'D': '\u010e',
	'E': '\u0112',
	'F': '\u1e1e',
	'G': '\u0120',
	'H': '\u0126',
	'I': '\u00cd',
	'J': '\u0134',
	'K': '\u0136',
	'L': '\u0139',
	'M': '\u039c',
	'N': '\u039d',
	'O': '\u00d6',
	'P': '\u0420',
	'Q': '\uff31',
	'R': '\u0156',
	'S': '\u0160',
	'T': '\u0162',
	'U': '\u016e',
	'V': '\u1e7e',
	'W': '\u0174',
	'X': '\u03a7',
	'Y': '\u1ef2',
	'Z': '\u017b',
}

// Munge replaces the first character of a string
// with a similar-looking unicode character
func Munge(str string) string {
	if len(str) < 1 {
		return str
	}

	first := str[0]
	out := []rune(str)
	if munged, ok := charMap[first]; ok {
		out[0] = munged
	}

	return string(out)
}
