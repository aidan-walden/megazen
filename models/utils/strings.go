package utils

import "regexp"

var re = regexp.MustCompile("[|&;$%@\"<>()+,?]")

func ValidTitleString(path string) string {
	return re.ReplaceAllString(path, "-")
}
