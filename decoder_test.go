package yamlreg

import (
	"bytes"
	"github.com/stretchr/testify/assert"
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
	goRegistry := parse_register.RegisterGoPrimitives(&parse_register.Registry{})
	t.Run("go primitives", func(t *testing.T) {
		cases := map[string]struct {
			reg         *parse_register.Registry
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
				expectedErr: "unsupported type at 2:8: for yaml field named \"missing\" into golang type \"yamlreg.decoderTestStruct\"",
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

		optReg := optional_parse_registry.Register(goRegistry)
		cases := map[string]struct {
			reg         *parse_register.Registry
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
}
