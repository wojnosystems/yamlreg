package yamlreg

import (
	"fmt"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/parser"
	"github.com/wojnosystems/go-parse-register"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
)

type ErrUnsupportedType struct {
	Line     int
	Column   int
	KeyName  string
	TypeName string
}

func (t ErrUnsupportedType) Error() string {
	if "" == t.KeyName {
		return fmt.Sprintf(`unsupported type at %d:%d: into golang type "%s"`, t.Line, t.Column, t.TypeName)
	}
	return fmt.Sprintf(`unsupported type at %d:%d: for yaml field named "%s" into golang type "%s"`, t.Line, t.Column, t.KeyName, t.TypeName)
}

type decoder struct {
	reader   io.Reader
	registry *parse_register.Registry
}

type Decoder interface {
	Decode(out interface{}) (err error)
}

func NewDecoder(r io.Reader, registry *parse_register.Registry) Decoder {
	if registry == nil {
		registry = &parse_register.Registry{}
	}
	return &decoder{
		reader:   r,
		registry: registry,
	}
}

func (d *decoder) Decode(out interface{}) (err error) {
	buffer, err := ioutil.ReadAll(d.reader)
	if err != nil {
		return
	}
	tokens := lexer.Tokenize(string(buffer))
	tree, err := parser.Parse(tokens, 0)
	if err != nil {
		return
	}
	return d.walk(tree, out)
}

func (d *decoder) walk(tree *ast.File, out interface{}) (err error) {
	err = requireValidOutputVariable(out)
	if err != nil {
		return
	}
	rv := reflect.ValueOf(out)
	elem := rv.Elem()
	for _, doc := range tree.Docs {
		err = d.walkNode(doc, elem)
		if err != nil {
			return
		}
	}
	return
}

func requireValidOutputVariable(out interface{}) (err error) {
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Ptr {
		return &ErrInvalidOutputArgument{
			Reason: "must be a reference",
		}
	}
	if v.IsNil() {
		return &ErrInvalidOutputArgument{
			Reason: "must not be nil",
		}
	}
	if !v.Elem().CanSet() {
		return &ErrInvalidOutputArgument{
			Reason: "must be settable",
		}
	}
	return
}

func (d *decoder) walkNode(node ast.Node, out reflect.Value) (err error) {
	var ok bool
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.CommentNode:
		// skip
		ok = true
	case *ast.NullNode:
		ok = true
	case *ast.IntegerNode:
		ok, err = d.registry.SetValue(out.Addr().Interface(), n.String())
	case *ast.FloatNode:
		ok, err = d.registry.SetValue(out.Addr().Interface(), n.String())
	case *ast.StringNode:
		ok, err = d.registry.SetValue(out.Addr().Interface(), n.Value)
	case *ast.MergeKeyNode:
		err = &ErrUnsupportedYAMLFeature{Feature: "MergeKeyNode"}
	case *ast.BoolNode:
		v := "false"
		if n.Value {
			v = "true"
		}
		ok, err = d.registry.SetValue(out.Addr().Interface(), v)
	case *ast.InfinityNode:
		ok, err = d.registry.SetValue(out.Addr().Interface(), "Inf")
	case *ast.NanNode:
		ok, err = d.registry.SetValue(out.Addr().Interface(), "NaN")
	case *ast.LiteralNode:
		err = d.walkNode(n.Value, out)
		ok = true
	case *ast.DirectiveNode:
		err = &ErrUnsupportedYAMLFeature{Feature: "DirectiveNode"}
	case *ast.TagNode:
		err = &ErrUnsupportedYAMLFeature{Feature: "TagNode"}
	case *ast.DocumentNode:
		err = d.walkNode(n.Body, out)
		ok = true
	case *ast.MappingNode:
		ok = true
		for _, value := range n.Values {
			err = d.walkNode(value, out)
			if err != nil {
				return
			}
		}
	case *ast.MappingKeyNode:
		err = &ErrUnsupportedYAMLFeature{Feature: "MappingKeyNode"}
	case *ast.MappingValueNode:
		outType := out.Type()
		if outType.Kind() == reflect.Struct {
			if keyNode, keyOk := n.Key.(*ast.StringNode); keyOk {
				_, fieldIndex, fieldFound := getFieldWithName(keyNode.Value, outType)
				if fieldFound {
					fieldValue := out.Field(fieldIndex)
					err = d.walkNode(n.Value, fieldValue)
					ok = true
				}
			}
		}
	case *ast.SequenceNode:
		if out.Kind() == reflect.Slice {
			ok = true
			sliceType := out.Type()
			elemType := reflect.SliceOf(sliceType).Elem()
			out.Set(reflect.MakeSlice(elemType, len(n.Values), len(n.Values)))
			for i, value := range n.Values {
				err = d.walkNode(value, out.Index(i))
				if err != nil {
					return
				}
			}
		}
	case *ast.AnchorNode:
		err = &ErrUnsupportedYAMLFeature{Feature: "AnchorNode"}
	case *ast.AliasNode:
		err = &ErrUnsupportedYAMLFeature{Feature: "AliasNode"}
	}
	if err != nil {
		return
	}
	if !ok {
		outType := out.Type()
		keyName := ""
		switch n := (node).(type) {
		case *ast.MappingValueNode:
			keyName = n.Key.String()
		}
		t := node.GetToken()
		err = &ErrUnsupportedType{
			Line:     t.Position.Line,
			Column:   t.Position.Column,
			KeyName:  keyName,
			TypeName: outType.PkgPath() + "." + outType.Name(),
		}
	}
	return
}

func getFieldWithName(nameFromYaml string, structType reflect.Type) (field reflect.StructField, index int, ok bool) {
	for index = 0; index < structType.NumField(); index++ {
		field = structType.Field(index)
		yamlTag := field.Tag.Get("yaml")
		tagParts := strings.Split(yamlTag, ",")
		if tagParts[0] == nameFromYaml || field.Name == nameFromYaml {
			ok = true
			return
		}
	}
	return
}
