package util

// Constants for all supported currencies
const (
	NGN = "NGN"
	RUB = "RUB"
	CNY = "CNY"
	USD = "USD"
	GBP = "GBP"
	EUR = "EUR"
	CAD = "CAD"
)

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case NGN, RUB, CNY, USD, GBP, EUR, CAD:
		return true
	}
	return false
}