package decimal

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/cockroachdb/apd/v3"
)

// Decimal wraps apd.Decimal with proper JSON support.
// Accepts both JSON numbers (123.45) and strings ("123.45") on input.
// Also handles the legacy object format produced by marshalling a bare apd.Decimal.
// Always marshals as a JSON string to preserve precision.
type Decimal struct {
	apd.Decimal
}

// From wraps an existing apd.Decimal.
func From(d apd.Decimal) Decimal {
	return Decimal{Decimal: d}
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Decimal.Text('f') + `"`), nil
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	// Case 1: JSON string — "123.45"
	// Case 2: JSON number — 123.45
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	if len(s) > 0 && s[0] != '{' {
		_, _, err := d.Decimal.SetString(s)
		if err != nil {
			return fmt.Errorf("invalid decimal value: %s", string(data))
		}
		return nil
	}

	// Case 3: legacy object format from bare apd.Decimal marshal, e.g.
	//   {"Form":0,"Negative":false,"Exponent":-2,"Coeff":"10050"}
	// or with Coeff as an object:
	//   {"Form":0,"Negative":false,"Exponent":-2,"Coeff":{"abs":[100,50],"neg":false}}
	var legacy struct {
		Negative bool            `json:"Negative"`
		Exponent int32           `json:"Exponent"`
		Coeff    json.RawMessage `json:"Coeff"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return fmt.Errorf("invalid decimal value: %s", string(data))
	}

	coeff := new(big.Int)
	coeffStr := string(legacy.Coeff)

	switch {
	case len(coeffStr) == 0 || coeffStr == "null" || coeffStr == "{}" || coeffStr == "\"\"":
		// Zero / empty coefficient → 0
	case coeffStr[0] == '"':
		// String representation: "10050"
		raw := coeffStr[1 : len(coeffStr)-1]
		if _, ok := coeff.SetString(raw, 10); !ok {
			return fmt.Errorf("invalid decimal coefficient: %s", raw)
		}
	case coeffStr[0] == '{':
		// Object representation from apd.BigInt: {"abs":[words...],"neg":false}
		var bigIntObj struct {
			Abs []uint64 `json:"abs"`
			Neg bool     `json:"neg"`
		}
		if err := json.Unmarshal(legacy.Coeff, &bigIntObj); err != nil {
			return fmt.Errorf("invalid decimal coefficient object: %s", coeffStr)
		}
		if len(bigIntObj.Abs) > 0 {
			// Reconstruct big.Int from little-endian base-2^64 words
			coeff.SetUint64(0)
			base := new(big.Int).SetUint64(0)
			base.SetBit(base, 64, 1) // 2^64
			for i := len(bigIntObj.Abs) - 1; i >= 0; i-- {
				coeff.Mul(coeff, base)
				coeff.Add(coeff, new(big.Int).SetUint64(bigIntObj.Abs[i]))
			}
		}
		if bigIntObj.Neg {
			coeff.Neg(coeff)
		}
	default:
		// Plain number: 10050
		if _, ok := coeff.SetString(coeffStr, 10); !ok {
			return fmt.Errorf("invalid decimal coefficient: %s", coeffStr)
		}
	}

	d.Decimal.Exponent = legacy.Exponent
	d.Decimal.Negative = legacy.Negative
	d.Decimal.Coeff.SetMathBigInt(coeff)
	return nil
}
