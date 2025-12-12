package stringcase

const (
	latinLowerMin rune = 'a'
	latinLowerMax rune = 'z'
	latinUpperMin rune = 'A'
	latinUpperMax rune = 'Z'

	snakeSeparator rune = '_'
	kebabSeparator rune = '-'
)

var separators = []rune{' ', '\t', '\n', snakeSeparator, kebabSeparator}
