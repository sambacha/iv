package primitives

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/ktye/iv/apl"
	"github.com/ktye/iv/apl/numbers"
	"github.com/ktye/iv/apl/operators"
)

//go:generate go run gen.go

type formatmap map[reflect.Type]string

var format5g formatmap = map[reflect.Type]string{
	reflect.TypeOf(numbers.Float(0)): "%.5g",
}
var formatJ formatmap = map[reflect.Type]string{
	reflect.TypeOf(numbers.Complex(0)): "%vJ%v",
}
var format5J formatmap = map[reflect.Type]string{
	reflect.TypeOf(numbers.Complex(0)): "%.5gJ%.5g",
}
var format2J formatmap = map[reflect.Type]string{
	reflect.TypeOf(numbers.Complex(0)): "%.2fJ%.2f",
}
var formatJR5 formatmap = map[reflect.Type]string{
	reflect.TypeOf(numbers.Float(0)):   "%.5g",
	reflect.TypeOf(numbers.Complex(0)): "%.5gJ%.5g",
}

var testCases = []struct {
	in, exp string
	formats map[reflect.Type]string
}{

	{"⍝ Basic numbers and arithmetics", "", nil},
	{"1", "1", nil},
	{"1+1", "2", nil},
	{"1-2", "¯1", nil}, // negative number
	{"¯1", "¯1", nil},
	{"1-¯2", "3", nil},
	{"1a90", "1a90", nil}, // a complex number
	{"1a60+1a300", "1a0", nil},
	{"1J1", "1.4142135623730951a45", nil},

	{"⍝ Vectors.", "", nil},
	{"1 2 3", "1 2 3", nil},
	{"1+1 2 3", "2 3 4", nil},
	{"1 2 3+¯1", "0 1 2", nil},
	{"1 2 3+4 5 6", "5 7 9", nil},

	{"⍝ Braces.", "", nil},
	{"1 2+3 4", "4 6", nil},
	{"1 (2+3) 4", "1 5 4", nil},
	{"(1 2)+3 4", "4 6", nil},
	{"1×2+3×4", "14", nil},
	{"1×(2+3)×4", "20", nil},
	{"(3×2)+3×4", "18", nil},
	{"3×2+3×4", "42", nil},

	{"⍝ Comparison", "", nil},
	{"1 2 3 4 5 > 2", "0 0 1 1 1", nil},     // greater than
	{"1 2 3 4 5 ≥ 3", "0 0 1 1 1", nil},     // greater or equal
	{"2 4 6 8 10<6", "1 1 0 0 0", nil},      // less than
	{"2 4 6 8 10≤6", "1 1 1 0 0", nil},      // less or equal
	{"1 2 3 ≠ 1.1 2 3", "1 0 0", nil},       // not equal
	{"3=3.1 3 ¯2 ¯3 3J0", "0 1 0 0 1", nil}, // equal
	{"2+2=2", "3", nil},                     // calculating with boolean values
	{"2×1 2 3=4 2 1", "0 2 0", nil},         // dyadic array
	{"-3<4", "¯1", nil},                     // monadic scalar
	{"-1 2 3=0 2 3", "0 ¯1 ¯1", nil},        // monadic array
	{"⍝ TODO Comparison tolerance is not implemented.", "", nil},

	{"⍝ Boolean, logical", "", nil},
	{"0 1 0 1 ^ 0 0 1 1", "0 0 0 1", nil}, // and
	{"0 1 0 1 ∧ 0 0 1 1", "0 0 0 1", nil}, // accept both ^ and ∧
	{"0^0 0 1 1", "0 0 0 0", nil},         // or
	{"0 0 1 1∨0 1 0 1", "0 1 1 1", nil},   // or
	{"1∨0 1 0 1", "1 1 1 1", nil},         // or
	{"0 0 1 1⍱0 1 0 1", "1 0 0 0", nil},   // nor
	{"0 0 1 1⍲0 1 0 1", "1 1 1 0", nil},   // nand
	{"~0", "1", nil},                      // scalar not
	{"~1.0", "0", nil},                    // scalar not
	{"~0 1", "1 0", nil},                  // array not
	{"⍝ TODO: dyadic: least common multiple, greatest common divisor", "", nil},
	// {"15 1 2 7 ^ 35 1 4 0", "105 1 4 0", nil}, // least common multiple
	// {"2 3 4 ∧ 0j1 1j2 2j3", "0J2 3J6 8J12", nil},// least common multiple
	// {"2j2 2j4 ∧ 5j5 4j4", "10J10 ¯4J12", nil},// least common multiple
	// {"15 1 2 7 ∨ 35 1 4 0", "5 1 2 7", nil}, // greatest common divisor

	{"⍝ Multiple expressions.", "", nil},
	{"1⋄2⋄3", "1", nil},

	{"⍝ Iota and reshape.", "", nil},
	{"⍳5", "1 2 3 4 5", nil},       // index generation
	{"⍳0", "", nil},                // empty array
	{"⍴⍳5", "5", nil},              // shape
	{"⍴5", "", nil},                // shape of scalar is empty
	{"⍴⍴5", "0", nil},              // shape of empty is 0
	{"⍴⍳0", "0", nil},              // empty array has zero dimensions
	{"⍴⍴⍳0", "1", nil},             // rank of empty array is 1
	{"2 3⍴1", "1 1 1\n1 1 1", nil}, // shape

	{"⍝ Magnitude, Residue, Ceil, Floor, Min, Max", "", nil},
	{"|1 ¯2 ¯3.2 2.2a20", "1 2 3.2 2.2", nil},                  // magnitude
	{"3 3 ¯3 ¯3|¯5 5 ¯4 4", "1 2 ¯1 ¯2", nil},                  // residue
	{"0.5|3.12 ¯1 ¯0.6", "0.12 0 0.4", format5g},               // residue
	{"¯1 0 1|¯5.25 0 2.41", "¯0.25 0 0.41", format5g},          // residue
	{"1j2|2j3 3j4 5j6", "1J1 ¯1J1 0J1", formatJ},               // complex residue
	{"4J6|7J10", "3J4", formatJ},                               // complex residue
	{"¯10 7J10 .3|17 5 10", "¯3 ¯5J7 0.1", formatJR5},          // residue
	{"⌊¯2.3 0.1 100 3.3", "¯3 0 100 3", nil},                   // floor
	{"⌊0.5 + 0.4 0.5 0.6", "0 1 1", nil},                       // floor
	{"⌊1j3.2 3.3j2.5 ¯3.3j¯2.5", "1J3 3J2 ¯3J¯3", formatJ},     // complex floor
	{"⌊1.5J2.5", "2J2", formatJ},                               // complex floor
	{"⌊1J2 1.2J2.5 ¯1.2J¯2.5", "1J2 1J2 ¯1J¯3", formatJ},       // complex floor
	{"⌈¯2.7 3 .5", "¯2 3 1", nil},                              // ceil
	{"⌈1.5J2.5", "1J3", formatJ},                               // complex ceil
	{"⌈1J2 1.2J2.5 ¯1.2J¯2.5", "1J2 1J3 ¯1J¯2", formatJ},       // complex ceil
	{"⌈¯2.3 0.1 100 3.3", "¯2 1 100 4", nil},                   // ceil
	{"⌈1.2j2.5 1.2j¯2.5", "1J3 1J¯2", formatJ},                 // ceil
	{"5⌊4 5 7", "4 5 5", nil},                                  // min
	{"¯2⌊¯3", "¯3", nil},                                       // min
	{"3.3 0 ¯6.7⌊3.1 ¯4 ¯5", "3.1 ¯4 ¯6.7", nil},               // min
	{"¯2.1 0.1 15.3 ⌊ ¯3.2 1 22", "¯3.2 0.1 15.3", nil},        // min
	{"5⌈4 5 7", "5 5 7", nil},                                  // max
	{"¯2⌈¯3", "¯2", nil},                                       // max
	{"3.3 0 ¯6.7⌈3.1 ¯4 ¯5", "3.3 0 ¯5", nil},                  // max
	{"¯2.01 0.1 15.3 ⌈ ¯3.2 ¯1.1 22.7", "¯2.01 0.1 22.7", nil}, // max

	{"⍝ Factorial, gamma, binomial.", "", nil},
	{"!4", "24", nil},                                       // factorial
	{"!1 2 3 4 5", "1 2 6 24 120", nil},                     // factorial
	{"!3J2", "¯3.0115J1.7702", format5J},                    // complex gamma
	{"!.5 ¯.05", "0.88623 1.0315", format5g},                // real gamma (APL2 doc: "0.0735042656 1.031453317"?)
	{"2!5", "10", nil},                                      // binomial
	{"3.2!5.2", "10.92", format5g},                          // binomial, floats with beta function
	{"3!¯2", "¯4", nil},                                     // binomial, negative R
	{"¯6!¯3", "¯10", nil},                                   // binomial negative L and R
	{"2 3 4!6 18 24", "15 816 10626", format5g},             // binomial
	{"3!.05 2.5 ¯3.6", "0.015437 0.3125 ¯15.456", format5g}, // binomial
	{"0 1 2 3!3", "1 3 3 1", nil},                           // binomial coefficients
	{"2!3J2", "1J5", format5J},                              // binomial complex

	{"⍝ Match, Not match, tally, depth", "", nil},
	{"≡5", "0", nil},          // depth
	{"≡⍳0", "1", nil},         // depth for empty array
	{`≡"alpha"`, "0", nil},    // a string is a scalarin APLv.
	{"≢2 3 4⍴⍳10", "2", nil},  // tally
	{"≢2", "1", nil},          // tally
	{"≢⍳0", "0", nil},         // tally
	{"1 2 3≡1 2 3", "1", nil}, // match
	{"3≡1⍴3", "0", nil},       // match shape
	{`""≡⍳0`, "0", nil},       // match empty string
	{`''≡⍳0`, "1", nil},       // this is false in other APLs (here '' is an empty array).
	{"2.0-1.0≡1>0", "1", nil}, // compare numbers of different type
	{"1≢2", "1", nil},         // not match
	{"1≢1", "0", nil},         // not match
	{"3≢1⍴3", "1", nil},       // not match
	{`""≢⍳0`, "1", nil},       // not match

	{"⍝ Left tack, right tack. ⊢ ⊣", "", nil},
	{"⊣1 2 3", "1 2 3", nil},      // monadic left: same
	{"3 2 1⊣1 2 3", "3 2 1", nil}, // dyadic left
	{"1 2 3⊢3 2 1", "3 2 1", nil}, // dyadic right
	{"⊢4", "4", nil},              // monadic right: same
	{"⊣/1 2 3", "1", nil},         // ⊣ reduction selects the first sub array
	{"⊢/1 2 3", "3", nil},         // ⊢ reduction selects the last sub array
	{"⊣/2 3⍴⍳6", "1 4", nil},      // ⊣ reduction over array
	{"⊢/2 3⍴⍳6", "3 6", nil},      // ⊢ reduction over array

	{"⍝ Array expressions.", "", nil},
	{"-⍳3", "¯1 ¯2 ¯3", nil},

	{"⍝ Ravel, enlist, catenate, join", "", nil},
	{",2 3⍴⍳6", "1 2 3 4 5 6", nil},     // ravel
	{"∊2 3⍴⍳6", "1 2 3 4 5 6", nil},     // enlist (identical for simple arrays)
	{"⍴,3", "1", nil},                   // scalar ravel
	{"⍴,⍳0", "0", nil},                  // ravel empty array
	{"1 2 3,4 5 6", "1 2 3 4 5 6", nil}, // catenate
	{`"abc",1 2`, `abc 1 2`, nil},
	{"(2 3⍴⍳6),2 2⍴7 8 9 10", "1 2 3 7 8\n4 5 6 9 10", nil},
	{"2 3≡2,3", "1", nil},                // catenate vector result
	{"(1 2 3,4 5 6)≡⍳6", "1", nil},       // catenate vector result
	{"0,2 3⍴1", "0 1 1 1\n0 1 1 1", nil}, // catenate scalar and array
	{"⍝ TODO ravel with axis", "", nil},
	{"⍝ TODO laminate", "", nil},

	{"⍝ Decode", "", nil},
	{"3⊥1 2 1", "16", nil},                                // decode
	{"3⊥4 3 2 1", "142", nil},                             // decode
	{"2⊥1 1 1 1", "15", nil},                              // decode
	{"1 2 3⊥3 2 1", "25", nil},                            // decode mixed radix
	{"1J1⊥1 2 3 4", "5J9", format5J},                      // decode complex
	{"24 60 60⊥2 23 12", "8592", nil},                     // convert 2h23min12s to seconds
	{"(2 1⍴2 10)⊥3 2⍴ 1 4 0 3 1 2", "5 24\n101 432", nil}, // decode arrays

	{"⍝ Reduce, reduce first, scan, scan first.", "", nil},
	{"+/1 2 3", "6", nil},                // reduce vector
	{"+⌿1 2 3", "6", nil},                // reduce vector (first axis)
	{"+/2 3 1 ⍴⍳6", "1 2 3\n4 5 6", nil}, // special case: reshape if axis length is 1
	{"⍴+/3", "", nil},                    // reduce scalar result
	{"⍴+/1 1⍴3", "1", nil},               // reduce vector result
	{"+/2 3⍴⍳6", "6 15", nil},            // reduce matrix
	{"+⌿2 3⍴⍳6", "5 7 9", nil},           // reduce matrix (first axis)
	{`+\1 2 3 4 5`, "1 3 6 10 15", nil},  // scan vector
	{`+\2 3⍴⍳6`, "1 3 6\n4 9 15", nil},   // scan array
	{`+⍀2 3⍴⍳6`, "1 2 3\n5 7 9", nil},    // scan first
	{`-\1 2 3`, "1 ¯1 2", nil},           // scan

	{"⍝ Pi times, circular, trigonometric", "", nil},
	{"○0 1 2", "0 3.1416 6.2832", format5g},         // pi times
	{"*○0J1", "¯1.00J0.00", format2J},               // Euler identity
	{"0 ¯1 ○ 1", "0 1.5708", format5g},              //
	{"1○(○1)÷2 3 4", "1 0.86603 0.70711", format5g}, //
	{"2○(○1)÷3", "0.5", format5g},                   //
	{"9 11○3.5J¯1.2", "3.5 ¯1.2", nil},              //
	// {"9 11∘.○3.5J¯1.2 2J3 3J4", "3.5 2 3\n¯1.2 3 4", nil}, // TODO outer product
	{"¯4○¯1", "0", nil},            //
	{"3○2", "¯2.185", format5g},    //
	{"2○1", "0.5403", format5g},    //
	{"÷3○2", "¯0.45766", format5g}, //
	{"1○○30÷180", "0.5", format5g},
	{"2○○45÷180", "0.70711", format5g},
	{"¯1○1", "1.5708", format5g},
	{"¯2○.54032023059", "0.99998", format5g},
	{"(¯1○.5)×180÷○1", "30", format5g},
	{"(¯3○1)×180÷○1", "45", format5g},
	{"5○1", "1.1752", format5g},
	{"6○1", "1.5431", format5g},
	{"¯5○1.175201194", "1", format5g},
	{"¯6○1.543080635", "1", format5g},

	{"⍝ Basic operators.", "", nil},
	{"+/1 2 3", "6", nil},                            // plus reduce
	{"1 2 3 +.× 4 3 2", "16", nil},                   // scalar product
	{"(2 3⍴⍳6) +.× 3 2⍴5+⍳6", "52 58\n124 139", nil}, // matrix multiplication

	{"⍝ Format as a string, Execute", "", nil},
	{"⍕10", "10", nil},   // format as string
	{`⍎"1+1"`, "2", nil}, // evaluate expression
	{"⍝ TODO: dyadic format with specification.", "", nil},
	{"⍝ TODO: dyadic execute with namespace.", "", nil},

	{"⍝ Grade up, grade down, sort.", "", nil},
	{"⍋23 11 13 31 12", "2 5 3 1 4", nil},                             // grade up
	{"⍋23 14 23 12 14", "4 2 5 1 3", nil},                             // identical subarrays
	{"⍋5 3⍴4 16 37 2 9 26 5 11 63 3 18 45 5 11 54", "2 4 1 5 3", nil}, // grade up rank 2                   //
	{"⍋22.5 1 15 3 ¯4", "5 2 4 3 1", nil},                             // grade up
	{"⍒33 11 44 66 22", "4 3 1 5 2", nil},                             // grade down                                                  //
	{"⍋'alpha'", "1 5 4 2 3", nil},                                    // strings grade up
	{"'ABCDE'⍒'BEAD'", "2 4 1 3", nil},                                // grade down with collating sequence
	{"⍝ TODO dyadic grade up/down is only implemented for vector L", "", nil},
	// {"A←423 11 13 31 12⋄A[⍋A]","11 12 13 23 31",nil}, // sort

	{"⍝ Reverse, rotate", "", nil},
	{"⌽1 2 3 4 5", "5 4 3 2 1", nil},                                                  // reverse vector
	{"⌽2 3⍴⍳6", "3 2 1\n6 5 4", nil},                                                  // reverse matrix
	{"⊖2 3⍴⍳6", "4 5 6\n1 2 3", nil},                                                  // reverse first
	{"⌽'DESSERTS'", "S T R E S S E D", nil},                                           // reverse strings
	{"1⌽1 2 3 4", "2 3 4 1", nil},                                                     // rotate vector
	{"10⌽1 2 3 4", "3 4 1 2", nil},                                                    // rotate vector
	{"¯1⌽1 2 3 4", "4 1 2 3", nil},                                                    // rotate vector negative
	{"(-7)⌽1 2 3 4", "2 3 4 1", nil},                                                  // rotate vector negative
	{"1 2⌽2 3⍴⍳6", "2 3 1\n6 4 5", nil},                                               // rotate array
	{"(2 2⍴2 ¯3 3 ¯2)⌽2 2 4⍴⍳16", "3 4 1 2\n6 7 8 5\n\n12 9 10 11\n15 16 13 14", nil}, // rotate array
	{"(2 3⍴2 ¯3 3 ¯2 1 2)⊖2 2 3⍴⍳12", "1 8 9\n4 11 6\n\n7 2 3\n10 5 12", nil},         // rotate array

	{"⍝ Variable assignments.", "", nil},
	{"X←3", "", nil},          // assign a number
	{"-X←3", "¯3", nil},       // assign a value and use it
	{"X←3⋄X←4", "", nil},      // assign and overwrite
	{"X←3⋄⎕←X", "3", nil},     // assign and check
	{"f←+", "", nil},          // assign a function
	{"f←+⋄⎕←3 f 3", "6", nil}, // assign a function and apply
	{"X←4⋄⎕←÷X", "0.25", nil}, // assign and use it in another expr

	{"⍝ Bracket indexing.", "", nil},
	{"⍝ TODO: A←⍳6 ⋄ ⎕←A[1]", "", nil}, //{"A←⍳6 ⋄ ⎕←A[1]", "x", nil}, // simple indexing

	{"⍝ IBM APL Language, 3rd edition, June 1976.", "", nil},
	{"1000×(1+.06÷1 4 12 365)*10×1 4 12 365", "1790.8476965428547 1814.0184086689414 1819.3967340322804 1822.0289545386752", nil},
	// the original prints as: "1790.85 1413.02 1819.4 1822.03"
	{"Area ← 3×4\nX←2+⎕←3×Y←4\nX\nY", "12\n14\n4", nil},

	{"⍝ Lambda expressions.", "", nil},
	{"{2×⍵}3", "6", nil},           // lambda in monadic context
	{"2{⍺+3{⍺×⍵}⍵+2}2", "14", nil}, // nested lambas
	{"2{(⍺+3){⍺×⍵}⍵+⍺{⍺+1+⍵}1+2}2", "40", nil},
	{"1{1+⍺{1+⍺{1+⍺+⍵}1+⍵}1+⍵}1", "7", nil},
	{"2{}4", "", nil},          // empty lambda expression ignores arguments
	{"{⍺×⍵}/2 3 4", "24", nil}, // TODO

	// Tool of thought.

	// github.com/DhavalDalal/APL-For-FP-Programmers
	// filter←{(⍺⍺¨⍵)⌿⍵} // 01-primes
	// primes1←{(2=+⌿0=X∘.|X)⌿X←⍳⍵} // 01-primes
	// primes2←{(~X∊X∘.×X)⌿X←2↓⍳⍵} // 01-primes
	// ⎕IO←0 ⋄ sieve ← {⍸⊃{~⍵[⍺]:⍵ ⋄ 0@(⍺×2↓⍳⌈(≢⍵)÷⍺)⊢⍵}/⌽(⊂0 0,(⍵-2)⍴1),⍳⍵} // 02-sieve
	// ⎕IO←0 ⋄ triples←{{⍵/⍨(2⌷x)=+⌿2↑x←×⍨⍵}⍉↑,1+⍳⍵ ⍵ ⍵}// 03-pythagoreans
	// ⎕IO←0 ⋄ '-:'⊣@(' '=⊢)¨(14⍴(4⍴1),0)(17⍴1 1 0)\¨⊂⍉(⎕D,6↑⎕A)[(12⍴16)⊤?10⍴2*48] // 04-MacAddress
	// life←{⊃1 ⍵∨.∧3 4=+⌿,1 0 ¯1∘.⊖1 0 ¯1⌽¨⊂⍵} // 05-life
	// life2←{3=s-⍵∧4=s←{+/,⍵}⌺3 3⊢⍵} // 05-life
}

