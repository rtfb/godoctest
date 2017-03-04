package godoctest

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
)

const (
	testFileTmpl = `package {{.PkgName}}

import (
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
)

{{range .FuncComments}}
	{{template "singleTest" .}}
{{end}}
`
	singleTestTmpl = `{{define "singleTest"}}
	func Test_{{.FuncName}}_gdt{{.Hash}}(t *testing.T) {
		{{print .PtrDataTable}}
		tests := []struct{
			{{.StructFields}}
		}{
			{{print .TestTableValues}}
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
	FuncName        string
	Hash            string
	StructFields    string
	Params          string
	PtrDataTable    string
	TestTableValues string
	ReturnValues    string
	Asserts         string
}

func prepForTemplate(id []intermediateData) []templateFuncData {
	var result []templateFuncData
	for _, d := range id {
		result = append(result, templateFuncData{
			FuncName:        d.FuncName,
			Hash:            d.Hash,
			StructFields:    makeStructFields(d.ParamTypeDefs, d.RetValTypeDefs),
			Params:          makeParams(d.ParamTypeDefs),
			PtrDataTable:    makePtrDataTable(d.PtrFields, d.PtrData),
			TestTableValues: makeTestTableValues(d.TestData),
			ReturnValues:    makeReturnValuesLHS(len(d.RetValTypeDefs)),
			Asserts:         makeAsserts(len(d.RetValTypeDefs)),
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
		if i > 0 {
			result.WriteString(", ")
		}
		fieldName := fmt.Sprintf("f%d", i)
		result.WriteString(p.arg(fieldName))
		i++
	}
	return result.String()
}

func makeReturnValuesLHS(nReturnValues int) string {
	result := bytes.Buffer{}
	for i := 0; i < nReturnValues; i++ {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString("r" + strconv.Itoa(i))
	}
	return result.String()
}

func makeAsserts(nReturnValues int) string {
	result := bytes.Buffer{}
	for i := 0; i < nReturnValues; i++ {
		if i > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(fmt.Sprintf("assert.Equal(t, test.e%d, r%d)", i, i))
	}
	return result.String()
}

func makePtrDataTable(ptrFields []string, ptrData ptrDataT) string {
	if ptrFields == nil {
		return ""
	}
	result := bytes.Buffer{}
	result.WriteString("ptrData := []struct{\n")
	for _, f := range ptrFields {
		result.WriteString("\t\t\t" + f + "\n")
	}
	result.WriteString("\t\t}{\n\t\t\t")
	for i, d := range ptrData {
		if i > 0 {
			result.WriteString("\n\t\t\t")
		}
		result.WriteString("{")
		for j, v := range d {
			if j > 0 {
				result.WriteString(", ")
			}
			result.WriteString(v.valueExpr)
		}
		result.WriteString("},")
	}
	result.WriteString("\n\t\t}")
	return result.String()
}

func makeTestTableValues(testData testDataT) string {
	result := bytes.Buffer{}
	for i, row := range testData {
		if i > 0 {
			result.WriteString("\n\t\t\t")
		}
		result.WriteByte('{')
		for j, v := range row {
			if j > 0 {
				result.WriteString(", ")
			}
			result.WriteString(v.valueExpr)
		}
		result.WriteString("},")
	}
	return result.String()
}
