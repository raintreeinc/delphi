package test

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

var (
	DUnit_Template = template.Must(template.New("").Parse(`unit {{.TestUnit}};
interface

uses
  {{range $test := .Tests -}}
  {{$test.UnitName}},
  {{ end }}
  TestFramework,
  rtTest,
  rtTestDUnit
  ;

type
{{range $test := .Tests}}
  // unit {{$test.UnitName}}
  {{$test.ClassName}} = class(TRtTestCase)
  published
  {{- range $func := $test.Funcs }}
    procedure {{ $func.Method }};
  {{- end }}
  end;
{{end}}

implementation

{{range $test := .Tests}}
{ {{$test.ClassName}} }
{{range $func := $test.Funcs }}
procedure {{$test.ClassName}}.{{$func.Method}};
begin
  RunTestCase({{$test.UnitName}}.{{$func.Call}});
end;
{{end}}
{{end}}

initialization
  {{range $test := .Tests -}}
  TestFramework.RegisterTest({{$test.ClassName}}.Suite);
  {{ end }}
end.
`))
)

func GenerateDUnit(tests []*TestFile, outfile string) error {
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
		TestUnit string
		Tests    []Test
	}

	dtests := Tests{}
	ext := filepath.Ext(outfile)
	dtests.TestUnit = filepath.Base(outfile[:len(outfile)-len(ext)])
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

	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()

	return DUnit_Template.Execute(f, dtests)
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix)) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

func trimPrefix(s, prefix string) string {
	if strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix)) {
		return s[len(prefix):]
	}
	return s
}
