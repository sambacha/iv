package primitives

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/ktye/iv/apl"
	"github.com/ktye/iv/apl/big"
	"github.com/ktye/iv/apl/numbers"
	"github.com/ktye/iv/apl/operators"
	aplstrings "github.com/ktye/iv/apl/strings"
	"github.com/ktye/iv/apl/xgo"
)

//go:generate go run gen.go

var testCases = []struct {
	in, exp string
	flag    int
}{
	{"⍝ Basic numbers and arithmetics", "", 0},
	{"1", "1", 0},
	{"1+1", "2", 0},
	{"1-2", "¯1", 0}, // negative number
	{"¯1", "¯1", 0},
	{"1-¯2", "3", 0},
	{"1a90", "0J1", cmplx}, // a complex number
	{"1a60+1a300", "1J0", cmplx},
	{"1J1", "1J1", cmplx},

	{"⍝ Vectors.", "", 0},
	{"1 2 3", "1 2 3", 0},
	{"1+1 2 3", "2 3 4", 0},
	{"1 2 3+¯1", "0 1 2", 0},
	{"1 2 3+4 5 6", "5 7 9", 0},

	{"⍝ Braces.", "", 0},
	{"1 2+3 4", "4 6", 0},
	{"(1 2)+3 4", "4 6", 0},
	{"1×2+3×4", "14", 0},
	{"1×(2+3)×4", "20", 0},
	{"(3×2)+3×4", "18", 0},
	{"3×2+3×4", "42", 0},
	// {"1 (2+3) 4", "1 5 4", 0}, not supported
	// {"1 2 (+/1 2 3) 4 5", "1 2 6 4 5", 0},

	{"⍝ Comparison", "", 0},
	{"1 2 3 4 5 > 2", "0 0 1 1 1", 0},         // greater than
	{"1 2 3 4 5 ≥ 3", "0 0 1 1 1", 0},         // greater or equal
	{"2 4 6 8 10<6", "1 1 0 0 0", 0},          // less than
	{"2 4 6 8 10≤6", "1 1 1 0 0", 0},          // less or equal
	{"1 2 3 ≠ 1.1 2 3", "1 0 0", 0},           // not equal
	{"3=3.1 3 ¯2 ¯3 3J0", "0 1 0 0 1", cmplx}, // equal
	{"2+2=2", "3", 0},                         // calculating with boolean values
	{"2×1 2 3=4 2 1", "0 2 0", 0},             // dyadic array
	{"-3<4", "¯1", 0},                         // monadic scalar
	{"-1 2 3=0 2 3", "0 ¯1 ¯1", 0},            // monadic array
	{"⍝ TODO Comparison tolerance is not implemented.", "", 0},

	{"⍝ Boolean, logical", "", 0},
	{"0 1 0 1 ^ 0 0 1 1", "0 0 0 1", 0}, // and
	{"0 1 0 1 ∧ 0 0 1 1", "0 0 0 1", 0}, // accept both ^ and ∧
	{"0^0 0 1 1", "0 0 0 0", 0},         // or
	{"0 0 1 1∨0 1 0 1", "0 1 1 1", 0},   // or
	{"1∨0 1 0 1", "1 1 1 1", 0},         // or
	{"0 0 1 1⍱0 1 0 1", "1 0 0 0", 0},   // nor
	{"0 0 1 1⍲0 1 0 1", "1 1 1 0", 0},   // nand
	{"~0", "1", 0},                      // scalar not
	{"~1.0", "0", 0},                    // scalar not
	{"~0 1", "1 0", 0},                  // array not

	{"⍝ Least common multiple, greatest common divisor", "", 0},
	{"30^36", "180", small},                     // lcm
	{"0^3", "0", 0},                             // lcm with 0
	{"3^0", "0", 0},                             // lcm with 0
	{"15 1 2 7 ^ 35 1 4 0", "105 1 4 0", small}, // least common multiple
	{"30∨36", "6", small},                       // gcm
	{"15 1 2 7 ∨ 35 1 4 0", "5 1 2 7", small},   // greatest common divisor
	{"0∨3", "3", 0},                             // gcm with 0
	{"3∨0", "3", 0},                             // gcm with 0
	{"3^3.6", "18", short | small},              // lcm
	//{"¯29J53^¯1J107", "¯853J¯329", 0},          // lcm
	//{"2 3 4 ∧ 0j1 1j2 2j3", "0J2 3J6 8J12", 0}, // least common multiple
	//{"2j2 2j4 ∧ 5j5 4j4", "10J10 ¯4J12", 0},    // least common multiple
	{"3∨3.6", "0.6", small}, // gcm
	//{"¯29J53∨¯1J107", "7J1", 0},                // gcm
	{"⍝ TODO: lcm and gcm of float and complex", "", 0},

	{"⍝ Multiple expressions.", "", 0},
	{"1⋄2⋄3", "1\n2\n3", 0},
	{"1⋄2", "1\n2", 0},
	{"1 2⋄3 4", "1 2\n3 4", 0},
	{"X←3 ⋄ Y←4", "", 0},

	{"⍝ Index origin.", "", 0},
	{"⎕IO←0 ⋄ ⍳3", "0 1 2", 0},
	{"⎕IO", "1", 0},
	{"⎕IO←0 ⋄ ⎕IO", "0", 0},

	{"⍝ Type, typeof.", "", 0},
	{"⌶'a'", "apl.String", 0},

	{"⍝ Bracket indexing.", "", 0},
	{"A←⍳6 ⋄ A[1]", "1", 0},
	{"A←2 3⍴⍳6 ⋄ A[1;] ⋄ ⍴A[1;]", "1 2 3\n3", 0},
	{"A←2 3⍴⍳6 ⋄ A[2;3]", "6", 0},
	{"A←2 3⍴⍳6 ⋄ A[2;2 3]", "5 6", 0},
	{"A←2 3⍴⍳6 ⋄ ⍴⍴A[2;3]", "0", 0},
	{"A←2 3 4 ⋄ A[]", "2 3 4", 0},
	{"⎕IO←0 ⋄ A←2 3⍴⍳6 ⋄ A[1;2]", "5", 0},
	{"5 6 7[2+1]", "7", 0},
	{"(2×⍳3)[2]", "4", 0},
	{"A←2 3 ⍴⍳6⋄A[A[1;1]+1;]", "4 5 6", 0},
	{"A←1 2 3⋄A[3]+1", "4", 0},
	{"A←1 2 3⋄1+A[3]", "4", 0},

	{"⍝ Scalar primitives with axis", "", 0},
	{"(2 3⍴⍳6)+[2]1 2 3", "2 4 6\n5 7 9", 0},
	{"1 2 3 +[2] 2 3⍴⍳6", "2 4 6\n5 7 9", 0},
	{"K←2 3⍴.1×⍳6⋄J←2 3 4⍴⍳24⋄N←J+[1 2]K⋄⍴N⋄N[1;2;3]⋄N[2;3;4]", "2 3 4\n7.2\n24.6", 0},

	{"⍝ Iota and reshape.", "", 0},
	{"⍳5", "1 2 3 4 5", 0},       // index generation
	{"⍳0", "", 0},                // empty array
	{"⍴⍳5", "5", 0},              // shape
	{"⍴5", "", 0},                // shape of scalar is empty
	{"⍴⍴5", "0", 0},              // shape of empty is 0
	{"⍴⍳0", "0", 0},              // empty array has zero dimensions
	{"⍴⍴⍳0", "1", 0},             // rank of empty array is 1
	{"2 3⍴1", "1 1 1\n1 1 1", 0}, // shape
	{"3⍴⍳0", "0 0 0", 0},         // reshape empty array
	{"⍴0 2⍴⍳0", "0 2", 0},        // reshape empty array
	{"⍴3 0⍴⍳0", "3 0", 0},        // reshape empty array
	{"⍴3 0⍴3", "3 0", 0},         // reshape empty array

	{"⍝ Where, interval index", "", 0},
	{"⍸1 0 1 0 0 0 0 1 0", "1 3 8", 0},
	{"⍸'e'='Pete'", "2 4", 0},
	{"⍸1=1", "1", 0},
	{"10 20 30⍸11 1 31 21", "1 0 3 2", 0},
	{"'AEIOU'⍸'DYALOG'", "1 5 1 3 4 2", 0},
	{"0.8 2 3.3⍸1.3 1.9 0.7 4 .6 3.2", "1 1 0 3 0 2", 0},

	{"⍝ Enlist, membership", "", 0},
	{"∊⍴⍳0", "0", 0},
	{"⍴∊⍴⍳0", "1", 0},
	{"∊2 3⍴⍳6", "1 2 3 4 5 6", 0},
	{"'BANANA'∊'AN'", "0 1 1 1 1 1", 0},
	{"5 1 2∊6 5 4 1 9", "1 1 0", 0},
	{"(2 3⍴8 3 5 8 4 8)∊1 8 9 3", "1 1 0\n1 0 1", 0},
	{"8 9 7 3∊⍳0", "0 0 0 0", 0},
	{"3.1 5.1 7.1∊2 2⍴1.1 3.1 5.1 4.1", "1 1 0", 0},
	{"19∊'CLUB'", "0", 0},
	{"'BE'∊'BOF'", "1 0", 0},
	{"'NADA'∊⍳0", "0 0 0 0", 0},
	{"(⌈/⍳0)∊⌊/⍳0", "0", 0},
	{"5 10 15∊⍳10", "1 1 0", 0},

	{"⍝ Without", "", 0},
	{"1 2 3 4 5~2 3 4", "1 5", 0},
	{"'RHYME'~'MYTH'", "R E", 0},
	{"1 2~⍳0", "1 2", 0},
	{"1~3", "1", 0},
	{"3~3", "", 0},
	{"⍴⍳0~1 2", "0", 0},
	{"5 10 15~⍳10", "15", 0},
	{"3 1 4 1 5 5~3 1 4 1 5 5~4 2 5 2 6", "4 5 5", 0}, // intersection

	{"⍝ Unique, union", "", 0},
	{"∪3", "3", 0},
	{"⍴∪3", "1", 0},
	{"∪ 22 10 22 22 21 10 5 10", "22 10 21 5", 0},
	{"∪2 7 1 8 2 8 1 8 2 8 4 5 9 0 4 4 9", "2 7 1 8 4 5 9 0", 0},
	{"∪'MISSISSIPPI'", "M I S P", 0},
	{"⍴∪⍳0", "0", 0},
	{"∪⍳0", "", 0},
	{"3∪3", "3", 0},
	{"⍴3∪3", "1", 0},
	{"3∪⍳0", "3", 0},
	{"(⍳0)∪3", "3", 0},
	{"⍴(⍳0)∪⍳0", "0", 0},
	{"1 2 3∪5 3 2 1 4", "1 2 3 5 4", 0},
	{"5 6 7∪1 2 3", "5 6 7 1 2 3", 0},

	{"⍝ Find", "", 0},
	{"'AN'⍷'BANANA'", "0 1 0 1 0 0", 0},
	{"'ANA'⍷'BANANA'", "0 1 0 1 0 0", 0},
	{"(2 2⍴1)⍷1 2 3", "0 0 0", 0},
	{"(2 2⍴5 6 8 9)⍷3 3⍴⍳9", "0 0 0\n0 1 0\n0 0 0", 0},
	{"4 5 6⍷3 3⍴⍳9", "0 0 0\n1 0 0\n0 0 0", 0},

	{"⍝ Magnitude, Residue, Ceil, Floor, Min, Max", "", 0},
	{"|1 ¯2 ¯3.2 2.2a20", "1 2 3.2 2.2", short | cmplx},      // magnitude
	{"3 3 ¯3 ¯3|¯5 5 ¯4 4", "1 2 ¯1 ¯2", 0},                  // residue
	{"0.5|3.12 ¯1 ¯0.6", "0.12 0 0.4", short},                // residue
	{"¯1 0 1|¯5.25 0 2.41", "¯0.25 0 0.41", short},           // residue
	{"1j2|2j3 3j4 5j6", "1J1 ¯1J1 0J1", cmplx},               // complex residue
	{"4J6|7J10", "3J4", cmplx},                               // complex residue
	{"¯10 7J10 .3|17 5 10", "¯3 ¯5J7 0.1", short | cmplx},    // residue
	{"⌊¯2.3 0.1 100 3.3", "¯3 0 100 3", 0},                   // floor
	{"⌊0.5 + 0.4 0.5 0.6", "0 1 1", 0},                       // floor
	{"⌊1j3.2 3.3j2.5 ¯3.3j¯2.5", "1J3 3J2 ¯3J¯3", cmplx},     // complex floor
	{"⌊1.5J2.5", "2J2", cmplx},                               // complex floor
	{"⌊1J2 1.2J2.5 ¯1.2J¯2.5", "1J2 1J2 ¯1J¯3", cmplx},       // complex floor
	{"⌈¯2.7 3 .5", "¯2 3 1", 0},                              // ceil
	{"⌈1.5J2.5", "1J3", cmplx},                               // complex ceil
	{"⌈1J2 1.2J2.5 ¯1.2J¯2.5", "1J2 1J3 ¯1J¯2", cmplx},       // complex ceil
	{"⌈¯2.3 0.1 100 3.3", "¯2 1 100 4", 0},                   // ceil
	{"⌈1.2j2.5 1.2j¯2.5", "1J3 1J¯2", cmplx},                 // ceil
	{"5⌊4 5 7", "4 5 5", 0},                                  // min
	{"¯2⌊¯3", "¯3", 0},                                       // min
	{"3.3 0 ¯6.7⌊3.1 ¯4 ¯5", "3.1 ¯4 ¯6.7", 0},               // min
	{"¯2.1 0.1 15.3 ⌊ ¯3.2 1 22", "¯3.2 0.1 15.3", 0},        // min
	{"5⌈4 5 7", "5 5 7", 0},                                  // max
	{"¯2⌈¯3", "¯2", 0},                                       // max
	{"3.3 0 ¯6.7⌈3.1 ¯4 ¯5", "3.3 0 ¯5", 0},                  // max
	{"¯2.01 0.1 15.3 ⌈ ¯3.2 ¯1.1 22.7", "¯2.01 0.1 22.7", 0}, // max

	{"⍝ Factorial, gamma, binomial.", "", 0},
	{"!4", "24", sfloat},                                         // factorial
	{"!1 2 3 4 5", "1 2 6 24 120", sfloat},                       // factorial
	{"!3J2", "¯3.0115J1.7702", cmplx | small},                    // complex gamma
	{"!.5 ¯.05", "0.88623 1.0315", short | small},                // real gamma (APL2 doc: "0.0735042656 1.031453317"?)
	{"2!5", "10", small},                                         // binomial
	{"3.2!5.2", "10.92", short | small},                          // binomial, floats with beta function
	{"3!¯2", "¯4", small},                                        // binomial, negative R
	{"¯6!¯3", "¯10", small},                                      // binomial negative L and R
	{"2 3 4!6 18 24", "15 816 10626", short | small},             // binomial
	{"3!.05 2.5 ¯3.6", "0.015437 0.3125 ¯15.456", short | small}, // binomial
	{"0 1 2 3!3", "1 3 3 1", small},                              // binomial coefficients
	{"2!3J2", "1J5", small | cmplx},                              // binomial complex

	{"⍝ Match, Not match, tally, depth", "", 0},
	{"≡5", "0", 0},          // depth
	{"≡⍳0", "1", 0},         // depth for empty array
	{`≡"alpha"`, "0", 0},    // a string is a scalarin APLv.
	{"≢2 3 4⍴⍳10", "2", 0},  // tally
	{"≢2", "1", 0},          // tally
	{"≢⍳0", "0", 0},         // tally
	{"1 2 3≡1 2 3", "1", 0}, // match
	{"3≡1⍴3", "0", 0},       // match shape
	{`""≡⍳0`, "0", 0},       // match empty string
	{`''≡⍳0`, "1", 0},       // this is false in other APLs (here '' is an empty array).
	{"2.0-1.0≡1>0", "1", 0}, // compare numbers of different type
	{"1≢2", "1", 0},         // not match
	{"1≢1", "0", 0},         // not match
	{"3≢1⍴3", "1", 0},       // not match
	{`""≢⍳0`, "1", 0},       // not match

	{"⍝ Left tack, right tack. ⊢ ⊣", "", 0},
	{"⊣1 2 3", "1 2 3", 0},      // monadic left: same
	{"3 2 1⊣1 2 3", "3 2 1", 0}, // dyadic left
	{"1 2 3⊢3 2 1", "3 2 1", 0}, // dyadic right
	{"⊢4", "4", 0},              // monadic right: same
	{"⊣/1 2 3", "1", 0},         // ⊣ reduction selects the first sub array
	{"⊢/1 2 3", "3", 0},         // ⊢ reduction selects the last sub array
	{"⊣/2 3⍴⍳6", "1 4", 0},      // ⊣ reduction over array
	{"⊢/2 3⍴⍳6", "3 6", 0},      // ⊢ reduction over array

	{"⍝ Array expressions.", "", 0},
	{"-⍳3", "¯1 ¯2 ¯3", 0},

	{"⍝ Ravel, enlist, catenate, join", "", 0},
	{",2 3⍴⍳6", "1 2 3 4 5 6", 0},     // ravel
	{"∊2 3⍴⍳6", "1 2 3 4 5 6", 0},     // enlist (identical for simple arrays)
	{"⍴,3", "1", 0},                   // scalar ravel
	{"⍴,⍳0", "0", 0},                  // ravel empty array
	{"1 2 3,4 5 6", "1 2 3 4 5 6", 0}, // catenate
	{`"abc",1 2`, `abc 1 2`, 0},
	{"(2 3⍴⍳6),2 2⍴7 8 9 10", "1 2 3 7 8\n4 5 6 9 10", 0},
	{"2 3≡2,3", "1", 0},                       // catenate vector result
	{"(1 2 3,4 5 6)≡⍳6", "1", 0},              // catenate vector result
	{"0,2 3⍴1", "0 1 1 1\n0 1 1 1", 0},        // catenate scalar and array
	{"0,[1]2 3⍴⍳6", "0 0 0\n1 2 3\n4 5 6", 0}, // catenate with axis
	{"(2 3⍴⍳6),[1]0", "1 2 3\n4 5 6\n0 0 0", 0},
	{"(2 3⍴⍳6),[1]5 4 3", "1 2 3\n4 5 6\n5 4 3", 0},
	{"⍴(3 5⍴⍳15),[1]3 3 5⍴-⍳45", "4 3 5", 0},
	{"⍴(3 5⍴⍳15),[2]3 3 5⍴-⍳45", "3 4 5", 0},

	{"⍝ Ravel with axis", "", 0},
	{",[0.5]1 2 3", "1 2 3", 0},
	{"⍴,[0.5]1 2 3", "1 3", 0},
	{",[1.5]1 2 3", "1\n2\n3", 0},
	{"⍴,[1.5]1 2 3", "3 1", 0},
	{"A←3 4⍴⍳12⋄⍴,[0.5]A", "1 3 4", 0},
	{"A←3 4⍴⍳12⋄⍴,[1.5]A", "3 1 4", 0},
	{"A←3 4⍴⍳12⋄⍴,[2.5]A", "3 4 1", 0},
	{"A←2 3⍴⍳6⋄⍴,[.1]A", "1 2 3", 0},
	{"A←2 3⍴⍳6⋄⍴,[1.1]A", "2 1 3", 0},
	{"A←2 3⍴⍳6⋄⍴,[2.1]A", "2 3 1", 0},
	{",[1.1]5 6 7", "5\n6\n7", 0},
	{"A←2 3 4⍴⍳24⋄A←,[1 2]A⋄⍴A⋄A[5;3]", "6 4\n19", 0},
	{"A←2 3 4⍴⍳24⋄⍴,[2 3]A", "2 12", 0},
	{"A←3 2 4⍴⍳24⋄⍴,[2 3]A", "3 8", 0},
	{"A←3 2 4⍴⍳24⋄⍴,[1 2]A", "6 4", 0},
	{"⍴,[⍳0]1 2 3", "3 1", 0},
	{"⍴,[⍳0]2 3⍴⍳6", "2 3 1", 0},
	{"A←3 2 5⍴⍳30⋄⍴,[⍳⍴⍴A],[.5]A", "6 5", 0}, // Turn array into matrix
	{"A←2 3 4⍴⍳24⋄(,[2 3]A)←2 12⍴-⍳24⋄⍴A⋄A[1;3;4]", "2 3 4\n¯12", 0},

	{"⍝ Laminate", "", 0},
	{"1 2 3,[0.5]4", "1 2 3\n4 4 4", 0},
	{"1 2 3,[1.5]4", "1 4\n2 4\n3 4", 0},
	{"⎕IO←0⋄1 2 3,[¯0.5]4", "1 2 3\n4 4 4", 0},
	{"'FOR',[.5]'AXE'", "F O R\nA X E", 0},
	{"'FOR',[1.1]'AXE'", "F A\nO X\nR E", 0},

	{"⍝ Table, catenate first", "", 0},
	{"⍪0", "0", 0},
	{"⍴⍪0", "1 1", 0},
	{"⍪⍳4", "1\n2\n3\n4", 0},
	{"⍪2 2⍴⍳4", "1 2\n3 4", 0},
	{"⍪2 2 2⍴⍳8", "1 2 3 4\n5 6 7 8", 0},
	{"10 20⍪2 2⍴⍳4", "10 20\n1 2\n3 4", 0},

	{"⍝ Decode", "", 0},
	{"3⊥1 2 1", "16", 0},
	{"3⊥4 3 2 1", "142", 0},
	{"2⊥1 1 1 1", "15", 0},
	{"1 2 3⊥3 2 1", "25", 0},
	{"1J1⊥1 2 3 4", "5J9", cmplx},
	{"24 60 60⊥2 23 12", "8592", 0},
	{"(2 1⍴2 10)⊥3 2⍴ 1 4 0 3 1 2", "5 24\n101 432", 0},

	{"⍝ Encode, representation", "", 0},
	{"2 2 2 2⊤15", "1 1 1 1", 0},
	{"10⊤5 15 125", "5 5 5", 0},
	{"⍴10⊤5 15 125", "3", 0},
	{"⍴(1 1⍴10)⊤5 15 125", "1 1 3", 0},
	{"0 10⊤5 15 125", "0 1 12\n5 5 5", 0},
	{"0 1⊤1.25 10.5", "1 10\n0.25 0.5", 0},
	{"24 60 60⊤8592", "2 23 12", 0},
	{"2 2 2 2 2⊤15", "0 1 1 1 1", 0},
	{"2 2 2⊤15", "1 1 1", 0},
	{"4 5 6⊤⍳0", "", 0},
	{"⍴4 5 6⊤⍳0", "3 0", 0},
	{"⍴(⍳0)⊤4 5 6", "0 3", 0},
	{"((⌊1+2⍟135)⍴2)⊤135", "1 0 0 0 0 1 1 1", float},
	{"24 60 60⊤162507", "21 8 27", 0},
	{"0 24 60 60⊤162507", "1 21 8 27", 0},
	{"10 10 10⊤215 345 7", "2 3 0\n1 4 0\n5 5 7", 0},
	{"(4 2⍴8 2)⊤15", "0 1\n0 1\n1 1\n7 1", 0},
	{"3 2J3⊤2", "0J2 ¯1J2", cmplx},
	{"0 2J3⊤2", "0J¯1 ¯1J2", cmplx},
	{"3 2J3⊤2", "0J2 ¯1J2", cmplx},
	{"3 2J3⊤2 1", "0J2 0J2\n¯1J2 ¯2J2", cmplx},
	{"10⊥2 2 2 2⊤15", "1111", 0},
	{"10 10 10⊤123", "1 2 3", 0},
	{"10 10 10⊤123 456", "1 4\n2 5\n3 6", 0},
	{"2 2 2⊤¯1", "1 1 1", 0},
	{"0 2 2⊤¯1", "¯1 1 1", 0},
	{"0 1⊤3.75 ¯3.75", "3 ¯4\n0.75 0.25", 0},
	{"1 0⊤0", "0 0", 0},
	{"0⊤0", "0", 0},
	{"0⊤0 0", "0 0", 0},
	{"0 0⊤0", "0 0", 0},
	{"1 0⊤234", "0 234", 0},

	{"⍝ Reduce, reduce first, reduce with axis", "", 0},
	{"+/1 2 3", "6", 0},
	{"+⌿1 2 3", "6", 0},
	{"+/2 3 1 ⍴⍳6", "1 2 3\n4 5 6", 0},
	{"⍴+/3", "", 0},
	{"⍴+/1 1⍴3", "1", 0},
	{"+/2 3⍴⍳6", "6 15", 0},
	{"+⌿2 3⍴⍳6", "5 7 9", 0},
	{"+/⍳0", "0", 0},
	{"+/[1]2 3⍴⍳6", "5 7 9", 0},
	{"+/[1]3 4⍴⍳12", "15 18 21 24", 0},
	{"+/[2]3 4⍴⍳12", "10 26 42", 0},
	{"×/[1]3 4 ⍴⍳12", "45 120 231 384", 0},
	{"÷/[2]2 1 4⍴2×⍳8", "2 4 6 8\n10 12 14 16", 0},
	{"÷/[2]2 0 3⍴0", "1 1 1\n1 1 1", 0},

	{"⍝ N-wise reduction", "", 0},
	{"6+/⍳6", "21", 0},
	{"4+/⍳6", "10 14 18", 0},
	{"5+/⍳6", "15 20", 0},
	{"3+/⍳6", "6 9 12 15", 0},
	{"1+/⍳6", "1 2 3 4 5 6", 0},
	{"0+/⍳0", "0", 0},
	{"⍴0+/⍳0", "1", 0},
	{"1+/⍳0", "", 0},
	{"¯1+/⍳0", "", 0},
	{"⍴4+/2 3⍴⍳6", "2 0", 0},
	{"2+/3 4⍴⍳12", "3 5 7\n11 13 15\n19 21 23", 0},
	{"¯2-/1 4 9 16 25", "3 5 7 9", 0},
	{"2-/1 4 9 16 25", "¯3 ¯5 ¯7 ¯9", 0},
	{"3×/⍳6", "6 24 60 120", 0},
	{"¯3×/⍳6", "6 24 60 120", 0},
	{"0×/⍳5", "1 1 1 1 1 1", 0},
	{"4+/[1]4 3⍴⍳12", "22 26 30", 0},
	{"3+/[1]4 3⍴⍳12", "12 15 18\n21 24 27", 0},
	{"2+/[1]4 3⍴⍳12", "5 7 9\n11 13 15\n17 19 21", 0},
	{"0×/[1]2 3⍴⍳12", "1 1 1\n1 1 1\n1 1 1", 0},
	{"1+/⍳6", "1 2 3 4 5 6", 0},
	{`S←0.0 n→f "%.0f" ⋄ +/1000+/⍳10000`, "45009500500", 0},

	{"⍝ Scan, scan first, scan with axis.", "", 0},
	{`+\1 2 3 4 5`, "1 3 6 10 15", 0},
	{`+\2 3⍴⍳6`, "1 3 6\n4 9 15", 0},
	{`+⍀2 3⍴⍳6`, "1 2 3\n5 7 9", 0},
	{`-\1 2 3`, "1 ¯1 2", 0},
	{"∨/0 0 1 0 0 1 0", "1", 0},
	{`^\1 1 1 0 1 1 1`, "1 1 1 0 0 0 0", 0},
	{`+\1 2 3 4 5`, "1 3 6 10 15", 0},
	{`+\[1]2 3⍴⍳6`, "1 2 3\n5 7 9", 0},

	{"⍝ Replicate, compress", "", 0},
	{"1 1 0 0 1/'STRAY'", "S T Y", 0},
	{"1 0 1 0/3 4⍴⍳12", "1 3\n5 7\n9 11", 0},
	{"1 0 1/1 2 3", "1 3", 0},
	{"1/1 2 3", "1 2 3", 0},
	{"3 2 1/1 2 3", "1 1 1 2 2 3", 0},
	{"1 0 1/2", "2 2", 0},
	{"⍴1/1", "1", 0},
	{"⍴⍴(,1)/2", "1", 0},
	{"3 4/1 2", "1 1 1 2 2 2 2", 0},
	{"1 0 1 0 1/⍳5", "1 3 5", 0},
	{"1 ¯2 3 ¯4 5/⍳5", "1 0 0 3 3 3 0 0 0 0 5 5 5 5 5", 0},
	{"2 0 1/2 3⍴⍳6", "1 1 3\n4 4 6", 0},
	{"0 1⌿2 3⍴⍳6", "4 5 6", 0},
	{"0 1⌿⍴⍳6", "6", 0},
	{"1 0 1/4", "4 4", 0},
	{"1 0 1/,3", "3 3", 0},
	{"1 0 1/1 1⍴5", "5 5", 0},
	{"1 2/[2]2 2 1⍴⍳4", "1\n2\n2\n\n3\n4\n4", 0},
	{"A←2 ¯1 1/[1]3 2 4⍴⍳24⋄⍴A⋄+/+/A", "4 2 4\n36 36 0 164", 0},
	{"⍴2/[2]3 2 4⍴⍳24", "3 4 4", 0},
	{"⍴¯1 1/[2]3 1 4⍴⍳12", "3 2 4", 0},
	{"⍴1 0 2 ¯1⌿[2]3 4⍴⍳12", "3 4", 0},
	{"0 1/[1]2 3⍴⍳6", "4 5 6", 0},
	{"B←2 2⍴'ABCD'⋄A←3 2⍴⍳6⋄(1 0 1/[1]A)←B⋄A", "A B\n3 4\nC D", 0},

	{"⍝ Expand, expand first", "", 0},
	{`1 0 1 0 0 1\1 2 3`, "1 0 2 0 0 3", 0},
	{`1 0 0\5`, "5 0 0", 0},
	{`0 1 0\3 1⍴7 8 9`, "0 7 0\n0 8 0\n0 9 0", 0},
	{`1 0 0 1 0 1\7 8 9`, "7 0 0 8 0 9", 0},
	{`⍴(⍳0)\3`, "0", 0},
	{`⍴(⍳0)\2 0⍴3`, "2 0", 0},
	{`⍴1 0 1\0 2⍴0`, "0 3", 0},
	{`0 0 0\2 0⍴0`, "0 0 0\n0 0 0", 0},
	{`1 0 1⍀2 3⍴⍳6`, "1 2 3\n0 0 0\n4 5 6", 0},
	{`0\⍳0`, "0", 0},
	{`1 ¯2 3 ¯4 5\3`, "3 0 0 3 3 3 0 0 0 0 3 3 3 3 3", 0},
	{`1 0 1\1 3`, "1 0 3", 0},
	{`1 0 1\2`, "2 0 2", 0},
	{`1 0 1 1\1 2 3`, "1 0 2 3", 0},
	{`1 0 1 1⍀3`, "3 0 3 3", 0},
	{`0 1\3 1⍴3 2 4`, "0 3\n0 2\n0 4", 0},
	{`0 0\5`, "0 0", 0},
	{`1 0 1⍀2 3⍴⍳6`, "1 2 3\n0 0 0\n4 5 6", 0},
	{`1 0 1\3 2⍴⍳6`, "1 0 2\n3 0 4\n5 0 6", 0},
	{`1 0 1 1\2 3⍴⍳6`, "1 0 2 3\n4 0 5 6", 0},
	{`1 0 1\[1]2 3⍴⍳6`, "1 2 3\n0 0 0\n4 5 6", 0},
	{"⍝ TODO expand with selective specification", "", 0},

	{"⍝ Pi times, circular, trigonometric", "", 0},
	{"○0 1 2", "0 3.1416 6.2832", short | small},            // pi times
	{"*○0J1", "¯1.00J0.00", cmplxf | small},                 // Euler identity
	{"0 ¯1 ○ 1", "0 1.5708", short | small},                 //
	{"1○(○1)÷2 3 4", "1 0.86603 0.70711", short | small},    //
	{"2○(○1)÷3", "0.5", short | small},                      //
	{"9 11○3.5J¯1.2", "3.5 ¯1.2", small},                    //
	{"9 11∘.○3.5J¯1.2 2J3 3J4", "3.5 2 3\n¯1.2 3 4", small}, //
	{"¯4○¯1", "0", small},                                   //
	{"3○2", "¯2.185", short | small},                        //
	{"2○1", "0.5403", short | small},                        //
	{"÷3○2", "¯0.45766", short | small},                     //
	{"1○○30÷180", "0.5", short | small},
	{"2○○45÷180", "0.70711", short | small},
	{"¯1○1", "1.5708", short | small},
	{"¯2○.54032023059", "0.99998", short | small},
	{"(¯1○.5)×180÷○1", "30", short | small},
	{"(¯3○1)×180÷○1", "45", short | small},
	{"5○1", "1.1752", short | small},
	{"6○1", "1.5431", short | small},
	{"¯5○1.175201194", "1", short | small},
	{"¯6○1.543080635", "1", short | small},

	{"⍝ Take, drop", "", 0}, // Monadic First and split are not implemented.
	{"5↑'ABCDEF'", "A B C D E", 0},
	{"5↑1 2 3", "1 2 3 0 0", 0},
	{"¯5↑1 2 3", "0 0 1 2 3", 0},
	{"2 3↑2 4⍴⍳8", "1 2 3\n5 6 7", 0},
	{"¯1 ¯2↑2 4⍴⍳8", "7 8", 0},
	{"1↑2", "2", 0},
	{"⍴1↑2", "1", 0},
	{"1 1 1↑2", "2", 0},
	{"⍴1 1 1↑2", "1 1 1", 0},
	{"(⍳0)↑2", "2", 0},
	{"⍴(⍳0)↑2", "", 0},
	{"2↑⍳0", "0 0", 0},
	{"2 3↑2", "2 0 0\n0 0 0", 0},
	{"4↓'OVERBOARD'", "B O A R D", 0},
	{"¯5↓'OVERBOARD'", "O V E R", 0},
	{"⍴10↓'OVERBOARD'", "0", 0},
	{"0 ¯2↓3 3⍴⍳9", "1\n4\n7", 0},
	{"¯2 ¯1↓3 3⍴⍳9", "1 2", 0},
	{"1↓3 3⍴⍳9", "4 5 6\n7 8 9", 0},
	{"1 1↓2 3 4⍴⍳24", "17 18 19 20\n21 22 23 24", 0},
	{"¯1 ¯1↓2 3 4⍴⍳24", "1 2 3 4\n5 6 7 8", 0},
	{"3↓12 31 45 10 57", "10 57", 0},
	{"¯3↓12 31 45 10 57", "12 31", 0},
	{"0 2↓3 5⍴⍳15", "3 4 5\n8 9 10\n13 14 15", 0},
	{"⍴3 1↓2 3⍴'ABCDEF'", "0 2", 0},
	{"⍴2 3↓2 3⍴'ABCDEF'", "0 0", 0},
	{"0↓4", "4", 0},
	{"⍴0↓4", "1", 0},
	{"0 0 0↓4", "4", 0},
	{"⍴0 0 0↓4", "1 1 1", 0},
	{"⍴1↓5", "0", 0},
	{"⍴0↓5", "1", 0},
	{"⍴1 2 3↓4", "0 0 0", 0},
	{"''↓5", "5", 0},
	{"⍴⍴''↓5", "0", 0},
	{"1↑2 3⍴⍳6", "1 2 3", 0},
	{"1↑[1]2 3⍴⍳6", "1 2 3", 0},
	{"1 3↑[1 2]2 3⍴⍳6", "1 2 3", 0},
	{"2↑[1]3 5⍴'GIANTSTORETRAIL'", "G I A N T\nS T O R E", 0},
	{"¯3↑[2]3 5⍴'GIANTSTORETRAIL'", "A N T\nO R E\nA I L", 0},
	{"3↑[1]2 3⍴⍳6", "1 2 3\n4 5 6\n0 0 0", 0},
	{"¯4↑[1]2 3⍴⍳6", "0 0 0\n0 0 0\n1 2 3\n4 5 6", 0},
	{"¯1 3↑[1 3]3 3 4⍴'HEROSHEDDIMESODABOARPARTLAMBTOTODAMP'", "L A M\nT O T\nD A M", 0},
	{"2↑[2]2 3 4⍴⍳24", "1 2 3 4\n5 6 7 8\n\n13 14 15 16\n17 18 19 20", 0},
	{"2↑[3]2 3 4⍴⍳24", "1 2\n5 6\n9 10\n\n13 14\n17 18\n21 22", 0},
	{"2 ¯2↑[3 2]2 3 4⍴⍳24", "5 6\n9 10\n\n17 18\n21 22", 0},
	{"2 ¯2↑[2 3]2 3 4⍴⍳24", "3 4\n7 8\n\n15 16\n19 20", 0},
	{"1↓[1]3 4⍴'FOLDBEATRODE'", "B E A T\nR O D E", 0},
	{"1↓[2]3 4⍴'FOLDBEATRODE'", "O L D\nE A T\nO D E", 0},
	{"A←3 4⍴'FOLDBEATRODE'⋄(1↓[1]A)≡1 0↓A", "1", 0},
	{"A←3 4⍴'FOLDBEATRODE'⋄(1↓[2]A)≡0 1↓A", "1", 0},
	{"A←3 2 4⍴⍳24⋄1 ¯1↓[2 3]A", "5 6 7\n\n13 14 15\n\n21 22 23", 0},
	{"A←3 2 4⍴⍳24⋄1 ¯1↓[3 2]A", "2 3 4\n\n10 11 12\n\n18 19 20", 0},
	{"A←2 3 4⍴⍳24⋄⍴1↓[2]A", "2 2 4", 0},
	{"A←2 3 4⍴⍳24⋄2↓[3]A", "3 4\n7 8\n11 12\n\n15 16\n19 20\n23 24", 0},
	{"A←2 3 4⍴⍳24⋄2 1↓[3 2]A", "7 8\n11 12\n\n19 20\n23 24", 0},

	{"⍝ Format as a string, Execute", "", 0},
	{"⍕10", "10", 0},   // format as string
	{`⍎"1+1"`, "2", 0}, // evaluate expression
	{"⍝ TODO: dyadic format with specification.", "", 0},
	{"⍝ TODO: dyadic execute with namespace.", "", 0},

	{"⍝ Grade up, grade down, sort.", "", 0},
	{"⍋23 11 13 31 12", "2 5 3 1 4", 0},                             // grade up
	{"⍋23 14 23 12 14", "4 2 5 1 3", 0},                             // identical subarrays
	{"⍋5 3⍴4 16 37 2 9 26 5 11 63 3 18 45 5 11 54", "2 4 1 5 3", 0}, // grade up rank 2
	{"⍋22.5 1 15 3 ¯4", "5 2 4 3 1", 0},                             // grade up
	{"⍒33 11 44 66 22", "4 3 1 5 2", 0},                             // grade down
	{"⍋'alpha'", "1 5 4 2 3", 0},                                    // strings grade up
	{"'ABCDE'⍒'BEAD'", "2 4 1 3", 0},                                // grade down with collating sequence
	{"⍝ TODO dyadic grade up/down is only implemented for vector L", "", 0},
	{"A←23 11 13 31 12⋄A[⍋A]", "11 12 13 23 31", 0}, // sort

	{"⍝ Reverse, revere first", "", 0},
	{"⌽1 2 3 4 5", "5 4 3 2 1", 0}, // reverse vector
	{"⌽2 3⍴⍳6", "3 2 1\n6 5 4", 0}, // reverse matrix
	{"⊖2 3⍴⍳6", "4 5 6\n1 2 3", 0}, // reverse first
	{"⌽[1]2 3⍴⍳6", "4 5 6\n1 2 3", 0},
	{"⊖[2]2 3⍴⍳6", "3 2 1\n6 5 4", 0},
	{"A←2 3⍴⍳12 ⋄ (⌽[1]A)←2 3⍴-⍳6⋄A", "¯4 ¯5 ¯6\n¯1 ¯2 ¯3", 0},

	{"⌽'DESSERTS'", "S T R E S S E D", 0}, // reverse strings
	{"⍝ Rotate", "", 0},
	{"1⌽1 2 3 4", "2 3 4 1", 0},                                                     // rotate vector
	{"10⌽1 2 3 4", "3 4 1 2", 0},                                                    // rotate vector
	{"¯1⌽1 2 3 4", "4 1 2 3", 0},                                                    // rotate vector negative
	{"(-7)⌽1 2 3 4", "2 3 4 1", 0},                                                  // rotate vector negative
	{"1 2⌽2 3⍴⍳6", "2 3 1\n6 4 5", 0},                                               // rotate array
	{"(2 2⍴2 ¯3 3 ¯2)⌽2 2 4⍴⍳16", "3 4 1 2\n6 7 8 5\n\n12 9 10 11\n15 16 13 14", 0}, // rotate array
	{"(2 3⍴2 ¯3 3 ¯2 1 2)⊖2 2 3⍴⍳12", "1 8 9\n4 11 6\n\n7 2 3\n10 5 12", 0},         // rotate array
	{"(2 4⍴0 1 ¯1 0 0 3 2 1)⌽[2]2 2 4⍴⍳16", "1 6 7 4\n5 2 3 8\n\n9 14 11 16\n13 10 15 12", 0},
	{"A←3 4⍴⍳12⋄(1 ¯1 2 ¯2⌽[1]A)←3 4⍴'ABCDEFGHIJKL'⋄A", "I F G L\nA J K D\nE B C H", 0},

	{"⍝ Transpose", "", 0},
	{"1 2 1⍉2 3 4⍴⍳6", "1 5 3\n2 6 4", 0},
	{"⍉3 1⍴1 2 3", "1 2 3", 0},
	{"⍴⍉2 3⍴⍳6", "3 2", 0},
	{"+/+/1 3 2⍉2 3 4⍴⍳24", "78 222", 0},
	{"+/+/3 2 1⍉2 3 4⍴⍳24", "66 72 78 84", 0},
	{"+/+/2 1 3⍉2 3 4⍴⍳24", "68 100 132", 0},
	{"1 1 1⍉2 3 3⍴⍳18", "1 14", 0},
	{"1 1 1⍉2 3 4⍴'ABCDEFGHIJKL',⍳12", "A 6", 0},
	{"1 1 2⍉2 3 4⍴'ABCDEFGHIJKL',⍳12", "A B C D\n5 6 7 8", 0},
	{"2 2 1⍉2 3 4⍴'ABCDEFGHIJKL',⍳12", "A 5\nB 6\nC 7\nD 8", 0},
	{"1 2 2⍉2 3 4⍴'ABCDEFGHIJKL',⍳12", "A F K\n1 6 11", 0},
	{"1 2 1⍉2 3 4⍴'ABCDEFGHIJKL',⍳12", "A E I\n2 6 10", 0},
	{"⍴⍴(⍳0)⍉5", "0", 0},
	{"⍴2 1 3⍉3 2 4⍴⍳24", "2 3 4", 0},
	{"⎕IO←0⋄⍴1 0 2⍉3 2 4⍴⍳24", "2 3 4", 0},
	{"A←3 3⍴⍳9⋄(1 1⍉A)←10 20 30⋄A", "10 2 3\n4 20 6\n7 8 30", 0},

	{"⍝ Enclose, string catenation, join strings, newline", "", 0},
	{`⊂'alpha'`, "alpha", 0},
	{`"+"⊂'alpha'`, "a+l+p+h+a", 0},
	{`⎕NL⊂"alpha" "beta" "gamma"`, "alpha\nbeta\ngamma", 0},
	{"`alpha`beta`gamma", "alpha beta gamma", 0},
	{"(`alpha`beta`gamma)", "alpha beta gamma", 0},
	{"`alpha`beta`gamma⋄", "alpha beta gamma", 0},

	{"⍝ Domino, solve linear system", "", 0},
	{"⌹2 2⍴2 0 0 1", "0.5 0\n0 1", short},
	// TODO: this fails for big.Float. Remove sfloat and debug
	{"(1 ¯2 0)⌹3 3⍴3 2 ¯1 2 ¯2 4 ¯1 .5 ¯1", "1\n¯2\n¯2", short | sfloat},
	// A←2a30
	// B←1a10
	// RHS←A+B**(¯1+⍳6)×○1÷3
	// S←⍉2 6⍴(6⍴1),*0J1×(¯1+⍳6)×○1÷3
	// ⍉RHS⌹S
	// With rational numbers:
	// A←3 3⍴9?100
	// B←3 3⍴9?100
	// 0=⌈/⌈/|B-A+.×B⌹A

	{"⍝ Dates, Times and durations", "", small},
	{"2018.12.23", "2018.12.23T00.00.00.000", small},       // Parse a time
	{"2018.12.23+12s", "2018.12.23T00.00.12.000", small},   // Add a duration to a time
	{"2018.12.24<2018.12.23", "0", small},                  // Times are comparable
	{"⌊/3s 2s 10s 4s", "2s", small},                        // Durations are comparable
	{"2018.12.23-1s", "2018.12.22T23.59.59.000", small},    // Substract a duration from a time
	{"2017.03.01-2017.02.28", "24h0m0s", small},            // Substract two times returns a duration
	{"2016.03.01-2016.02.28", "48h0m0s", small},            // Leap years are covered
	{"3m-62s", "1m58s", small},                             // Substract two durations
	{"-3s", "-3s", small},                                  // Negate a duration
	{"×¯3h 0s 2m 2015.01.02", "¯1 0 1 1", small},           // Signum
	{"(|¯1s)+|1s", "2s", small},                            // Absolute value of a duration
	{"3×1h", "3h0m0s", small},                              // Uptype numbers to seconds and multiply durations
	{"1m × ⍳5", "1m0s 2m0s 3m0s 4m0s 5m0s", small},         // Generate a duration vector
	{"⍴⍪2018.12.23 + 1h×(¯1+⍳24)", "24 1", small},          // Table with all starting hours in a day
	{"4m×42.195", "2h48m46.8s", small},                     //
	{"⌈2018.12.23+3.5s", "2018.12.23T00.00.04.000", small}, // Ceil rounds to seconds
	{"⌊3h÷42.195", "4m15s", small},                         // Floor truncates seconds.

	{"⍝ Basic operators.", "", 0},
	{"+/1 2 3", "6", 0},                            // plus reduce
	{"1 2 3 +.× 4 3 2", "16", 0},                   // scalar product
	{"(2 3⍴⍳6) +.× 3 2⍴5+⍳6", "52 58\n124 139", 0}, // matrix multiplication
	{`-\×\+\1 2 3`, "1 ¯2 16", 0},                  // chained monadic operators
	{"+/+/+/+/1 2 3", "6", 0},
	{`+.×/2 3 4`, "24", 0},
	{`S←0.0 n→f "%.0f"⋄ +.×.*/2 3 4`, "2417851639229258349412352", 0},
	{`+.*.×/2 3 4`, "24", 0},

	{"⍝ Identify item for reduction over empty array", "", 0},
	{"+/⍳0", "0", 0},
	{"-/⍳0", "0", 0},
	{"×/⍳0", "1", 0},
	{"÷/⍳0", "1", 0},
	{"|/⍳0", "0", 0},
	{"⌊/⍳0", fmt.Sprintf("¯%v", float64(math.MaxFloat64)), 0},
	{"⌈/⍳0", fmt.Sprintf("%v", float64(math.MaxFloat64)), 0},
	{"*/⍳0", "1", 0},
	{"!/⍳0", "1", 0},
	{"^/⍳0", "1", 0},
	{"∧/⍳0", "1", 0},
	{"∨/⍳0", "0", 0},
	{"</⍳0", "0", 0},
	{"≤/⍳0", "1", 0},
	{"=/⍳0", "1", 0},
	{"≥/⍳0", "1", 0},
	{">/⍳0", "0", 0},
	{"≠/⍳0", "0", 0},
	{"⊤/⍳0", "0", 0},
	{"⌽/⍳0", "0", 0},
	{"⊖/⍳0", "0", 0},
	{"∨/0 3⍴ 1", "", 0},
	{"∨/3 3⍴ ⍳0", "0 0 0", 0},
	{"∪/⍳0", "0", 0},
	// These are implemented as operators and do not parse.
	// {"//⍳0", "0", 0},
	// {"⌿/⍳0", "0", 0},
	// {`\/⍳0`, "0", 0},
	// {`⍀/⍳0`, "0", 0},

	{"⍝ Outer product", "", 0},
	{"10 20 30∘.+1 2 3", "11 12 13\n21 22 23\n31 32 33", 0},
	{"(⍳3)∘.=⍳3", "1 0 0\n0 1 0\n0 0 1", 0},
	{"1 2 3∘.×4 5 6", "4 5 6\n8 10 12\n12 15 18", 0},

	{"⍝ Each", "", 0},
	{"-¨1 2 3", "¯1 ¯2 ¯3", 0},   // monadic each
	{"1+¨1 2 3", "2 3 4", 0},     // dyadic each
	{"1 2 3+¨1", "2 3 4", 0},     // dyadic each
	{"1 2 3+¨4 5 6", "5 7 9", 0}, // dyadic each
	{"1+¨1", "2", 0},             // dyadic each

	{"⍝ Commute, duplicate", "", 0},
	{"∘.≤⍨1 2 3", "1 1 1\n0 1 1\n0 0 1", 0},
	{"+/∘(÷∘⍴⍨)⍳10", "5.5", 0}, // mean value
	{"⍴⍨3", "3 3 3", 0},
	{"3-⍨4", "1", 0},
	{"+/2*⍨2 2⍴4 7 1 8", "65 65", 0},
	{"3-⍨4", "1", 0},

	{"⍝ Composition", "", 0},
	{"+/∘⍳¨2 4 6", "3 10 21", 0}, // Form I
	{"1∘○ 10 20 30", "¯0.54402 0.91295 ¯0.98803", short | small},
	{"+∘÷/40⍴1", "1.618", short},       // Form IV, golden ratio (continuous-fraction)
	{"(*∘0.5)4 16 25", "2 4 5", float}, // Form III

	{"⍝ Power operator", "", 0},
	{"⍟⍣2 +2 3 4", "¯0.36651 0.094048 0.32663", short | float}, // log log
	// TODO: 1+∘÷⍣=1 oscillates for big.Float.
	// TODO: Add comparison tolerance and remove sfloat.
	{"1+∘÷⍣=1", "1.618", short | small}, // fixed point iteration golden ratio
	{"⍝ TODO: function inverse", "", 0},

	{"⍝ Rank operator", "", 0},
	{`+\⍤0 +2 3⍴1`, "1 1 1\n1 1 1", 0},
	{`+\⍤1 +2 3⍴1`, "1 2 3\n1 2 3", 0},
	{"⍴⍤1 +2 3⍴1", "3\n3", 0},
	{"⍴⍤2 +2 3 5⍴1", "3 5\n3 5", 0},
	{"4 5+⍤1 0 2 +2 2⍴7 8 9 10", "11 12\n13 14\n\n12 13\n14 15", 0},
	{"⍉2 2 2⊤⍤1 0 ⍳5", "0 0 0 1 1\n0 1 1 0 0\n1 0 1 0 1", 0},
	{"⍳⍤1 +3 1⍴⍳3", "1 0 0\n1 2 0\n1 2 3", 0},

	{"⍝ At", "", 0},
	{"(10 20@2 4)⍳5", "1 10 3 20 5", 0},
	{"10 20@2 4⍳5", "1 10 3 20 5", 0},
	{"((2 3⍴10 20)@2 4)4 3⍴⍳12", "1 2 3\n10 20 10\n7 8 9\n20 10 20", 0},
	{"⍴@(0.5∘<)3 3⍴1 4 0.2 0.3 0.3 4", "5 5 0.2\n0.3 0.3 5\n5 5 0.2", 0},
	{"÷@2 4 ⍳5", "1 0.5 3 0.25 5", 0},
	{"⌽@2 4 ⍳5", "1 4 3 2 5", 0},
	{"10×@2 4⍳5", "1 20 3 40 5", 0},
	{`(+\@2 4)4 3⍴⍳12`, "1 2 3\n4 9 15\n7 8 9\n10 21 33", 0},
	{"0@(2∘|)⍳5", "0 2 0 4 0", 0},
	{"÷@(2∘|)⍳5", "1 2 0.33333 4 0.2", short},
	{"⌽@(2∘|)⍳5", "5 2 3 4 1", 0},

	{"⍝ Stencil", "", 0},
	{"{⌈/⌈/⍵}⌺(3 3) ⊢3 3⍴⍳25", "5 6 6\n8 9 9\n8 9 9", 0},

	{"⍝ Assignment, specification", "", 0},
	{"X←3", "", 0},          // assign a number
	{"-X←3", "¯3", 0},       // assign a value and use it
	{"X←3⋄X←4", "", 0},      // assign and overwrite
	{"X←3⋄⎕←X", "3", 0},     // assign and check
	{"f←+", "", 0},          // assign a function
	{"f←+⋄⎕←3 f 3", "6", 0}, // assign a function and apply
	{"X←4⋄⎕←÷X", "0.25", 0}, // assign and use it in another expr
	{"A←2 3 ⋄ A", "2 3", 0}, // assign a vector

	{"⍝ Indexed assignment", "", 0},
	{"A←2 3 4 ⋄ A[1]←1 ⋄ A", "1 3 4", 0},
	{"A←2 2⍴⍳4 ⋄ +A[1;1]←3 ⋄ A", "3\n3 2\n3 4", 0},
	{"A←⍳5 ⋄ A[2 3]←10 ⋄ A", "1 10 10 4 5", 0},
	{"A←2 3⍴⍳6 ⋄ A[;2 3]←2 2⍴⍳4 ⋄ A", "1 1 2\n4 3 4", 0},

	{"⍝ TODO: choose/reach indexed assignment", "", 0},

	{"⍝ Multiple assignment", "", 0},
	{"A←B←C←D←1 ⋄ A B C D", "1 1 1 1", 0},

	{"⍝ Vector assignment", "", 0},
	{"(A B C)←2 3 4 ⋄ A ⋄ B ⋄ C ", "2\n3\n4", 0},
	{"-A B C←1 2 3 ⋄ A B C", "¯1 ¯2 ¯3\n1 2 3", 0},

	{"⍝ Modified assignment", "", 0},
	{"A←1 ⋄ A+←1 ⋄ A", "2", 0},
	{"A←1 2⋄ A+←1 ⋄ A", "2 3", 0},
	{"A←1 2 ⋄ A+←3 4 ⋄ A", "4 6", 0},
	{"A←1 2 ⋄ A{⍺+⍵}←3 ⋄ A", "4 5", 0},
	{"A B C←1 2 3 ⋄ A B C +← 4 5 6 ⋄ A B C", "5 7 9", 0},

	// Selective specification APL2 p.41, DyaRef p.21
	{"⍝ Selective assignment/specification", "", 0},
	{"A←10 20 30 40 ⋄ (2↑A)←100 200 ⋄ A", "100 200 30 40", 0},
	{"A←'ABCD' ⋄ (3↑A)←1 2 3 ⋄ A", "1 2 3 D", 0},
	{"A←1 2 3 ⋄ ((⍳0)↑A)←4 ⋄ A", "4 4 4", 0},
	//{"A←1 2 3 ⋄ (4↑A)←4 ⋄ A", "4 4 4", 0}, // overtake is ignored
	{"A←2 3⍴⍳6 ⋄ (,A)←2×⍳6 ⋄ A", "2 4 6\n8 10 12", 0},
	{"A←3 4⍴⍳12 ⋄ (4↑,⍉A)←10 20 30 40 ⋄ ,A ", "10 40 3 4 20 6 7 8 30 10 11 12", 0},
	{"A←2 3⍴'ABCDEF' ⋄ A[1;1 3]←8 9 ⋄ A", "8 B 9\nD E F", 0},
	{"A←2 3 4 ⋄ A[]←9 ⋄ A", "9 9 9", 0},
	{"A←4 3⍴⍳12 ⋄ (1 0 0/A)←1 4⍴⍳4 ⋄ A[3;1]", "3", 0}, // single element axis are collapsed
	{"A←3 2⍴⍳6 ⋄ (1 0/A)←'ABC' ⋄ A", "A 2\nB 4\nC 6", 0},
	{"A←4 5 6 ⋄ (1 ¯1  1/A)←7 8 9 ⋄ A", "7 5 9", 0},
	{"A←3 2⍴⍳6 ⋄ B←2 2⍴'ABCD' ⋄ (1 0 1/[1]A)←B ⋄ A", "A B\n3 4\nC D", 0},
	{"A←5 6 7 8 9 ⋄ (2↓A)←⍳3 ⋄ A", "5 6 1 2 3", 0},
	{"A←3 4⍴'ABCDEFGHIJKL' ⋄ (1 ¯1↓A)←2 3⍴⍳6 ⋄ A", "A B C D\n1 2 3 H\n4 5 6 L", 0},
	{"A←2 3⍴⍳6 ⋄ (1↓[1]A)←9 8 7 ⋄ A", "1 2 3\n9 8 7", 0},
	{"A←2 3 4⍴⍳12⋄(¯1 2↓[3 2]A)←0⋄A", "1 2 3 4\n5 6 7 8\n0 0 0 12\n\n1 2 3 4\n5 6 7 8\n0 0 0 12", 0},
	{`A←'ABC' ⋄ (1 0 1 0 1\A)←⍳5 ⋄ A`, "1 3 5", 0},
	{`A←2 3⍴⍳6 ⋄ (1 0 1 1\A)←10×2 4⍴⍳8 ⋄ A`, "10 30 40\n50 70 80", 0},
	{`A←3 2⍴⍳6 ⋄ (1 1 0 0 1\[1]A)←5 2⍴-⍳10 ⋄ A`, "¯1 ¯2\n¯3 ¯4\n¯9 ¯10", 0},
	{"A←2 3⍴⍳6 ⋄ (,A)←10×⍳6 ⋄ A", "10 20 30\n40 50 60", 0},
	{"A←2 3 4⍴⍳24 ⋄ (,[2 3]A)←2 12⍴-⍳24⋄⍴A⋄A[2;3;]", "2 3 4\n¯21 ¯22 ¯23 ¯24", 0},
	{"A←'GROWTH' ⋄ (2 3⍴A)←2 3⍴-⍳6 ⋄ (4⍴A)←⍳4 ⋄ A", "1 2 3 4 ¯5 ¯6", 0},
	{"A←3 4⍴⍳12 ⋄ (⌽A)←3 4⍴'STOPSPINODER' ⋄ A", "P O T S\nN I P S\nR E D O", 0},
	{"A←2 3⍴⍳6 ⋄ (⌽[1]A)←2 3⍴-⍳6 ⋄ A", "¯4 ¯5 ¯6\n¯1 ¯2 ¯3", 0},
	{"A←⍳6 ⋄ (2⌽A)←10×⍳6 ⋄ A", "50 60 10 20 30 40", 0},
	{"A←3 4⍴⍳12 ⋄ (1 ¯1 2 ¯2⊖A)←3 4⍴4×⍳12 ⋄ A", "36 24 28 48\n4 40 44 16\n20 8 12 32", 0},
	{"A←3 4⍴⍳12 ⋄ (1 ¯1 2 ¯2⌽[1]A)←3 4⍴4×⍳12 ⋄ A", "36 24 28 48\n4 40 44 16\n20 8 12 32", 0},
	{"A←⍳5 ⋄ (2↑A)← 10 20 ⋄ A", "10 20 3 4 5", 0},
	{"A←2 3⍴⍳6 ⋄ (¯2↑[2]A)←2 2⍴10×⍳4 ⋄ A", "1 10 20\n4 30 40", 0},
	{"A←3 3⍴⍳9 ⋄ (1 1⍉A)←10 20 30 ⋄ A", "10 2 3\n4 20 6\n7 8 30", 0},
	{"A←3 3⍴'STYPIEANT' ⋄ (⍉A)←3 3⍴⍳9 ⋄ A", "1 4 7\n2 5 8\n3 6 9", 0},
	{"⍝ First (↓) and Pick (⊃) are not implemented", "", 0},

	{"⍝ IBM APL Language, 3rd edition, June 1976.", "", 0},
	{"1000×(1+.06÷1 4 12 365)*10×1 4 12 365", "1790.8 1814 1819.4 1822", short},
	{"Area ← 3×4\nX←2+⎕←3×Y←4\nX\nY", "12\n14\n4", 0},

	{"⍝ Lambda expressions.", "", 0},
	{"{2×⍵}3", "6", 0},           // lambda in monadic context
	{"2{⍺+3{⍺×⍵}⍵+2}2", "14", 0}, // nested lambas
	{"2{(⍺+3){⍺×⍵}⍵+⍺{⍺+1+⍵}1+2}2", "40", 0},
	{"1{1+⍺{1+⍺{1+⍺+⍵}1+⍵}1+⍵}1", "7", 0},
	{"2{}4", "", 0}, // empty lambda expression ignores arguments
	{"{⍺×⍵}/2 3 4", "24", 0},
	{`{1:1+2⋄{1:1+⍵}3}4`, "3", 0},

	{"⍝ Evaluation order", "", 0},
	{"A←1⋄A+(A←2)", "4", 0},
	{"A+A←3", "6", 0},
	{"A←1⋄A{(⍺ ⍵)}A+←10", "11 10", 0},

	{"⍝ Lexical scoping", "", 0},
	{"A←1⋄{A←2⋄A}0⋄A", "2\n1", 0},
	{"X←{A←3⋄B←4⋄0:ignored⋄42}0⋄X⋄A⋄B", "42\nA\nB", 0},
	{"{A←1⋄{A←⍵}⍵+1}1", "2", 0},
	{"A←1⋄S←{A←2}0⋄A", "1", 0},
	{"A←1⋄S←{A⊢←2}0⋄A", "2", 0}, // overwrite a global
	{"A←1⍴1⋄S←{A[1]←2}0⋄A", "2", 0},
	{"A←1⋄{A+←1⋄A}0⋄A", "2\n2", 0},
	{"+X←{A←3⋄B←4}0", "4", 0},

	{"⍝ Default left argument", "", 0},
	{"f←{⍺←3⋄⍺+⍵}⋄ f 4 ⋄ 1 f 4", "7\n5", 0},

	{"⍝ Recursion", "", 0},
	{`S←0.0 n→f "%.0f" ⋄ f←{⍵≤1: 1 ⋄ ⍵×∇⍵-1} ⋄ f 10`, "3628800", 0},
	{"S←0{⍺>20:⍺⋄⍵∇⎕←⍺+⍵}1", "1\n2\n3\n5\n8\n13\n21\n34", 0},

	{"⍝ Tail call", "", 0},
	{"{⍵>1000:⍵⋄∇⍵+1}1", "1001", 0},

	{"⍝ Trains, forks, atops", "", 0},
	{"-,÷ 5", "¯0.2", 0},
	{"(-,÷)5", "¯5 0.2", 0},
	{"3(+×-)1", "8", 0},
	{"(+⌿÷≢)3+⍳13", "10", 0},
	{"(⍳{⍺/⍵}⍳)3", "1 2 2 3 3 3", 0},
	{"(2/⍳)3", "1 1 2 2 3 3", 0},
	{"6(+,-,×,÷)2", "8 4 12 3", 0},
	{"6(⌽+,-,×,÷)2", "3 12 4 8", 0},
	{"(⍳12) (⍳∘1 ≥)9", "9", 0},
	{"(*-)1", "0.36788", short | float},
	{"2(*-)1", "2.7183", short | float},
	{"1(*-)2", "0.36788", short | float},
	{"3(÷+×-)1", "0.125", 0},
	{"(÷+×-)4", "¯0.0625", 0},
	{"(⌊÷+×-)4", "¯0.25", 0},
	{"6(⌊÷+×-)4", "0.2", 0},
	{"(3+*)4", "57.598", short | float}, // Agh fork
	//{"(⍳(/∘⊢)⍳)3", "1 2 2 3 3 3", 0}, // The hybrid token does not parse.

	{"⍝ π", "", 0},
	{".5*⍨6×+/÷2*⍨⍳1000", "3.1406", short | float},
	{"4×-/÷¯1+2×⍳100", "3.1316", short},
	{"4×+/{(⍵ ⍴ 1 0 ¯1 0)÷⍳⍵}100", "3.1216", short},

	{"⍝ Conways game of life", "", 0},
	{"A←5 5⍴(23⍴2)⊤1215488⋄l←{3=S-⍵∧4=S←({+/,⍵}⌺3 3)⍵}⋄(l⍣8)A", "0 0 0 0 0\n0 0 0 0 0\n0 0 0 0 1\n0 0 1 0 1\n0 0 0 1 1", 0},

	{"⍝ Go interface: package strings", "", 0},
	{`u←s→toupper ⋄ u "alpha"`, "ALPHA", 0},
	{`";" s→join "alpha" "beta" `, "alpha;beta", 0},

	{"⍝ Dictionaries", "", 0},
	{"D←`alpha#1 2 3⋄D[`alpha]←`xyz⋄D", "alpha: xyz", 0},
	{"D←`alpha#1⋄D[`alpha`beta]←3 4⋄D", "alpha: 3\nbeta: 4", 0},
	{"D←`a`b`c#1⋄D⋄#D", "a: 1\nb: 1\nc: 1\na b c", 0},
	{"D←`a`b`c#1 2 3⋄G←D[`a`c]⋄G", "a: 1\nc: 3", 0},

	{"⍝ Object, xgo example", "", 0},
	{"X←xgo→t 0⋄X[`V]←`a`b⋄X[`V]", "a b", 0},
	{"X←xgo→t 0⋄X[`I]←55⋄X[`inc]⍨0⋄X[`I]", "56", small},
	{"X←xgo→t 0⋄X[`V]←'abcd'⋄X[`join]⍨'+'", "4 a+b+c+d", small},
	{"S←xgo→s 0⋄#[1]S", "sum", 0},

	{"⍝ Examples from github.com/DhavalDalal/APL-For-FP-Programmers", "", 0},
	// filter←{(⍺⍺¨⍵)⌿⍵} // 01-primes
	{"f←{(2=+⌿0=X∘.|X)⌿X←⍳⍵} ⋄ f 42", "2 3 5 7 11 13 17 19 23 29 31 37 41", 0},        // 01-primes
	{"⎕IO←0 ⋄ f←{(~X∊X∘.×X)⌿X←2↓⍳⍵} ⋄ f 42", "2 3 5 7 11 13 17 19 23 29 31 37 41", 0}, // 01-primes
	// ⎕IO←0 ⋄ sieve ← {⍸⊃{~⍵[⍺]:⍵ ⋄ 0@(⍺×2↓⍳⌈(≢⍵)÷⍺)⊢⍵}/⌽(⊂0 0,(⍵-2)⍴1),⍳⍵} // 02-sieve
	// ⎕IO←0 ⋄ triples←{{⍵/⍨(2⌷x)=+⌿2↑x←×⍨⍵}⍉↑,1+⍳⍵ ⍵ ⍵}// 03-pythagoreans
	// ⎕IO←0 ⋄ '-:'⊣@(' '=⊢)¨(14⍴(4⍴1),0)(17⍴1 1 0)\¨⊂⍉(⎕D,6↑⎕A)[(12⍴16)⊤?10⍴2*48] // 04-MacAddress
	// life←{⊃1 ⍵∨.∧3 4=+⌿,1 0 ¯1∘.⊖1 0 ¯1⌽¨⊂⍵} // 05-life
	// life2←{3=s-⍵∧4=s←{+/,⍵}⌺3 3⊢⍵} // 05-life

	// Trees: https://youtu.be/hzPd3umu78g

	//https://github.com/theaplroom/apl-sound-wave/blob/master/src/DSP.dyalog

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

const (
	short  int = 1 << iota // short format
	float                  // only for floating point towers
	small                  // normal tower only
	sfloat                 // short float only
	cmplx                  // short %g complex format
	cmplxf                 // short %f complex format
)

func TestNormal(t *testing.T) {
	testApl(t, nil, 0)
}

func TestBig(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	testApl(t, big.SetBigTower, cmplx|small|float)
}

func TestPrecise(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	testApl(t, func(a *apl.Apl) { big.SetPreciseTower(a, 256) }, small|sfloat)
}

func testApl(t *testing.T, tower func(*apl.Apl), skip int) {
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

		// Skip tests for unsupported numberic types
		if skip&tc.flag != 0 {
			continue
		}

		var buf strings.Builder
		a := apl.New(&buf)
		numbers.Register(a)
		if tower != nil {
			tower(a)
		}
		Register(a)
		operators.Register(a)
		aplstrings.Register(a)
		xgo.Register(a)

		// Set numeric formats.
		m := make(map[reflect.Type]string)
		m[reflect.TypeOf(rat0)] = "%.5g"
		if tc.flag&short != 0 {
			m[reflect.TypeOf(numbers.Float(0))] = "%.5g"
			m[reflect.TypeOf(big.Float{})] = "%.5g"
		}
		if tc.flag&cmplx != 0 {
			m[reflect.TypeOf(numbers.Complex(0))] = "%.5gJ%.5g"
			m[reflect.TypeOf(big.Complex{})] = "%.5gJ%.5g"
		}
		if tc.flag&cmplxf != 0 {
			m[reflect.TypeOf(numbers.Complex(0))] = "%.2fJ%.2f"
		}
		for t, f := range m {
			if num, ok := a.Tower.Numbers[t]; ok {
				num.Format = f
				a.Tower.Numbers[t] = num
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

var rat0, _ = big.ParseRat("0")
var spaces = regexp.MustCompile(`  *`)
var newline = regexp.MustCompile(`\n *`)
