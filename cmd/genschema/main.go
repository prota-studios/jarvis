package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/sirupsen/logrus"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"sort"
)

var (
	sheetId    = flag.String("sheet-id", "1pfZxMB9O0w1ByAsqCpqr9FybKG4e3VPJUCEXU6QfSxQ", "Sheet ID that has the info")
	credsFile  = flag.String("creds-file", "creds.json", "Google credentials to use to connect to the document.")
	srcFile    = flag.String("src-file", "pkg/connectors/connectors.go", "Source file for adding the schema file to...")
	structName = flag.String("struct-name", "Connection", "name of the struct in the source file to manage")
)

func main() {
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsFile(*credsFile))
	if err != nil {
		logrus.Fatal(err)
	}
	fields := []string{}
	if sheetId != nil {
		var s *sheets.BatchGetValuesResponse
		s, err = sheetsService.Spreadsheets.Values.BatchGet(*sheetId).Ranges("A1:ZZZ1").Do()
		if err != nil {
			logrus.Fatal(err)
		}
		if len(s.ValueRanges) != 1 {
			logrus.Fatal("no value ranges in the first row... is the sheet populated?")
		}

		for _, vr := range s.ValueRanges {
			for _, c := range vr.Values[0] {
				fields = append(fields, c.(string))
			}
		}
		fset := token.NewFileSet()
		var f *ast.File
		f, err = parser.ParseFile(fset, *srcFile, nil, parser.AllErrors)
		if err != nil {
			logrus.Fatal(err)
		}
		var cStruct *ast.StructType
		for _, dec := range f.Decls {
			d := dec.(*ast.GenDecl)
			if d != nil {
				if len(d.Specs) == 0 {
					logrus.Fatal("can't read starter struct")
				}
				spec := d.Specs[0].(*ast.TypeSpec)
				if spec != nil {
					if spec.Name.Name == *structName {
						cStruct = spec.Type.(*ast.StructType)
						sort.Strings(fields)
						cStruct.Fields.List = []*ast.Field{}
						for i, rr := range fields {

							ff := &ast.Field{
								Doc: nil,
								Names: []*ast.Ident{
									{
										NamePos: token.Pos(i),
										Name:    strcase.ToCamel(rr),
									},
								},
								Tag: &ast.BasicLit{
									Kind:  token.STRING,
									Value: fmt.Sprintf("`json:\"%s\"`", strcase.ToSnake(rr)),
								},
								Type: &ast.BasicLit{
									Kind:  token.STRING,
									Value: "string",
								},
							}
							cStruct.Fields.List = append(cStruct.Fields.List, ff)
						}
					}
				}
			}
		}
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, f)
		ioutil.WriteFile(*srcFile, buf.Bytes(), 0777)
	}

}
