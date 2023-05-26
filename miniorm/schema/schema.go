package schema

import (
	"go/ast"
	"reflect"

	"github.com/qingants/pandora/miniorm/dialect"
)

// type User struct {
// 	Name string `miniorm:"PRIMARY KEY"`
// 	Age  int
// }

type Field struct {
	Name string
	Type string
	Tag  string
}

type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

func (s *Schema) GetField(name string) *Field {
	return s.fieldMap[name]
}

func Parse(dest any, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("miniorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	// var s *Schema
	// s.fieldMap = make(map[string]*Field)
	// s.fieldNames = make([]string, 0, len(d.Fields))
	// for _, f := range d.Fields {
	// 	s.fieldMap[f.Name] = &Field{
	// 		Name: f.Name,
	// 		Type: f.Type,
	// 		Tag:  f.Tag,
	// 	}
	// }

	return schema
}
