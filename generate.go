package godoctest

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	testFileTmpl = `package {{.PkgName}}

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

{{range .FuncComments}}
	{{template "singleTest" .}}
{{end}}
`
	singleTestTmpl = `{{define "singleTest"}}
	func Test_{{.FuncName}}_gdt{{.Hash}}(t *testing.T) {
		tests := []struct{
			{{.StructFields}}
		}{
			{{print .TestValues}}
		}
		for _, test := range tests {
			{{.ReturnValues}} := {{.FuncName}}({{.Params}})
			{{.Asserts}}
		}
	}
{{end}}
`
)

func GenPkgTests(fc *fileComments) string {
	t := template.Must(template.New("File tests template").Parse(testFileTmpl))
	t = template.Must(t.Parse(singleTestTmpl))
	var buf bytes.Buffer
	err := t.Execute(&buf, map[string]interface{}{
		"PkgName":      fc.pkg.Name,
		"FuncComments": prepForTemplate(extract(fc.funcComments)),
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}

type templateFuncData struct {
	FuncName     string
	Hash         string
	StructFields string
	Params       string
	TestValues   string
	ReturnValues string
	Asserts      string
}

func prepForTemplate(id []intermediateData) []templateFuncData {
	var result []templateFuncData
	for _, d := range id {
		result = append(result, templateFuncData{
			FuncName:     d.FuncName,
			Hash:         d.Hash,
			StructFields: makeStructFields(d.ParamTypeDefs, d.RetValTypeDefs),
			Params:       makeParams(d.ParamTypeDefs),

			TestValues:   d.TestValues,
			ReturnValues: d.ReturnValues,
			Asserts:      d.Asserts,
		})
	}
	return result
}

func makeStructFields(paramTypeDefs, retValTypeDefs []*typeDef) string {
	i := 0
	result := bytes.Buffer{}
	for _, p := range paramTypeDefs {
		fieldName := fmt.Sprintf("f%d", i)
		result.WriteString(p.field(fieldName) + "\n")
		i++
	}
	i = 0
	for _, r := range retValTypeDefs {
		fieldName := fmt.Sprintf("e%d", i)
		result.WriteString(r.field(fieldName) + "\n")
		i++
	}
	return result.String()
}

func makeParams(paramTypeDefs []*typeDef) string {
	i := 0
	result := bytes.Buffer{}
	for _, p := range paramTypeDefs {
		fieldName := fmt.Sprintf("f%d", i)
		result.WriteString(p.arg(fieldName) + ", ")
		i++
	}
	return result.String()
}
