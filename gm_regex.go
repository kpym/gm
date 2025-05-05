package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

type SubstRule struct {
	Pattern *regexp.Regexp
	Replace []byte
}

type SubstRuleList []SubstRule

// sedEngines holds the sed engines for markdown and HTML.
var (
	reMdRules   SubstRuleList
	reHtmlRules SubstRuleList
)

// NewRule creates a new Rule with the given type from a string argument.
// The argument has format "<delim>pattern><delim>replace[<delim>[<comment>]]" where <delm> is a delimiter (e.g., /, |, #, @).
// The delimiter can not be escaped so it has to be not present in the pattern or replace string.
func NewRule(arg string) (SubstRule, error) {
	var rule SubstRule
	// decode first rune as delimiter
	delimiter, _ := utf8.DecodeRuneInString(arg)
	if delimiter == utf8.RuneError {
		return rule, fmt.Errorf("invalid delimiter in argument: %s", arg)
	}
	// split the argument using the delimiter
	parts := strings.SplitN(arg, string(delimiter), 4)
	if len(parts) < 3 {
		return rule, fmt.Errorf("invalid argument format: %s", arg)
	}
	// compile the pattern
	pattern, err := regexp.Compile(parts[1])
	if err != nil {
		return rule, fmt.Errorf("invalid pattern in argument: %s", arg)
	}
	// create the rule
	rule = SubstRule{
		Pattern: pattern,
		Replace: []byte(parts[2]),
	}
	return rule, nil
}

// hasRune checks if the rune is present in the two strings.
func hasRune(r rune, s1, s2 string) bool {
	for _, c := range s1 {
		if c == r {
			return true
		}
	}
	for _, c := range s2 {
		if c == r {
			return true
		}
	}
	return false
}

// String returns the string representation of the SubstRule.
// It tries to find a delimiter that is not in the pattern or replace string.
// The unlimited loop may be a problem in production, but this method is only used for debugging purposes.
func (r SubstRule) String() string {
	// delimiter to check first
	const delimiterChars = "/|#@!$%^&*()[]{}<>?;:'\"\\`~"
	// the pattern as string
	pattern := r.Pattern.String()
	// try to find a delimiter that is not in the pattern or replace string
	for _, delim := range delimiterChars {
		if !hasRune(delim, pattern, string(r.Replace)) {
			return fmt.Sprintf("%c%s%c%s%c", delim, pattern, delim, r.Replace, delim)
		}
	}
	// start enumerating delimiters from rune 0x80
	for delim := rune(0x80); true; delim++ {
		if !hasRune(delim, pattern, string(r.Replace)) {
			return fmt.Sprintf("%c%s%c%s%c", delim, pattern, delim, r.Replace, delim)
		}
	}
	// this should never happen
	return "error: no available delimiter"
}

// mergeWithFiles check for every line if it is a file name,
// if it is, it will read the file and insert the content
// in place of the line in the slice of lines.
// It will return a slice of lines.
func mergeWithFiles(s string) []string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // skip empty lines
		}
		if _, err := os.Stat(line); err == nil {
			// If the line is a file, read it and append its content.
			fileContent, err := os.ReadFile(line)
			if err != nil {
				continue // skip invalid files
			}
			// Append the file linses to the result.
			for _, fileLine := range strings.Split(string(fileContent), "\n") {
				fileLine = strings.TrimSpace(fileLine)
				if fileLine != "" {
					result = append(result, fileLine) // Append non-empty lines from the file
				}
			}
		} else {
			// Otherwise, just append the line as is.
			result = append(result, line)
		}
	}
	return result
}

// DecodeRules convert all lines in the string to SubstRule objects.
// Even if it find errors, it will return the rules found so far.
func DecodeRules(rules string) ([]SubstRule, error) {
	var e error
	rulesList := mergeWithFiles(rules)
	var result []SubstRule
	for _, rule := range rulesList {
		if strings.TrimSpace(rule) == "" {
			continue // skip empty lines
		}
		substRule, err := NewRule(rule)
		if err != nil {
			if e == nil {
				e = fmt.Errorf("%s has error '%s'", rule, err.Error())
			}
			continue // skip invalid rules
		}
		result = append(result, substRule)
	}
	return result, e
}

// Apply applies the substitution rule to the input byte slice and returns the result.
func (rules SubstRuleList) Apply(input []byte) []byte {
	for _, rule := range rules {
		input = rule.Pattern.ReplaceAll(input, rule.Replace)
	}
	return input
}
