package socks

import "regexp"

func isIPv4(input string) bool {
	regex := regexp.MustCompile(`^(?:(?:[0-9]{1,3}\.){3}[0-9]{1,3})$`)
	return regex.MatchString(input)
}

func isIPv6(input string) bool {
	regex := regexp.MustCompile(`^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
	return regex.MatchString(input)
}

func isDomain(input string) bool {
	regex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	return regex.MatchString(input)
}