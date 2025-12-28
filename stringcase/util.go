package stringcase

import (
	"slices"
	"strings"

	"github.com/leshless/golibrary/xslices"
)

type word struct {
	runes       []rune
	isShorthand bool
}

// for consistency
func (w word) toLower() []rune {
	return w.runes
}

func (w word) toUpper() []rune {
	if w.isShorthand {
		return xslices.Map(w.runes, toUpper)
	}

	runes := make([]rune, len(w.runes))
	copy(runes, w.runes)
	runes[0] = toUpper(runes[0])

	return runes
}

func isLower(char rune) bool {
	return (latinLowerMin <= char) && (char <= latinLowerMax)
}

func isUpper(char rune) bool {
	return (latinUpperMin <= char) && (char <= latinUpperMax)
}

func toUpper(char rune) rune {
	if !isLower(char) {
		return char
	}

	return char - latinLowerMin + latinUpperMin
}

func toLower(char rune) rune {
	if !isUpper(char) {
		return char
	}

	return char - latinUpperMin + latinLowerMin
}

// split tries to "parce" input phrase by words also trying to detect shorthand words and mark them
// Notice that this will produce expected result for phrases in all cases which are considered by this package
// however, for some mixed-case phrases something pretty messy may be returned
func split(runes []rune) []word {
	words := make([]word, 0)
	var (
		current     strings.Builder
		isShorthand bool
	)

	startNextWord := func() {
		runes := []rune(current.String())

		if len(runes) != 0 {
			words = append(words, word{
				runes:       runes,
				isShorthand: isShorthand,
			})
		}

		current.Reset()
		isShorthand = false
	}

	for i := range len(runes) {
		if slices.Contains(separators, runes[i]) {
			startNextWord()
			continue
		}

		if !isUpper(runes[i]) {
			current.WriteRune(runes[i])
			continue
		}

		if (i != 0 && isUpper(runes[i-1])) || i == len(runes)-1 || isUpper(runes[i+1]) {
			isShorthand = true
		}

		if i != 0 && !isUpper(runes[i-1]) {
			startNextWord()
		}
		if i != len(runes)-1 && isLower(runes[i+1]) {
			startNextWord()
		}

		current.WriteRune(toLower(runes[i]))
	}

	startNextWord()

	return words
}

func casify(text string, separator string, isUpper bool) string {
	runes := []rune(text)
	words := split(runes)

	return strings.Join(xslices.Map(words, func(w word) string {
		if isUpper {
			return string(w.toUpper())
		}

		return string(w.toLower())
	}), separator)
}
