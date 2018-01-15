package test

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"runtime"
)

func CreateFile(filename string, T *template.Template, data interface{}) error {
	var result bytes.Buffer
	if err := T.Execute(&result, data); err != nil {
		return err
	}

	content := result.Bytes()
	if runtime.GOOS == "windows" {
		content = bytes.Replace(content, []byte("\n"), []byte("\r\n"), -1)
	}

	if err := ioutil.WriteFile(filename, content, 0755); err != nil {
		return err
	}

	return nil
}

var (
	DPR_Template = template.Must(template.New("").Parse(`// AUTOMATICALLY GENERATED BY "delphi test"
program {{.Project}};

{$APPTYPE CONSOLE}
uses
  FastMM4,
  FastCode,
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
end.
`))

	DOF_Template = template.Must(template.New("").Parse(`
[FileVersion]
Version=7.0
[Compiler]
A=8
B=0
C=1
D=1
E=0
F=0
G=1
H=1
I=1
J=0
K=0
L=1
M=0
N=1
O=0
P=1
Q=0
R=0
S=1
T=0
U=0
V=1
W=0
X=1
Y=1
Z=1
ShowHints=1
ShowWarnings=1
UnsafeType=0
UnsafeCode=0
UnsafeCast=0
[Linker]
MapFile=3
OutputObjs=0
ConsoleApp=1
DebugInfo=1
RemoteSymbols=0
MinStackSize=16384
MaxStackSize=1048576
ImageBase=4194304
ExeDescription=
[Directories]
OutputDir={{.OutputDir}}
UnitOutputDir={{.BuildDir}}
PackageDLLOutputDir=
PackageDCPOutputDir=
SearchPath={{range $include := .Search}}{{$include}};{{end}}
DebugSourceDirs={{range $include := .Search}}{{$include}};{{end}}
Conditionals={{range $define := .Define}}{{$define}};{{end}}
`))
	CFG_Template = template.Must(template.New("").Parse(`
-$A8
-$B-
-$C+
-$D+
-$E-
-$F-
-$G+
-$H+
-$I+
-$J-
-$K-
-$L+
-$M-
-$N+
-$O-
-$P+
-$Q-
-$R-
-$S+
-$T-
-$U-
-$V+
-$W-
-$X+
-$YD
-$Z1
-GD
-cg
-vn
-AWinTypes=Windows;WinProcs=Windows;DbiTypes=BDE;DbiProcs=BDE;DbiErrs=BDE;
-H+
-W+
-M
-$M16384,1048576
-K$00400000
-E"{{.OutputDir}}"
-N"{{.BuildDir}}"
-LE"c:\Program Files (x86)\borland\delphi7\Projects\Bpl"
-LN"c:\Program Files (x86)\borland\delphi7\Projects\Bpl"
-U"{{range $include := .Search}}{{$include}};{{end}}"
-O"{{range $include := .Search}}{{$include}};{{end}}"
-I"{{range $include := .Search}}{{$include}};{{end}}"
-R"{{range $include := .Search}}{{$include}};{{end}}"
-D{{range $define := .Define}}{{$define}};{{end}}
-w-SYMBOL_LIBRARY
-w-SYMBOL_PLATFORM
-w-UNIT_LIBRARY
-w-UNIT_PLATFORM
-w-HRESULT_COMPAT
-w-UNSAFE_TYPE
-w-UNSAFE_CODE
-w-UNSAFE_CAST
`))
)
