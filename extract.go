package godoctest

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type Extractor struct {
	fset *token.FileSet
}

func NewExtractor() *Extractor {
	return &Extractor{
		fset: token.NewFileSet(),
	}
}

func (e *Extractor) Run(dir string) []*fileComments {
	// include tells parser.ParseDir which files to include.
	include := func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go")
	}
	pkgs, err := parser.ParseDir(e.fset, "testdata", include, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	fcs := extractComments(pkgs, e.fset)
	extractFuncs(fcs, e.fset)
	return fcs
}

type fileComments struct {
	pkg          *ast.Package
	fileName     string
	file         *ast.File
	comments     map[int]*ast.CommentGroup
	funcComments map[string]funcData
}

func (fc *fileComments) TestFileName() string {
	fn := fc.fileName
	ext := path.Ext(fn)
	return fn[:len(fn)-len(ext)] + "_gdt_test.go"
}

type funcData struct {
	declLine    int
	commentLine int
	decl        *ast.FuncType
	comment     *ast.CommentGroup
}

func containsTestAnnotation(cg *ast.CommentGroup) bool {
	for _, c := range cg.List {
		if strings.Contains(c.Text, "@test") {
			return true
		}
	}
	return false
}

func extractComments(pkgs map[string]*ast.Package, fs *token.FileSet) []*fileComments {
	var fcs []*fileComments
	for _, v := range pkgs {
		for fk, fv := range v.Files {
			var fc fileComments
			fc.comments = map[int]*ast.CommentGroup{}
			fc.pkg = v
			fc.fileName = fk
			fc.file = fv
			for _, c := range fv.Comments {
				if !containsTestAnnotation(c) {
					continue
				}
				fc.comments[fs.Position(c.Pos()).Line] = c
			}
			fcs = append(fcs, &fc)
		}
	}
	return fcs
}

func extractFuncs(fcs []*fileComments, fs *token.FileSet) {
	for _, fc := range fcs {
		fc.funcComments = map[string]funcData{}
		for _, d := range fc.file.Decls {
			switch typedDecl := d.(type) {
			case *ast.FuncDecl:
				declLine := fs.Position(d.Pos()).Line
				endLine := fs.Position(typedDecl.Type.End()).Line
				_, found := fc.comments[endLine+1]
				if !found {
					continue
				}
				fc.funcComments[typedDecl.Name.Name] = funcData{
					declLine:    declLine,
					commentLine: endLine + 1,
					decl:        typedDecl.Type,
					comment:     fc.comments[endLine+1],
				}
			}
		}
		fc.comments = nil // release for GC
	}
}

func cgToStr(cg *ast.CommentGroup) string {
	var s bytes.Buffer
	for i, c := range cg.List {
		if i > 0 {
			s.WriteByte('\n')
		}
		s.WriteString(c.Text)
	}
	return s.String()
}

func extractTestValues(cg *ast.CommentGroup) string {
	blockComment := strings.HasPrefix(cg.List[0].Text, "/*")
	var s bytes.Buffer
	if blockComment {
		s1 := strings.TrimPrefix(cg.List[0].Text, "/*")
		s2 := strings.TrimSuffix(s1, "*/")
		s3 := strings.TrimPrefix(strings.Trim(s2, " \t\n"), "@test = {")
		s.WriteString(strings.TrimSuffix(strings.Trim(s3, " \t\n"), "}"))
	} else {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "@test =") {
				continue
			}
			if strings.Trim(c.Text[2:], " ") == "}" {
				continue
			}
			s.WriteString(c.Text[2:])
		}
	}
	return s.String()
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

func prepForTemplate(fcm map[string]funcData) []templateFuncData {
	var result []templateFuncData
	i := 0
	for k, v := range fcm {
		i++
		structFields, params := makeStructFieldsAndParams(v.decl.Params)
		resultFields, retVals, asserts := makeResults(v.decl.Results)
		result = append(result, templateFuncData{
			FuncName:     k,
			Hash:         strconv.Itoa(i), // TODO: come up with smth better
			StructFields: structFields + "\n" + resultFields,
			Params:       params,
			ReturnValues: retVals,
			Asserts:      asserts,
			TestValues:   extractTestValues(v.comment),
		})
	}
	return result
}

func makeStructFieldsAndParams(params *ast.FieldList) (string, string) {
	var fieldLines []string
	var args []string
	for i, f := range params.List {
		fieldName := fmt.Sprintf("f%d", i)
		fieldLines = append(fieldLines, fmt.Sprintf("%s %s", fieldName,
			makeTypeStr(f.Type)))
		args = append(args, "test."+fieldName)
	}
	return strings.Join(fieldLines, "\n"), strings.Join(args, ", ")
}

// TODO: it can be pointers, variadic params and such. Lots of work remaining
// here. But maybe I'm overcomplicating? Maybe just copy source code bytes from
// the range [typeExpr.Pos():typeExpr.End()]?
func makeTypeStr(typeExpr ast.Expr) string {
	switch typedExpr := typeExpr.(type) {
	case *ast.Ident:
		return typedExpr.Name
	default:
		panic("TODO")
	}
}

// TODO: results may be nil. Make sure this is handled
func makeResults(results *ast.FieldList) (string, string, string) {
	var fieldLines []string
	var resultList []string
	var asserts []string
	for i, f := range results.List {
		fieldLines = append(fieldLines, fmt.Sprintf("e%d %s", i,
			makeTypeStr(f.Type)))
		resultList = append(resultList, fmt.Sprintf("r%d", i))
		asserts = append(asserts, fmt.Sprintf("assert.Equal(t, test.e%d, r%d)", i, i))
	}
	j1 := strings.Join(fieldLines, "\n")
	j2 := strings.Join(resultList, ", ")
	j3 := strings.Join(asserts, "\n")
	return j1, j2, j3
}
