// +build ignore

package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	flagOut = flag.String("o", "doc.go", "out file")
	flagPkg = flag.String("pkg", os.Getenv("GOPATH")+"/src/github.com/S1mpleTheBest/hilink", "go package")
)

func main() {
	flag.Parse()

	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, *flagPkg, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	if len(pkgs) != 1 {
		log.Fatalf("invalid package count in %s", *flagPkg)
	}
	var pkgName string
	for pkgName = range pkgs {
	}
	if pkgName != "hilink" {
		log.Fatalf("invalid package name %s", pkgName)
	}

	buf := new(bytes.Buffer)
	buf.WriteString(hdr)

	buf.WriteString("var methodParamMap = map[string][]string{\n")

	for _, f := range pkgs[pkgName].Files {
		for _, d := range f.Decls {
			fd, typ, ok := getRecvType(d)
			if !ok || typ != "Client" || !fd.Name.IsExported() || fd.Name.Name == "Do" {
				continue
			}

			str := `"` + fd.Name.Name + `": {`
			for _, p := range fd.Type.Params.List {
				for _, n := range p.Names {
					str += `"` + n.Name + `",`
				}
			}
			str += "},\n"
			buf.WriteString(str)
		}
	}
	buf.WriteString("}\n\n")

	buf.WriteString("var methodCommentMap = map[string]string{\n")
	for _, f := range pkgs[pkgName].Files {
		for _, d := range f.Decls {
			fd, typ, ok := getRecvType(d)
			if !ok || typ != "Client" || !fd.Name.IsExported() || fd.Name.Name == "Do" {
				continue
			}

			str := `"` + fd.Name.Name + `": "` + strings.TrimSpace(strings.Replace(fd.Doc.Text(), "\n", " ", -1)) + "\",\n"
			buf.WriteString(str)
		}
	}
	buf.WriteString("}\n\n")

	if err = ioutil.WriteFile(*flagOut, buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("gofmt", "-s", "-w", *flagOut)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// getRecvType returns the receiver type.
func getRecvType(d ast.Decl) (*ast.FuncDecl, string, bool) {
	fd, ok := d.(*ast.FuncDecl)
	if !ok || fd.Recv == nil {
		return nil, "", false
	}

	se, ok := fd.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return nil, "", false
	}
	i, ok := se.X.(*ast.Ident)
	if !ok {
		return nil, "", false
	}

	return fd, i.Name, true
}

const (
	hdr = `package main

// Code generated by gen.go. DO NOT EDIT.

`
)
