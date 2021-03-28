package yamlreg

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wojnosystems/go-optional"
	"github.com/wojnosystems/go-optional-parse-registry"
	"github.com/wojnosystems/go-parse-register"
	"math"
	"testing"
)

type decoderNestedStruct struct {
	Path string
}

type decoderTestStruct struct {
	Name        string `yaml:"name"`
	Age         int
	FluffFactor float32
	FavNums     []int
	DoTrue      bool
	Config      decoderNestedStruct
	Nullable    *string
}

type decoderTestNanInf struct {
	Nan float64
	Inf float64
}

type decoderTestOpt struct {
	Name optional.String
}

func TestDecoder_Decode(t *testing.T) {
	goRegistry := parse_register.GoPrimitives()
	t.Run("go primitives", func(t *testing.T) {
		cases := map[string]struct {
			reg         parse_register.ValueSetter
			input       string
			expected    decoderTestStruct
			expectedErr string
		}{
			"ok": {
				reg: goRegistry,
				input: `---
name: "puppy"
Age: 30
FluffFactor: 3.14159
# Comment
FavNums:
 - 2
 - 13
 - 27
DoTrue: true
Config:
 Path: "/thing"
Nullable: null
`,
				expected: decoderTestStruct{
					Name:        "puppy",
					Age:         30,
					FluffFactor: 3.14159,
					FavNums:     []int{2, 13, 27},
					DoTrue:      true,
					Config:      decoderNestedStruct{Path: "/thing"},
				},
			},
			"extra field": {
				reg: goRegistry,
				input: `---
missing: "puppy"
`,
				expected:    decoderTestStruct{},
				expectedErr: "unsupported type at 2:8: for yaml field named \"missing\" into golang type \"github.com/wojnosystems/yamlreg.decoderTestStruct\"",
			},
			"literal": {
				reg: goRegistry,
				input: `---
Name: >
  Literal Value
`,
				expected: decoderTestStruct{
					Name: "Literal Value",
				},
			},
			"comment": {
				reg: goRegistry,
				input: `---
Name: a name # comment
`,
				expected: decoderTestStruct{
					Name: "a name",
				},
			},
			"unsupported types": {
				reg: parse_register.New(),
				input: `---
Name: name
`,
				expected:    decoderTestStruct{},
				expectedErr: "unsupported type at 2:7: into golang type \".string\"",
			},
		}
		for caseName, c := range cases {
			t.Run(caseName, func(t *testing.T) {
				buffer := bytes.NewReader([]byte(c.input))
				dec := NewDecoder(buffer, c.reg)
				var actual decoderTestStruct
				err := dec.Decode(&actual)
				if c.expectedErr != "" {
					assert.EqualError(t, err, c.expectedErr)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, c.expected, actual)
				}
			})
		}
	})

	t.Run("float NaN/Inf", func(t *testing.T) {
		buffer := bytes.NewReader([]byte(`---
Nan: .nan
Inf: .inf
`))
		dec := NewDecoder(buffer, goRegistry)
		var actual decoderTestNanInf
		err := dec.Decode(&actual)
		assert.NoError(t, err)
		assert.True(t, math.IsNaN(actual.Nan))
		assert.True(t, math.IsInf(actual.Inf, 1))
	})

	t.Run("optional", func(t *testing.T) {

		optReg := optional_parse_registry.RegisterFluent(goRegistry)
		cases := map[string]struct {
			reg         parse_register.ValueSetter
			input       string
			expected    decoderTestOpt
			expectedErr string
		}{
			"ok": {
				reg: optReg,
				input: `---
Name: "puppy"
`,
				expected: decoderTestOpt{
					Name: optional.StringFrom("puppy"),
				},
			},
			"not set": {
				reg: optReg,
				input: `---
`,
				expected: decoderTestOpt{},
			},
		}
		for caseName, c := range cases {
			t.Run(caseName, func(t *testing.T) {
				buffer := bytes.NewReader([]byte(c.input))
				dec := NewDecoder(buffer, c.reg)
				var actual decoderTestOpt
				err := dec.Decode(&actual)
				if c.expectedErr != "" {
					assert.EqualError(t, err, c.expectedErr)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, c.expected, actual)
				}
			})
		}
	})

	t.Run("custom map type", func(t *testing.T) {
		buffer := bytes.NewReader([]byte(`---
custom:
  key1: 1
  key2: 2
`))
		dec := NewDecoder(buffer, goRegistry)
		var actual customMapTypeWrapper
		err := dec.Decode(&actual)
		require.NoError(t, err)
		assert.Equal(t, len(actual.Custom), 2)
		assert.Equal(t, 1, actual.Custom["key1"])
		assert.Equal(t, 2, actual.Custom["key2"])
	})

	t.Run("custom map type nested custom", func(t *testing.T) {
		buffer := bytes.NewReader([]byte(`---
custom:
  key1:
    sub1: 1
  key2:
    sub2: 2
    sub23: 23
`))
		dec := NewDecoder(buffer, goRegistry)
		var actual customMapTypeNestedWrapper
		err := dec.Decode(&actual)
		require.NoError(t, err)
		assert.Equal(t, 1, actual.Custom["key1"]["sub1"])
		assert.Equal(t, 2, actual.Custom["key2"]["sub2"])
		assert.Equal(t, 23, actual.Custom["key2"]["sub23"])
	})
}

type customMapType map[string]int

type customMapTypeWrapper struct {
	Custom customMapType `yaml:"custom"`
}

type customMapNestedType map[string]customMapType

type customMapTypeNestedWrapper struct {
	Custom customMapNestedType `yaml:"custom"`
}
