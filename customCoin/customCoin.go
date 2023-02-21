package customCoin

import (
	"errors"
	"fmt"
	"regexp"

	bigNumber "lukechampine.com/uint128"
)

var (
	errInvalidCoin = errors.New("coin is invalid")
)

var (
	reDnmString = `[a-zA-Z][a-zA-Z0-9/]{2,127}`
	reDecAmt    = `[[:digit:]]+(?:\.[[:digit:]]+)?|\.[[:digit:]]+`
	reSpc       = `[[:space:]]*`
	pattern     = fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmt, reSpc, reDnmString)
	parseRe     = regexp.MustCompile(pattern)
)

// Parse parses a coin into amount and denom.
func Parse(c string) (amount bigNumber.Uint128, denom string, err error) {
	parsed := parseRe.FindStringSubmatch(c)

	if len(parsed) != 3 {
		num, _ := bigNumber.FromString("0")
		return num, "", errInvalidCoin
	}

	amountStr := parsed[1]
	denom = parsed[2]

	amount, error := bigNumber.FromString(amountStr)
	if error != nil {
		num, _ := bigNumber.FromString("0")
		return num, "", errInvalidCoin
	}

	return
}
