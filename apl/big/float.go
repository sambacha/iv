package big

import (
	"math/big"
	"strings"

	"github.com/ktye/iv/apl"
	"github.com/ktye/iv/apl/big/bigfloat"
	"github.com/ktye/iv/apl/numbers"
)

type Float struct {
	*big.Float
}

func (f Float) String(a *apl.Apl) string {
	// TODO: Use Format %.15f => Text('f', 15), with leading -.
	return strings.Replace(f.Float.Text('g', -1), "-", "¯", -1)
}

func ParseFloat(s string, prec uint) (apl.Number, bool) {
	s = strings.Replace(s, "¯", "-", -1)
	z, _, err := big.NewFloat(0).SetPrec(prec).Parse(s, 10)
	if err != nil {
		return nil, false
	}
	return Float{z}, true
}

func (f Float) ToIndex() (int, bool) {
	if f.IsInt() == false {
		return 0, false
	}
	i, _ := f.Float.Int64()
	n := int(i)
	if big.NewFloat(float64(n)).Cmp(f.Float) == 0 {
		return n, true
	}
	return 0, false
}

func (f Float) cpy() *big.Float {
	return f.Float.Copy(f.Float)
}

func (f Float) Equals(R apl.Value) (apl.Bool, bool) {
	return f.Float.Cmp(R.(Float).Float) == 0, true
}

func (f Float) Less(R apl.Value) (apl.Bool, bool) {
	return f.Float.Cmp(R.(Float).Float) < 0, true
}

func (f Float) Add() (apl.Value, bool) {
	return f, true
}
func (f Float) Add2(R apl.Value) (apl.Value, bool) {
	z := f.cpy()
	return Float{z.Add(z, R.(Float).Float)}, true
}

func (f Float) Sub() (apl.Value, bool) {
	return Float{f.Float.Neg(f.Float)}, true
}
func (f Float) Sub2(R apl.Value) (apl.Value, bool) {
	z := f.cpy()
	return Float{z.Sub(z, R.(Float).Float)}, true
}

func (f Float) Mul() (apl.Value, bool) {
	return apl.Index(f.Float.Sign()), true
}
func (f Float) Mul2(R apl.Value) (apl.Value, bool) {
	z := f.cpy()
	return Float{z.Mul(z, R.(Float).Float)}, true
}

func (f Float) Div() (apl.Value, bool) {
	one := Float{f.cpy().SetInt64(1)}
	return one.Div2(f)
}
func (f Float) Div2(R apl.Value) (apl.Value, bool) {
	if f.Float.IsInf() {
		return numbers.Inf, true
	}
	if R.(Float).Float.IsInf() {
		return numbers.NaN, true
	}
	lz := f.Float.Sign() == 0
	rz := R.(Float).Float.Sign() == 0
	if lz && rz {
		return numbers.NaN, true
	} else if lz {
		z := f.cpy().SetInt64(0)
		return Float{z}, true
	} else if rz {
		return numbers.Inf, true
	}
	return Float{f.cpy().Quo(f.Float, R.(Float).Float)}, true
}

func (f Float) Pow() (apl.Value, bool) {
	z := bigfloat.Exp(f.Float)
	if z.IsInf() {
		return numbers.Inf, true
	}
	return Float{z}, true
}
func (f Float) Pow2(R apl.Value) (apl.Value, bool) {
	if f.Float.Cmp(f.Float) < 0 {
		return nil, false
	}
	z := bigfloat.Pow(f.Float, R.(Float).Float)
	if z.IsInf() {
		return numbers.Inf, true
	}
	return Float{z}, true
}