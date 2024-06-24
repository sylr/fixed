**Summary**

A fixed place numeric library designed for performance.

The c++ version is available [here](https://github.com/robaho/cpp_fixed).

All numbers have a fixed **6** decimal places, and the maximum permitted value
is +- 999_999_999_999, or just under 1000 billion.

The library is safe for concurrent use. It has built-in support for binary and
json marshalling.

It is ideally suited for high performance trading financial systems. All common
math operations are completed with 0 allocs.

**Design Goals**

Primarily developed to improve performance in [go-trader](https://github.com/robaho/go-trader).
Using Fixed rather than decimal.Decimal improves the performance by over 20%,
and a lot less GC activity as well. You can review these changes under the
'fixed' branch.

If you review the go-trader code, you will quickly see that I use dot imports
for the fixed and common packages. Since this is a "business/user" app and not
systems code, this provides 2 major benefits: less verbose code, and I can
easily change the implementation of Fixed without changing lots of LOC - just
the import statement, and some of the wrapper methods in common.

The fixed.Fixed API uses NaN for reporting errors in the common case, since often code is chained like:
```
   result := someFixed.Mul(NewS("123.50"))
```
and this would be a huge pain with error handling. Since all operations involving a NaN result in a NaN,
 any errors quickly surface anyway.

**Performance**

<pre>
goos: darwin
goarch: arm64
pkg: github.com/robaho/fixed
BenchmarkAddFixed-10          	1000000000	         0.6231 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddDecimal-10        	28439079	        41.01 ns/op	      80 B/op	       2 allocs/op
BenchmarkAddBigInt-10         	186084148	         6.444 ns/op	       0 B/op	       0 allocs/op
BenchmarkAddBigFloat-10       	29718478	        39.50 ns/op	      48 B/op	       1 allocs/op
BenchmarkMulFixed-10          	548802238	         2.180 ns/op	       0 B/op	       0 allocs/op
BenchmarkMulDecimal-10        	27966894	        41.38 ns/op	      80 B/op	       2 allocs/op
BenchmarkMulBigInt-10         	173213432	         6.921 ns/op	       0 B/op	       0 allocs/op
BenchmarkMulBigFloat-10       	83455525	        14.31 ns/op	       0 B/op	       0 allocs/op
BenchmarkDivFixed-10          	487855944	         2.459 ns/op	       0 B/op	       0 allocs/op
BenchmarkDivDecimal-10        	 4430803	       270.7 ns/op	     392 B/op	      13 allocs/op
BenchmarkDivBigInt-10         	52796953	        21.76 ns/op	       8 B/op	       1 allocs/op
BenchmarkDivBigFloat-10       	23765239	        49.45 ns/op	      24 B/op	       2 allocs/op
BenchmarkCmpFixed-10          	1000000000	         0.3121 ns/op	       0 B/op	       0 allocs/op
BenchmarkCmpDecimal-10        	256577332	         4.681 ns/op	       0 B/op	       0 allocs/op
BenchmarkCmpBigInt-10         	348158856	         3.436 ns/op	       0 B/op	       0 allocs/op
BenchmarkCmpBigFloat-10       	383045502	         3.118 ns/op	       0 B/op	       0 allocs/op
BenchmarkStringFixed-10       	35242720	        33.09 ns/op	      16 B/op	       1 allocs/op
BenchmarkStringNFixed-10      	35731167	        32.77 ns/op	      16 B/op	       1 allocs/op
BenchmarkStringDecimal-10     	10620631	       111.8 ns/op	      64 B/op	       5 allocs/op
BenchmarkStringBigInt-10      	19592462	        60.31 ns/op	      24 B/op	       2 allocs/op
BenchmarkStringBigFloat-10    	 5391572	       223.7 ns/op	     184 B/op	       8 allocs/op
BenchmarkWriteTo-10           	70938927	        16.43 ns/op	      15 B/op	       0 allocs/op
PASS
ok  	github.com/robaho/fixed	29.001s
</pre>

The "decimal" above is the common [shopspring decimal](https://github.com/shopspring/decimal) library

**Compatibility with SQL drivers**

By default `Fixed` implements `decomposer.Decimal` interface for database
drivers that support it. To use `sql.Scanner` and `driver.Valuer` implementation
flag `sql_scanner` must be specified on build.
