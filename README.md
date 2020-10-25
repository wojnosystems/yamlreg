# Overview

Load a YAML file into a structure with custom types. Register your handlers. Pass them to the decoder, load your data into your structures.

# How to use it

```go
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
	registry := parse_register.GoPrimitives()
	registry.Register(reflect.TypeOf((*SemVer)(nil)).Elem(), func(settableDst interface{}, value string) (err error){
		version := settableDst.(*SemVer)
		parts := strings.Split(value,".")
		version.Major = parts[0]
		version.Minor = parts[1]
		version.Patch = parts[2]
		return
	})
	dec := yamlreg.NewDecoder(bytes.NewReader([]byte("1.2.3")), registry)
	_ = dec.Decode(&myFileVersion)
	fmt.Printf("%v\n", myFileVersion)
}
```

Produces:

```
{1 2 3}
```

The Major, Minor, and Patch fields of myFileVersion struct are set to 1, 2, and 3, respectively.

# Why did you write this?

Honestly, I think the way the current YAML and even the JSON libraries handle custom struct deserialization is broken. Say you want to support a custom deserialization, as exemplified above for SemVer, with the current yaml.v2 or yaml.v3, how do you do that?

According to the documentation, you need to add a method called UnmarshalYAML for your custom structure. However, what if you don't control the code for the object you're deserializing? You cannot add a method to an object in a different package. This leads to solutions like wrapping objects in custom compositions to implement the new methods, but then you need to unwrap the outer class to use the actual value. And, this also means that your code is tightly coupled to the deserialization strategy. Your code now depends on and must import the yaml library. Maybe somebody wants to use your code but with JSON? Now they depend on YAML and will never use it.

This method allows you to register any custom deserialization logic to convert a parsed string key into a value in any type of your choice. The parse registry comes with default go handlers if you want, but you can add any handlers for any reflect types you desire, including overwriting the Go primitives for custom handling.

# Future work

* I didn't have time to support all of the YAML features. I did not do any work with aliases, tags, isolated literals, or mergekeys. I only did enough to support the major literals, mapping, and sequences.
* More tests with more coverage
