package log

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"
)

var messageFileTemplate = `# Message / error codes

| Code | Explanation |
|------|-------------|
{{range $key,$val := . -}}
| ` + "`{{ $key }}`" + ` | {{ $val }} |
{{end}}
`

// GetMessageCodes parses the specified go source file and returns a key-value mapping of message codes and their
// associated description or an error.
func GetMessageCodes(filename string) (map[string]string, error) {
	result := map[string]string{}
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return result, fmt.Errorf("failed to parse file %s (%w)", filename, err)
	}

	constDocs, err := extractConstDocs(filename, fset, f)
	if err != nil {
		return result, err
	}

	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok == token.CONST {
				for _, spec := range genDecl.Specs {
					if vSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vSpec.Names {
							for _, val := range vSpec.Values {
								d, ok := constDocs[name.Name]
								if !ok || d == "" {
									return result, fmt.Errorf("constant %s is not documented", name.Name)
								}
								basicVal, ok := val.(*ast.BasicLit)
								if !ok {
									return result, fmt.Errorf(
										"the value of constant %s should be a basic string",
										name.Name,
									)
								}
								docParts := strings.Split(strings.TrimSpace(d), "\n")
								resultingDocParts := []string{}
								for _, part := range docParts {
									if !strings.HasPrefix(part, "goland:") {
										resultingDocParts = append(resultingDocParts, strings.TrimSpace(part))
									}
								}
								result[strings.Trim(basicVal.Value, `"`)] = strings.Join(resultingDocParts, " ")
							}
						}
					}
				}
			}
		}
	}

	return result, nil
}

func extractConstDocs(
	filename string,
	fset *token.FileSet,
	f *ast.File,
) (map[string]string, error) {
	p, err := doc.NewFromFiles(fset, []*ast.File{f}, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse docs from file %s (%w)", filename, err)
	}

	constDocs := map[string]string{}

	for _, c := range p.Consts {
		for _, name := range c.Names {
			constDocs[name] = c.Doc
		}
	}
	return constDocs, nil
}

// GenerateMessageCodeFiles generates the contents of the CODES.md file and returns them.
func GenerateMessageCodesFile(filename string) (string, error) {
	codes, err := GetMessageCodes(filename)
	if err != nil {
		return "", err
	}
	tpl, err := template.New("CODES.md.tpl").Parse(messageFileTemplate)
	if err != nil {
		return "", fmt.Errorf("bug: failed to parse template (%w)", err)
	}
	wr := &bytes.Buffer{}
	if err := tpl.Execute(wr, codes); err != nil {
		return "", fmt.Errorf("failed to render codes template (%w)", err)
	}
	return wr.String(), nil
}

// WriteMessageCodesFile generates and writes the CODES.md file
func WriteMessageCodesFile(sourceFile string, destinationFile string) error {
	data, err := GenerateMessageCodesFile(sourceFile)
	if err != nil {
		return err
	}
	fh, err := os.Create(destinationFile)
	if err != nil {
		return fmt.Errorf("failed to open destination file %s (%w)", destinationFile, err)
	}
	if _, err = fh.Write([]byte(data)); err != nil {
		_ = fh.Close()
		return fmt.Errorf("failed to write to destination file %s (%w)", destinationFile, err)
	}
	if err := fh.Close(); err != nil {
		return fmt.Errorf("failed to close destination file %s (%w)", destinationFile, err)
	}
	return nil
}

// MustWriteMessageCodesFile is identical to WriteMessageCodesFile but panics on error.
func MustWriteMessageCodesFile(sourceFile string, destinationFile string) {
	if err := WriteMessageCodesFile(sourceFile, destinationFile); err != nil {
		panic(err)
	}
}
