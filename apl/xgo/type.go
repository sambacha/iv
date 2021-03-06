package xgo

import (
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/ktye/iv/apl"
)

// New returns an initialization function for the given type.
func New(t reflect.Type) create {
	return create{t}
}

type Value reflect.Value

// xgo values are copied by reference.
func (v Value) Copy() apl.Value {
	return v
}

func (v Value) String(f apl.Format) string {
	keys := v.Keys()
	if keys == nil {
		return fmt.Sprintf("xgo.Value (not a struct) %T", v)
	}
	var buf strings.Builder
	tw := tabwriter.NewWriter(&buf, 1, 0, 1, ' ', 0)
	for _, k := range keys {
		val := v.At(k)
		s := ""
		if val == nil {
			s = "?"
		} else {
			s = val.String(f)
		}
		fmt.Fprintf(tw, "%s:\t%s\n", k.String(f), s)
	}
	tw.Flush()
	s := buf.String()
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}

// Keys returns the field names, if the value is a struct.
// It does not return the method names.
// It returns nil, if the Value is not a struct.
func (v Value) Keys() []apl.Value {
	val := reflect.Value(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}
	t := val.Type()
	n := t.NumField()
	res := make([]apl.Value, n)
	for i := 0; i < n; i++ {
		res[i] = apl.String(t.Field(i).Name)
	}
	return res
}

func (v Value) Methods() []string {
	val := reflect.Value(v)
	t := val.Type()
	n := t.NumMethod()
	res := make([]string, n)
	for i := range res {
		res[i] = lower(t.Method(i).Name)
	}
	return res
}

// Field returns the value of a field or a method with the given name.
func (v Value) At(key apl.Value) apl.Value {
	name, ok := key.(apl.String)
	if ok == false {
		return nil
	}
	val := reflect.Value(v)
	var zero reflect.Value
	Name := upper(string(name))
	m := val.MethodByName(Name)
	if m != zero {
		return Function{Name: Name, Fn: m}
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}
	sf := val.FieldByName(Name)
	if sf == zero {
		return nil
	}
	rv, err := Convert(sf)
	if err != nil {
		return nil
	}
	return rv
}

func (v Value) Set(key apl.Value, fv apl.Value) error {
	field, ok := key.(apl.String)
	if ok == false {
		return fmt.Errorf("key must be a string")
	}
	val := reflect.Value(v).Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct: cannot set field")
	}
	sf := val.FieldByName(string(field))
	var zero reflect.Value
	if sf == zero {
		return fmt.Errorf("%v: field does not exist: %s", val.Type(), field)
	}
	sv, err := export(fv, sf.Type())
	if err != nil {
		return err
	}
	sf.Set(sv)
	return nil
}

type create struct {
	reflect.Type
}

func (t create) String(f apl.Format) string {
	return fmt.Sprintf("new %v", t.Type)
}
func (t create) Copy() apl.Value {
	return t
}

func (t create) Call(a *apl.Apl, L, R apl.Value) (apl.Value, error) {
	v := reflect.New(t.Type)
	return Value(v), nil
}

func upper(s string) string {
	return firstrune(s, unicode.ToUpper)
}

func lower(s string) string {
	return firstrune(s, unicode.ToLower)
}

func firstrune(s string, f func(r rune) rune) string {
	var buf strings.Builder
	for i, r := range s {
		if i == 0 {
			buf.WriteRune(f(r))
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
