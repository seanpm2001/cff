package main

import (
	"errors"
	"fmt"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/cff/internal"
	"go.uber.org/cff/internal/flag"
	"go.uber.org/cff/internal/pkg"
	"github.com/jessevdk/go-flags"
	"go.uber.org/multierr"
)

// options defines the arguments needed to run the cff CLI.
type options struct {
	Files              []file    `long:"file" value-name:"FILE[=OUTPUT]"`
	InstrumentAllTasks bool      `long:"instrument-all-tasks"`
	GenMode            flag.Mode `long:"genmode" choice:"base" choice:"source-map" choice:"modifier" required:"no" default:"base"`
	Quiet              bool      `long:"quiet"`

	Args struct {
		ImportPath string `positional-arg-name:"importPath"`
	} `positional-args:"yes" required:"yes"`
}

// file is the value of the --file option.
// Two forms are supported:
//
//	--file=NAME
//	--file=NAME=OUTPUT
//
// For example,
//
//	--file=foo.go=_gen/foo.go --file=bar.go=_gen/bar.go
type file struct {
	Name       string // NAME portion of the argument
	OutputPath string // OUTPUT portion of the argument
}

func (f *file) String() string {
	if len(f.OutputPath) == 0 {
		return f.Name
	}
	return f.Name + "=" + f.OutputPath
}

func (f *file) UnmarshalFlag(name string) error {
	var output string
	if i := strings.IndexByte(name, '='); i >= 0 {
		name, output = name[:i], name[i+1:]
	}

	if len(name) == 0 {
		return errors.New("file name cannot be empty")
	}

	f.Name = name
	f.OutputPath = output
	return nil
}

func newCLIParser() (*flags.Parser, *options) {
	var opts options
	parser := flags.NewParser(&opts, flags.HelpFlag)
	parser.Name = "cff"

	// This is more readable than embedding the descriptions in the options
	// above.
	parser.FindOptionByLongName("file").Description =
		"Process only the file named NAME inside the package. All other files " +
			"will be ignored. NAME must be the name of the file, not the full path. " +
			"Optionally, OUTPUT may be provided as the path to which the generated " +
			"code for FILE will be written. By default, this defaults to adding a " +
			"_gen suffix to the file name."
	parser.FindOptionByLongName("instrument-all-tasks").Description =
		"Infer a name for tasks that do not specify cff.Instrument and opt-in " +
			"to instrumentation by default."
	parser.FindOptionByLongName("genmode").Description =
		"Use the specified CFF code generation mode."
	parser.Args()[0].Description = "Import path of a package containing CFF flows."

	return parser, &opts
}

func main() {
	log.SetFlags(0) // don't include timestamps
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("%+v", err)
	}
}

// The ArchiveLoaderFactory registers new flags with the CLI parser.
// For non-monorepo cases, we need to use the following as the LoaderFactory:
//
//	&pkg.GoPackagesLoaderFactory{
//	  BuildFlags: []string{"-tags=cff"},
//	}
var _loaderFactory pkg.LoaderFactory = new(pkg.ArchiveLoaderFactory)

func run(args []string) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("You've encountered a CFFv2 bug! Please report this http://t.uber.com/cff-bug")
			panic(err)
		}
	}()

	parser, f := newCLIParser()
	loader, err := _loaderFactory.RegisterFlags(parser)
	if err != nil {
		return err
	}

	if _, err := parser.ParseArgs(args); err != nil {
		return err
	}

	// For each --file, this is a mapping from FILE to OUTPUT.
	outputs := make(map[string]string)
	for _, file := range f.Files {
		if _, ok := outputs[file.Name]; ok {
			return fmt.Errorf(
				"invalid argument --file=%v: file already specified before", file)
		}
		outputs[file.Name] = file.OutputPath
	}

	fset := token.NewFileSet()
	pkgs, err := loader.Load(fset, f.Args.ImportPath)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	processor := internal.Processor{
		Fset:               fset,
		InstrumentAllTasks: f.InstrumentAllTasks,
		GenMode:            f.GenMode,
	}

	// If --file was provided, only the requested files will be processed.
	// Otherwise all files will be processed.
	hadFiles := len(f.Files) > 0
	var processed, errored int
	for _, pkg := range pkgs {
		for i, path := range pkg.CompiledGoFiles {
			name := filepath.Base(path)
			output, ok := outputs[name]
			if hadFiles && !ok {
				// --file was provided and this file wasn't included.
				continue
			}

			if len(output) == 0 {
				// x/y/foo.go => x/y/foo_gen.go
				output = filepath.Join(
					filepath.Dir(path),
					// foo.go => foo + _gen.go
					strings.TrimSuffix(name, filepath.Ext(name))+"_gen.go",
				)
			}

			processed++
			if perr := processor.Process(pkg, pkg.Syntax[i], output); perr != nil {
				errored++
				err = multierr.Append(err, perr)
			}
		}
	}

	if !f.Quiet {
		log.Printf("Processed %d files with %d errors", processed, errored)
	}
	return err
}
