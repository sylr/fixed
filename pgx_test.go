//go:build pgx

package fixed

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestNumericValue(t *testing.T) {
	tests := []struct {
		name    string
		input   Fixed
		wantNaN bool
		wantInt int64
		wantExp int32
	}{
		{"zero", ZERO, false, 0, -nPlaces},
		{"positive integer", NewI(42, 0), false, 42 * scale, -nPlaces},
		{"negative integer", NewI(-7, 0), false, -7 * scale, -nPlaces},
		{"positive decimal", NewS("123.456"), false, 123456000, -nPlaces},
		{"negative decimal", NewS("-0.001"), false, -1000, -nPlaces},
		{"NaN", NaN, true, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := tt.input.NumericValue()
			if err != nil {
				t.Fatalf("NumericValue() error = %v", err)
			}
			if !n.Valid {
				t.Fatal("NumericValue() returned invalid Numeric")
			}
			if n.NaN != tt.wantNaN {
				t.Fatalf("NaN = %v, want %v", n.NaN, tt.wantNaN)
			}
			if tt.wantNaN {
				return
			}
			if n.Int.Int64() != tt.wantInt {
				t.Fatalf("Int = %d, want %d", n.Int.Int64(), tt.wantInt)
			}
			if n.Exp != tt.wantExp {
				t.Fatalf("Exp = %d, want %d", n.Exp, tt.wantExp)
			}
		})
	}
}

func TestScanNumeric(t *testing.T) {
	tests := []struct {
		name     string
		numeric  pgtype.Numeric
		expected Fixed
	}{
		{
			name:     "zero",
			numeric:  pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true},
			expected: ZERO,
		},
		{
			name:     "positive integer",
			numeric:  pgtype.Numeric{Int: big.NewInt(42), Exp: 0, Valid: true},
			expected: NewI(42, 0),
		},
		{
			name:     "negative integer",
			numeric:  pgtype.Numeric{Int: big.NewInt(-7), Exp: 0, Valid: true},
			expected: NewI(-7, 0),
		},
		{
			name:     "decimal from scaled int",
			numeric:  pgtype.Numeric{Int: big.NewInt(123456), Exp: -3, Valid: true},
			expected: NewS("123.456"),
		},
		{
			name:     "small decimal",
			numeric:  pgtype.Numeric{Int: big.NewInt(-1), Exp: -3, Valid: true},
			expected: NewS("-0.001"),
		},
		{
			name:     "NaN",
			numeric:  pgtype.Numeric{NaN: true, Valid: true},
			expected: NaN,
		},
		{
			name:     "invalid (NULL)",
			numeric:  pgtype.Numeric{Valid: false},
			expected: ZERO,
		},
		{
			name:     "infinity",
			numeric:  pgtype.Numeric{InfinityModifier: pgtype.Infinity, Valid: true},
			expected: NaN,
		},
		{
			name:     "large exponent",
			numeric:  pgtype.Numeric{Int: big.NewInt(5), Exp: 6, Valid: true},
			expected: NewI(5000000, 0),
		},
		{
			name:     "precision loss (truncation)",
			numeric:  pgtype.Numeric{Int: big.NewInt(12345678), Exp: -8, Valid: true},
			expected: NewS("0.123456"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f Fixed
			err := f.ScanNumeric(tt.numeric)
			if err != nil {
				t.Fatalf("ScanNumeric() error = %v", err)
			}
			if f != tt.expected {
				t.Fatalf("got %s (fp=%d), want %s (fp=%d)", f.String(), f.fp, tt.expected.String(), tt.expected.fp)
			}
		})
	}
}

func TestNumericRoundTrip(t *testing.T) {
	values := []Fixed{
		ZERO,
		NewS("1"),
		NewS("-1"),
		NewS("123.456"),
		NewS("-123.456"),
		NewS("0.000001"),
		NewS("-0.000001"),
		NewS("999999999999"),
		NewF(3.14),
		NaN,
	}

	for _, original := range values {
		t.Run(original.String(), func(t *testing.T) {
			n, err := original.NumericValue()
			if err != nil {
				t.Fatalf("NumericValue() error = %v", err)
			}

			var restored Fixed
			err = restored.ScanNumeric(n)
			if err != nil {
				t.Fatalf("ScanNumeric() error = %v", err)
			}

			if original != restored {
				t.Fatalf("round-trip failed: %s (fp=%d) -> %s (fp=%d)", original.String(), original.fp, restored.String(), restored.fp)
			}
		})
	}
}

func BenchmarkNumericValue(b *testing.B) {
	f := NewS("123.456")
	for b.Loop() {
		_, _ = f.NumericValue()
	}
}

func BenchmarkScanNumeric(b *testing.B) {
	n := pgtype.Numeric{Int: big.NewInt(123456000), Exp: -6, Valid: true}
	var f Fixed
	for b.Loop() {
		_ = f.ScanNumeric(n)
	}
}
