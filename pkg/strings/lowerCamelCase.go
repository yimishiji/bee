package strings

import "strings"

// camelCase converts a _ delimited string to lower camel case
// e.g. very_important_person => veryImportantPerson
func LowerCamelCase(in string) string {
	tokens := strings.Split(in, "_")

	index := 0
	for i := range tokens {
		if index == 0 {
			tokens[i] = strings.Trim(tokens[i], " ")
		} else {
			tokens[i] = strings.Title(strings.Trim(tokens[i], " "))
		}
		index++
	}

	return strings.Join(tokens, "")
}

func UrlStyleString(in string) string {

	tokens := strings.Split(in, "_")

	for i := range tokens {
		tokens[i] = strings.ToLower(strings.Trim(tokens[i], " "))
	}
	return strings.Join(tokens, "-")
}
