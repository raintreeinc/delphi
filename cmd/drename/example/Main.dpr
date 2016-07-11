program Main;

{$APPTYPE CONSOLE}

uses
  SysUtils,
  MathUtils;

type
  {$IFDEF EXTENDED}
  state = record
    X, Y: Extended;
  end;
  {$ELSE}
  state = record
    X, Y, Z: Extended;
  end;
  {$ENDIF}

var
  s1, s2: state;
begin
  s1.X := 0;
  s1.Y := 0;

  s2.X := cs(s1.X);
  s2.Y := sn(s1.Y);

  s2.X := MathUtils.lg(s1.X);
end.
