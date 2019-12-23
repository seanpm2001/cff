package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/types/typeutil"

	tmpls "bindata/src/go.uber.org/cff/internal/templates"
)

const _genTemplate = "gen.tmpl"

type generator struct {
	fset *token.FileSet

	typeIDs    *typeutil.Map // map[types.Type]int
	nextTypeID int

	// File path to which generated code is written.
	outputPath string
}

func newGenerator(fset *token.FileSet, outputPath string) *generator {
	return &generator{
		fset:       fset,
		typeIDs:    new(typeutil.Map),
		nextTypeID: 1,
		outputPath: outputPath,
	}
}

func (g *generator) GenerateFile(f *file) error {
	if len(f.Flows) == 0 {
		// Don't regenerate files that don't have flows.
		return nil
	}

	bs, err := ioutil.ReadFile(f.Filepath)
	if err != nil {
		return err
	}

	// Output buffer
	var buff bytes.Buffer

	// This tracks positioning information for the file.
	posFile := g.fset.File(f.AST.Pos())

	addImports := make(map[string]string) // import path -> name or empty for implicit name
	var lastOff int
	for _, flow := range f.Flows {
		// Everything from previous position up to this flow call.
		if _, err := buff.Write(bs[lastOff:posFile.Offset(flow.Pos())]); err != nil {
			return err
		}

		// Generate code for the flow.
		if err := g.generateFlow(f, flow, &buff, addImports); err != nil {
			return err
		}

		lastOff = posFile.Offset(flow.End())
	}

	// Write remaining code as-is.
	if _, err := buff.Write(bs[lastOff:]); err != nil {
		return err
	}

	// Parse the generated file and clean up.

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, posFile.Name(), buff.Bytes(), parser.ParseComments)
	if err != nil {
		// When there is a parsing error, we should output the file to a temporary file to help debugging
		// the template.
		tmpFile, tmpErr := ioutil.TempFile("", "*.go")
		if tmpErr == nil {
			if _, writeErr := buff.WriteTo(tmpFile); writeErr == nil {
				err = fmt.Errorf("%v\noutputted temporary file to %s", err, tmpFile.Name())
			}
		}

		return err
	}

	// Removing imports before adding "fmt", "context", and maybe "sync" since we
	// would cause a panic within astutil when removing cffImportPath as
	// AddNamedImport won't have an associated token.Pos.
	// See T3136343 for moar details.

	// The user code will have imports to cffImportPath but we should remove
	// them because it will be unused.
	if _, ok := f.UnnamedImports[cffImportPath]; ok {
		astutil.DeleteImport(fset, file, cffImportPath)
	}
	for _, name := range f.Imports[cffImportPath] {
		astutil.DeleteNamedImport(fset, file, name, cffImportPath)
	}

	newImports := make([]string, 0, len(addImports))
	for imp := range addImports {
		newImports = append(newImports, imp)
	}
	sort.Strings(newImports)

	for _, importPath := range newImports {
		astutil.AddNamedImport(fset, file, addImports[importPath], importPath)
	}

	// Remove build tag.
	for _, cg := range file.Comments {
		// Only consider comments before the "package" clause.
		if cg.Pos() >= file.Package {
			break
		}

		// Replace +build cff with +build !cff.
		for _, c := range cg.List {
			if strings.TrimSpace(strings.TrimPrefix(c.Text, "//")) == "+build cff" {
				// Tricking Phab not to consider this file to be generated.
				c.Text = "// @g" + "enerated by CFF"
				break
			}
		}
	}

	buff.Reset()
	err = format.Node(&buff, fset, file)

	return ioutil.WriteFile(g.outputPath, buff.Bytes(), 0644)
}

// generateFlow runs the CFF template for the given flow and writes it to w, modifying addImports if the template
// requires additional imports to be added.
func (g *generator) generateFlow(file *file, f *flow, w io.Writer, addImports map[string]string) error {
	t := tmpls.MustAssetString(_genTemplate)
	tmpl := template.Must(template.New("cff").Funcs(template.FuncMap{
		"type":     g.typePrinter(file, addImports),
		"typeHash": g.printTypeHash,
		"expr":     g.printExpr,
		"import": func(importPath string) string {
			if names := file.Imports[importPath]; len(names) > 0 {
				// already imported
				return names[0]
			}

			name, ok := addImports[importPath]
			if !ok {
				addImports[importPath] = ""
			}

			// TODO(abg): If the name is already taken, we will want to use
			// a named import. This can be done by having the compiler record
			// a list of unavailable names in the scope where cff.Flow was
			// called.
			if name == "" {
				name = filepath.Base(importPath)
			}

			return name
		},
	}).Parse(t))

	return tmpl.Execute(w, flowTemplateData{Flow: f})
}

// typePrinter returns the qualifier for the type to form an identifier using that type, modifying addImports if the
// type refers to a package that is not already imported
func (g *generator) typePrinter(f *file, addImports map[string]string) func(types.Type) string {
	return func(t types.Type) string {
		return types.TypeString(t, func(pkg *types.Package) string {
			for _, imp := range f.AST.Imports {
				ip, _ := strconv.Unquote(imp.Path.Value)

				if !isPackagePathEquivalent(pkg, ip) {
					continue
				}

				// Using a named import.
				if imp.Name != nil {
					return imp.Name.Name
				}

				// Unnamed imports use the package's name.
				return pkg.Name()
			}

			// The generated code needs a package (pkg) to be imported to form the qualifier, but it wasn't imported
			// by the user already and it isn't in this package (f.Package)
			if !isPackagePathEquivalent(pkg, f.Package.Types.Path()) {
				addImports[pkg.Path()] = pkg.Name()
				return pkg.Name()
			}

			// The type is defined in the same package
			return ""
		})
	}
}

func (g *generator) printTypeHash(t types.Type) string {
	return strconv.Itoa(g.typeID(t))
}

func (g *generator) printExpr(e ast.Expr) string {
	var buff bytes.Buffer
	format.Node(&buff, g.fset, e)
	return buff.String()
}

func (g *generator) typeID(t types.Type) int {
	if i := g.typeIDs.At(t); i != nil {
		return i.(int)
	}

	id := g.nextTypeID
	g.nextTypeID++
	g.typeIDs.Set(t, id)
	return id
}

type flowTemplateData struct {
	Flow *flow
}