func testCompare(got, exp string) bool {
	got = strings.TrimSpace(got)
	gotlines := strings.Split(got, "\n")
	explines := strings.Split(exp, "\n")
	if len(gotlines) != len(explines) {
		return false
	}
	for i, g := range gotlines {
		e := explines[i]
		gf := strings.Fields(g)
		ef := strings.Fields(e)
		if len(gf) != len(ef) {
			return false
		}
		for k := range gf {
			if gf[k] != ef[k] {
				return false
			}
		}
	}
	return true
}

func TestApl(t *testing.T) {
	// Compare result with expectation but ignores differences in whitespace.
	for i, tc := range testCases {

		if strings.HasPrefix(tc.in, "⍝") {
			if strings.HasPrefix(tc.in, "⍝ TODO") {
				t.Log(tc.in)
			} else {
				t.Log("\n" + tc.in)
			}
			continue
		}

		var buf strings.Builder
		a := apl.New(&buf)
		numbers.Register(a)
		Register(a)
		operators.Register(a)

		// Set numeric formats.
		if tc.formats != nil {
			for t, f := range tc.formats {
				if num, ok := a.Tower.Numbers[t]; ok {
					num.Format = f
					a.Tower.Numbers[t] = num
				}
			}
		}

		lines := strings.Split(tc.in, "\n")
		for k, s := range lines {
			t.Logf("\t%s", s)
			if err := a.ParseAndEval(s); err != nil {
				t.Fatalf("tc%d:%d: %s: %s\n", i+1, k+1, tc.in, err)
			}
		}
		got := buf.String()
		t.Log(got)

		g := got
		g = spaces.ReplaceAllString(g, " ")
		g = newline.ReplaceAllString(g, "\n")
		g = strings.TrimSpace(g)
		if g != tc.exp {
			fmt.Printf("%q != %q\n", g, tc.exp)
			t.Fatalf("tc%d:\nin>\n%s\ngot>\n%s\nexpected>\n%s", i+1, tc.in, got, tc.exp)
		}
	}
}

var spaces = regexp.MustCompile(`  *`)
var newline = regexp.MustCompile(`\n *`)
