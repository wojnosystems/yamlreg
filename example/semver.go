package main

import (
	"bytes"
	"fmt"
	"github.com/wojnosystems/go-parse-register"
	"github.com/wojnosystems/yamlreg"
	"reflect"
	"strings"
)

type SemVer struct {
	Major string
	Minor string
	Patch string
}

func main() {
	var myFileVersion SemVer
	registry := parse_register.RegisterGoPrimitives(&parse_register.Registry{})
	registry.Register(reflect.TypeOf((*SemVer)(nil)).Elem(), func(settableDst interface{}, value string) (err error) {
		version := settableDst.(*SemVer)
		parts := strings.Split(value, ".")
		version.Major = parts[0]
		version.Minor = parts[1]
		version.Patch = parts[2]
		return
	})
	dec := yamlreg.NewDecoder(bytes.NewReader([]byte("1.2.3")), registry)
	_ = dec.Decode(&myFileVersion)
	fmt.Printf("%v\n", myFileVersion)
}
