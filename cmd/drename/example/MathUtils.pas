unit MathUtils;
interface

function lg(x: Extended): Extended;
function sn(x: Extended): Extended;
function cs(x: Extended): Extended;

implementation
uses
  Math;

function lg(x: Extended): Extended;
begin
  Result := Math.Log2(x);
end;

function sn(x: Extended): Extended;
begin
  Result := Sin(x);
end;

function cs(x: Extended): Extended;
begin
  Result := Cos(x);
end;

end.