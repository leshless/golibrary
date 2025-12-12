package stringcase

import "strings"

// unfortunately this is the only thing we couldn't handle by magic func
func LowerCamel(text string) string {
	runes := []rune(text)
	words := split(runes)

	var result strings.Builder

	result.WriteString(string(words[0].toLower()))
	for i := 1; i < len(words); i++ {
		result.WriteString(string(words[i].toUpper()))
	}

	return result.String()
}

func UpperCamel(text string) string {
	return casify(text, "", true)
}

func LowerSnake(text string) string {
	return casify(text, string(snakeSeparator), false)
}

func UpperSnake(text string) string {
	return casify(text, string(snakeSeparator), true)
}

func LowerKebab(text string) string {
	return casify(text, string(kebabSeparator), false)
}

func UpperKebab(text string) string {
	return casify(text, string(kebabSeparator), true)
}
