package domain

import "github.com/ktye/iv/apl"

// ToScalar accepts scalars and converts single element arrays to scalars.
func ToScalar(child SingleDomain) SingleDomain {
	return scalar{child, true}
}

// IsScalar accepts values that are not arrays.
func IsScalar(child SingleDomain) SingleDomain {
	return scalar{child, false}
}

type scalar struct {
	child   SingleDomain
	convert bool
}

func (s scalar) To(a *apl.Apl, V apl.Value) (apl.Value, bool) {
	v := V
	if ar, ok := V.(apl.Array); ok {
		if s.convert == false {
			return V, false
		}
		if n := ar.Size(); n != 1 {
			return V, false
		}
		v = ar.At(0)
	}
	return propagate(a, v, s.child)
}
func (s scalar) String(f apl.Format) string {
	name := "scalar"
	if s.convert {
		name = "toscalar"
	}
	if s.child == nil {
		return name
	}
	return name + " " + s.child.String(f)
}
