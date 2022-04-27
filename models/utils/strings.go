package utils

import "regexp"

var re = regexp.MustCompile("[|&;$%@\"<>()+,?/]")

func ValidPathString(path string) string {
	return re.ReplaceAllString(path, "-")
}
