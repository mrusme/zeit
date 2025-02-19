package z

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

var fractional bool

func fmtDuration(dur time.Duration) string {
	return fmtHours(decimal.NewFromFloat(dur.Hours()))
}

func fmtHours(hours decimal.Decimal) string {
	if fractional {
		return hours.StringFixed(2)
	} else {
		return fmt.Sprintf(
			"%s:%02s",
			hours.Floor(), // hours
			hours.Sub(hours.Floor()).
				Mul(decimal.NewFromFloat(.6)).
				Mul(decimal.NewFromInt(100)).
				Floor())
	}
}
