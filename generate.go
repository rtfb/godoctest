package main

import (
	"bytes"
	"html/template"
)

const (
	testFileTmpl = `package {{.PkgName}}

import (
	"testing"
)

{{range .FuncComments}}
	{{template "singleTest" .}}
{{end}}
`
	singleTestTmpl = `{{define "singleTest"}}
	func Test_{{.FuncName}}_gdc{{.Hash}}(t *testing.T) {
		tests := []struct{
			{StructFields}
		}{
			{TestValues}
		}
		for _, test := range tests {
			{ReturnValues} := {{.FuncName}}({Params})
			{Asserts}
		}
	}
{{end}}
`
)

func genPkgTests(fc *fileComments) string {
	t := template.Must(template.New("File tests template").Parse(testFileTmpl))
	t = template.Must(t.Parse(singleTestTmpl))
	var buf bytes.Buffer
	err := t.Execute(&buf, map[string]interface{}{
		"PkgName":      fc.pkg.Name,
		"FuncComments": prepForTemplate(fc.funcComments),
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
