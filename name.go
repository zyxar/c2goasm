package main

import (
	"strings"

	"github.com/ianlancetaylor/demangle"
)

func parseProtoName(name string) string {
	if name == "" || name[0] != '_' {
		return name
	}
	if strings.HasPrefix(name, "__Z") { // guess cc source compiled on darwin
		name = name[1:]
	}
	if strings.HasPrefix(name, "_Z") { // cc mangled name
		n, err := demangle.ToString(name, demangle.NoParams, demangle.NoTemplateParams)
		if err != nil {
			return name
		}
		return strings.Replace(n, "::", "", 100)
	}
	return name[1:]
}
