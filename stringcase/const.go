package stringcase

type caseStyle uint8

const (
	caseStyleLower caseStyle = iota
	caseStyleUpper
	caseStyleTitle
)

const (
	latinLowerMin rune = 'a'
	latinLowerMax rune = 'z'
	latinUpperMin rune = 'A'
	latinUpperMax rune = 'Z'

	snakeSeparator rune = '_'
	kebabSeparator rune = '-'
)

var separators = []rune{' ', '\t', '\n', snakeSeparator, kebabSeparator}
