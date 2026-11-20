package message

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// DocStruct documents a structure.
type DocStruct struct {
	Name        string
	Description string
	Fields      []DocField
}

// DocField is a record if documentation inside a struct.
type DocField struct {
	Name     string
	DataType string
	Comment  string
	Embedded *DocStruct
}

// DocType contains an extra description for types.
type DocType struct {
	Type

	Description string
}

// Documentation is the entire documentation for the messages.
type Documentation struct {
	// Message is the main message object
	Message DocStruct
	// Payloads is a map for payload codes to struct docs
	Payloads map[DocType]*DocStruct
}

// DocumentMessages returns a documentation for the message format.
func DocumentMessages() Documentation {
	typeDocs := getTypeDocs()

	msg := document(&Message{}, typeDocs)
	payloads := map[DocType]*DocStruct{}
	for _, messageType := range ListTypes() {
		payload, err := messageType.Payload()
		if err != nil {
			panic(err)
		}
		var payloadDoc *DocStruct = nil
		if payload != nil {
			d := document(payload, typeDocs)
			payloadDoc = &d
		}
		typeVal := reflect.ValueOf(messageType)
		payloads[DocType{
			Type:        messageType,
			Description: typeDocs[typeVal.Kind().String()],
		}] = payloadDoc
	}
	return Documentation{
		Message:  msg,
		Payloads: payloads,
	}
}

func getTypeDocs() map[string]string {
	// Parse this package's own source, so the result does not depend on where the generator runs.
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot find the source of the message package")
	}
	names, err := filepath.Glob(filepath.Join(filepath.Dir(self), "*.go"))
	if err != nil {
		panic(err)
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, name := range names {
		f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}
		files = append(files, f)
	}
	p, err := doc.NewFromFiles(fset, files, "go.containerssh.io/containerssh/auditlog/message")
	if err != nil {
		panic(err)
	}
	space := regexp.MustCompile(`\s+`)
	typeDocs := map[string]string{}
	for _, t := range p.Types {
		typeDocs[t.Name] = space.ReplaceAllString(strings.ReplaceAll(t.Doc, "\n", " "), " ")
		previousName := ""
		ast.Inspect(
			t.Decl, func(node ast.Node) bool {
				switch nodeType := node.(type) {
				case *ast.TypeSpec:
					previousName = nodeType.Name.Name
				case *ast.StructType:
					for _, field := range nodeType.Fields.List {
						comment := space.ReplaceAllString(strings.ReplaceAll(field.Comment.Text(), "\n", " "), " ")
						for _, name := range field.Names {
							if comment != "" {
								typeDocs[previousName+"."+name.String()] = comment
							}
						}
					}
				}
				return true
			},
		)
	}
	return typeDocs
}

func document(entity interface{}, docs map[string]string) DocStruct {
	t := reflect.TypeOf(entity)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Interface {
		t = reflect.ValueOf(t).Elem().Type()
	}
	if t.Kind() != reflect.Struct {
		panic(t.Name() + " is not a struct")
	}
	return documentType(t, docs)
}

func resolveTypeName(fieldType reflect.Type) string {
	switch fieldType.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int:
		return "int"
	case reflect.Int8:
		return "int8"
	case reflect.Int16:
		return "int16"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Uint:
		return "uint"
	case reflect.Uint8:
		return "uint8"
	case reflect.Uint16:
		return "uint16"
	case reflect.Uint32:
		return "uint32"
	case reflect.Uint64:
		return "uint64"
	case reflect.Bool:
		return "bool"
	case reflect.Slice:
		return "[]" + resolveTypeName(fieldType.Elem())
	case reflect.Ptr:
		return "*" + resolveTypeName(fieldType.Elem())
	case reflect.Interface:
		return resolveTypeName(reflect.ValueOf(fieldType).Elem().Type())
	case reflect.Struct:
		return fieldType.Kind().String()
	case reflect.Uintptr:
		return "*uint"
	default:
		panic("unknown type: " + fieldType.Name() + " (" + fieldType.Kind().String() + ")")
	}
}

func documentType(t reflect.Type, docs map[string]string) DocStruct {
	var fields []DocField
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldType := f.Type
		if fieldType.Kind() == reflect.Func {
			continue
		}

		// The binary format uses the JSON tag names as CBOR map keys.
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" {
			name = f.Name
		}
		record := DocField{
			Name:     name,
			DataType: resolveTypeName(fieldType),
			Comment:  docs[t.Name()+"."+f.Name],
		}
		if f.Type.Kind() == reflect.Struct {
			embedded := documentType(f.Type, docs)
			record.Embedded = &embedded
		}
		fields = append(
			fields, record,
		)
	}
	return DocStruct{
		Name:        t.Name(),
		Description: docs[t.Name()],
		Fields:      fields,
	}
}
