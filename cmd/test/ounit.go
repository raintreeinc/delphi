package test

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var (
	OUnit_Template = template.Must(template.New("").Parse(`// DO NOT MODIFY MANUALLY, AUTO-GENERATED
// delphi test -ounit {{.Project}} .

program {{.Project}};

{$APPTYPE CONSOLE}
{$WARN DUPLICATE_CTOR_DTOR OFF}
uses
  FastMM4,
{$if CompilerVersion < 32.0}
  FastCode,
{$ifend}
  rtTest,
  Forms,
  
  {{range $index, $test := .Tests}}
  {{$test.UnitName}},
  {{end}}

  rtFlag;

var
  lVerbose: Boolean;
begin
  Application.Initialize;

  lVerbose := Flag.Bool('v', False, 'verbose output');
  Flag.Check;

  {{range $test_index, $test := .Tests}}
  RunTests('{{$test.UnitName}}', [
    {{range $index, $func := $test.Funcs}}{{if $index}},{{end}}
    TestCase('{{$func}}', {{$test.UnitName}}.{{$func}})
    {{- end}}
  ], lVerbose);
  {{end}}

  if DebugHook <> 0 then
  begin
    WriteLn;
    Write('Press ENTER to quit');
    ReadLn;
  end;

end.
`))
)

func GenerateOUnit(tests []*TestFile, outfile string) error {
	type Func struct {
		Method string
		Call   string
	}

	type Test struct {
		UnitName  string
		ClassName string
		Funcs     []Func
	}

	type Tests struct {
		Project string
		Tests    []Test
	}

	dtests := Tests{}
	ext := filepath.Ext(outfile)
	dtests.Project = filepath.Base(outfile[:len(outfile)-len(ext)])
	for _, test := range tests {
		dtest := Test{}
		dtest.UnitName = test.UnitName
		dtest.ClassName = "T" + strings.Title(trimPrefix(trimSuffix(dtest.UnitName, "_Test"), "rt")) + "Test"
		for _, fn := range test.Funcs {
			dtest.Funcs = append(dtest.Funcs, Func{
				Method: trimPrefix(fn, "Test_"),
				Call:   fn,
			})
		}
		dtests.Tests = append(dtests.Tests, dtest)
	}

	var buf bytes.Buffer
	err := OUnit_Template.Execute(&buf, dtests)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outfile, windowsLineEndings(buf.Bytes()), 0755)
}