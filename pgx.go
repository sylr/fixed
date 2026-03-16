//go:build pgx

package fixed

import (
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	_ pgtype.NumericValuer  = (*Fixed)(nil)
	_ pgtype.NumericScanner = (*Fixed)(nil)
)

// NumericValue implements the pgtype.NumericValuer interface for efficient
// pgx encoding without falling back to the driver.Valuer path.
func (f Fixed) NumericValue() (pgtype.Numeric, error) {
	if f.IsNaN() {
		return pgtype.Numeric{NaN: true, Valid: true}, nil
	}
	return pgtype.Numeric{
		Int:   big.NewInt(f.fp),
		Exp:   -nPlaces,
		Valid: true,
	}, nil
}

// ScanNumeric implements the pgtype.NumericScanner interface for efficient
// pgx decoding without falling back to the sql.Scanner path.
func (f *Fixed) ScanNumeric(v pgtype.Numeric) error {
	if !v.Valid {
		*f = ZERO
		return nil
	}

	if v.NaN {
		*f = NaN
		return nil
	}

	if v.InfinityModifier != pgtype.Finite {
		*f = NaN
		return nil
	}

	// Convert to our fixed-point representation: fp = Int * 10^(Exp + nPlaces)
	//
	// We need the value scaled by 10^nPlaces. The Numeric stores value as
	// Int * 10^Exp, so we need Int * 10^(Exp + nPlaces).
	targetExp := int(v.Exp) + nPlaces

	bi := new(big.Int).Set(v.Int)

	if targetExp > 0 {
		p := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(targetExp)), nil)
		bi.Mul(bi, p)
	} else if targetExp < 0 {
		p := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-targetExp)), nil)
		bi.Div(bi, p)
	}

	if !bi.IsInt64() {
		*f = NaN
		return nil
	}

	*f = Fixed{fp: bi.Int64()}
	return nil
}
