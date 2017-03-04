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
	pkgs, err := parser.ParseDir(e.fset, dir, include, parser.ParseComments)
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

type intermediateData struct {
	FuncName       string
	Hash           string
	ParamTypeDefs  []*typeDef
	RetValTypeDefs []*typeDef
	TestValues     string
}

func extract(fcm map[string]funcData) []intermediateData {
	var result []intermediateData
	i := 0
	for k, v := range fcm {
		i++
		typeDefs := extractTypeDefs(v.decl.Params)
		retValDefs := extractRetValDefs(v.decl.Results)
		result = append(result, intermediateData{
			FuncName:       k,
			Hash:           strconv.Itoa(i), // TODO: come up with smth better
			ParamTypeDefs:  typeDefs,
			RetValTypeDefs: retValDefs,
			TestValues:     extractTestValues(v.comment),
		})
	}
	return result
}

func extractTypeDefs(params *ast.FieldList) []*typeDef {
	var result []*typeDef
	for _, f := range params.List {
		result = append(result, makeTypeDef(f.Type))
	}
	return result
}

func extractRetValDefs(results *ast.FieldList) []*typeDef {
	var retValDefs []*typeDef
	for _, f := range results.List {
		retValDefs = append(retValDefs, makeTypeDef(f.Type))
	}
	return retValDefs
}

type typeDef struct {
	typeName   string
	isPtr      bool
	isNil      bool
	isEllipsis bool
}

func (d *typeDef) field(fieldName string) string {
	if d.isEllipsis {
		return fieldName + " []" + d.typeName
	}
	if d.isPtr {
		return fieldName + " *" + d.typeName
	}
	return fieldName + " " + d.typeName
}

func (d *typeDef) arg(fieldName string) string {
	arg := "test." + fieldName
	if d.isEllipsis {
		arg += "..."
	}
	return arg
}

func makeTypeDef(typeExpr ast.Expr) *typeDef {
	switch typedExpr := typeExpr.(type) {
	case *ast.Ident:
		return &typeDef{
			typeName: typedExpr.Name,
		}
	case *ast.StarExpr:
		switch typedExpr2 := typedExpr.X.(type) {
		case *ast.Ident:
			return &typeDef{
				typeName: typedExpr2.Name,
				isPtr:    true,
			}
		default:
			panic("Should this be handled?")
		}
	case *ast.Ellipsis:
		switch typedExpr2 := typedExpr.Elt.(type) {
		case *ast.Ident:
			return &typeDef{
				typeName:   typedExpr2.Name,
				isEllipsis: true,
			}
		default:
			panic("Should this be handled?")
		}
	default:
		panic("TODO")
	}
}

// TODO: it can be pointers, variadic params and such. Lots of work remaining
// here. But maybe I'm overcomplicating? Maybe just copy source code bytes from
// the range [typeExpr.Pos():typeExpr.End()]?
func makeTypeStr(typeExpr ast.Expr) string {
	switch typedExpr := typeExpr.(type) {
	case *ast.Ident:
		return typedExpr.Name
	case *ast.StarExpr:
		switch typedExpr2 := typedExpr.X.(type) {
		case *ast.Ident:
			// TODO: looks like to handle pointers I'll need to duplicate each
			// field, having a fN for the value and a pN for a pointer to it,
			// allowing pN to be nil.
			return "*" + typedExpr2.Name
		default:
			panic("Should this be handled?")
		}
	case *ast.Ellipsis:
		switch typedExpr2 := typedExpr.Elt.(type) {
		case *ast.Ident:
			return "[]" + typedExpr2.Name
		default:
			panic("Should this be handled?")
		}
	default:
		panic("TODO")
	}
}
