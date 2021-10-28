package main

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

type tplData struct {
	Title string
	Codes map[string]string
}

var messageFileTemplate = `# {{ .Title }} 

| Code | Explanation |
|------|-------------|
{{range $key,$val := .Codes -}}
| ` + "`{{ $key }}`" + ` | {{ $val }} |
{{end}}
`

// getMessageCodes parses the specified go source file and returns a key-value mapping of message codes and their
// associated description or an error.
func getMessageCodes(filename string) (map[string]string, error) {
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
								resultingDoc := strings.Join(resultingDocParts, " ")
								resultingDoc = strings.TrimPrefix(
									resultingDoc,
									fmt.Sprintf("%s indicates that ", name.Name),
								)
								resultingDoc = strings.ToUpper(resultingDoc[:1]) + resultingDoc[1:]
								result[strings.Trim(basicVal.Value, `"`)] = resultingDoc
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

// generateMessageCodesFile generates the contents of the CODES.md file and returns them.
func generateMessageCodesFile(filename string, title string) (string, error) {
	codes, err := getMessageCodes(filename)
	if err != nil {
		return "", err
	}
	tpl, err := template.New("CODES.md.tpl").Parse(messageFileTemplate)
	if err != nil {
		return "", fmt.Errorf("bug: failed to parse template (%w)", err)
	}
	wr := &bytes.Buffer{}
	if err := tpl.Execute(wr, &tplData{
		Title: title,
		Codes: codes,
	}); err != nil {
		return "", fmt.Errorf("failed to render codes template (%w)", err)
	}
	return wr.String(), nil
}

// writeMessageCodesFile generates and writes the CODES.md file
func writeMessageCodesFile(sourceFile string, destinationFile string, title string) error {
	data, err := generateMessageCodesFile(sourceFile, title)
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

// mustWriteMessageCodesFile is identical to writeMessageCodesFile but panics on error.
func mustWriteMessageCodesFile(sourceFile string, destinationFile string, title string) {
	if err := writeMessageCodesFile(sourceFile, destinationFile, title); err != nil {
		panic(err)
	}
}

func main() {
	sourceFile := os.Args[1]
	destinationFile := os.Args[2]
	title := os.Args[3]

	mustWriteMessageCodesFile(sourceFile, destinationFile, title)
	fmt.Printf("Generated %s.", destinationFile)
}
