package consts

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/deluan/navidrome/resources"
)

func getBanner() string {
	data, _ := resources.Asset("banner.txt")
	return strings.TrimRightFunc(string(data), unicode.IsSpace)
}

func Banner() string {
	version := "Version: " + Version()
	padding := strings.Repeat(" ", 52-len(version))
	return fmt.Sprintf("%s\n%s%s\n", getBanner(), padding, version)
}
