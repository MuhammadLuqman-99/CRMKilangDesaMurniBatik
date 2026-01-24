// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Common currencies with their decimal places
var currencyDecimals = map[string]int{
	"USD": 2, "EUR": 2, "GBP": 2, "JPY": 0, "CNY": 2,
	"AUD": 2, "CAD": 2, "CHF": 2, "HKD": 2, "SGD": 2,
	"SEK": 2, "NOK": 2, "DKK": 2, "NZD": 2, "MXN": 2,
	"INR": 2, "BRL": 2, "KRW": 0, "ZAR": 2, "RUB": 2,
	"TRY": 2, "PLN": 2, "THB": 2, "IDR": 0, "MYR": 2,
	"PHP": 2, "VND": 0, "AED": 2, "SAR": 2, "COP": 2,
	"CLP": 0, "PEN": 2, "ARS": 2, "CZK": 2, "HUF": 2,
	"ILS": 2, "TWD": 2, "PKR": 2, "EGP": 2, "NGN": 2,
	"BDT": 2, "UAH": 2, "RON": 2, "KES": 2, "GHS": 2,
}

// Money represents a monetary amount with currency.
// Uses fixed-point arithmetic to avoid floating-point precision issues.
type Money struct {
	// Amount in smallest currency unit (e.g., cents for USD)
	Amount   int64  `json:"amount" bson:"amount"`
	Currency string `json:"currency" bson:"currency"`
}

// MoneyError represents a money-related error.
var (
	ErrInvalidCurrency       = errors.New("invalid currency code")
	ErrCurrencyMismatch      = errors.New("currency mismatch")
	ErrNegativeAmount        = errors.New("amount cannot be negative")
	ErrDivisionByZero        = errors.New("division by zero")
	ErrInvalidPercentage     = errors.New("invalid percentage value")
	ErrOverflow              = errors.New("arithmetic overflow")
)

// NewMoney creates a new Money value from the smallest currency unit.
func NewMoney(amount int64, currency string) (Money, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if _, ok := currencyDecimals[currency]; !ok {
		return Money{}, ErrInvalidCurrency
	}
	return Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

// NewMoneyFromFloat creates a new Money value from a float amount.
func NewMoneyFromFloat(amount float64, currency string) (Money, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	decimals, ok := currencyDecimals[currency]
	if !ok {
		return Money{}, ErrInvalidCurrency
	}

	// Convert to smallest unit
	multiplier := math.Pow(10, float64(decimals))
	amountInSmallestUnit := int64(math.Round(amount * multiplier))

	return Money{
		Amount:   amountInSmallestUnit,
		Currency: currency,
	}, nil
}

// MustNewMoney creates a new Money value, panicking on error.
func MustNewMoney(amount int64, currency string) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

// Zero returns a zero money value for the given currency.
func Zero(currency string) (Money, error) {
	return NewMoney(0, currency)
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive returns true if the amount is positive.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsNegative returns true if the amount is negative.
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// Abs returns the absolute value.
func (m Money) Abs() Money {
	if m.Amount < 0 {
		return Money{Amount: -m.Amount, Currency: m.Currency}
	}
	return m
}

// Negate returns the negated value.
func (m Money) Negate() Money {
	return Money{Amount: -m.Amount, Currency: m.Currency}
}

// Add adds two money values.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}

	// Check for overflow
	if other.Amount > 0 && m.Amount > math.MaxInt64-other.Amount {
		return Money{}, ErrOverflow
	}
	if other.Amount < 0 && m.Amount < math.MinInt64-other.Amount {
		return Money{}, ErrOverflow
	}

	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Subtract subtracts another money value.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}

	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

// Multiply multiplies by a factor.
func (m Money) Multiply(factor float64) Money {
	return Money{
		Amount:   int64(math.Round(float64(m.Amount) * factor)),
		Currency: m.Currency,
	}
}

// MultiplyInt multiplies by an integer factor.
func (m Money) MultiplyInt(factor int64) (Money, error) {
	// Check for overflow
	if factor != 0 {
		if m.Amount > 0 && factor > 0 && m.Amount > math.MaxInt64/factor {
			return Money{}, ErrOverflow
		}
		if m.Amount < 0 && factor > 0 && m.Amount < math.MinInt64/factor {
			return Money{}, ErrOverflow
		}
		if m.Amount > 0 && factor < 0 && factor < math.MinInt64/m.Amount {
			return Money{}, ErrOverflow
		}
		if m.Amount < 0 && factor < 0 && m.Amount < math.MaxInt64/factor {
			return Money{}, ErrOverflow
		}
	}

	return Money{
		Amount:   m.Amount * factor,
		Currency: m.Currency,
	}, nil
}

// Divide divides by a divisor.
func (m Money) Divide(divisor float64) (Money, error) {
	if divisor == 0 {
		return Money{}, ErrDivisionByZero
	}

	return Money{
		Amount:   int64(math.Round(float64(m.Amount) / divisor)),
		Currency: m.Currency,
	}, nil
}

// DivideInt divides by an integer divisor.
func (m Money) DivideInt(divisor int64) (Money, error) {
	if divisor == 0 {
		return Money{}, ErrDivisionByZero
	}

	return Money{
		Amount:   m.Amount / divisor,
		Currency: m.Currency,
	}, nil
}

// Percentage calculates a percentage of the amount.
func (m Money) Percentage(percent float64) (Money, error) {
	if percent < 0 || percent > 100 {
		return Money{}, ErrInvalidPercentage
	}

	return Money{
		Amount:   int64(math.Round(float64(m.Amount) * percent / 100)),
		Currency: m.Currency,
	}, nil
}

// Split splits the amount into n equal parts.
// Returns the parts and any remainder that couldn't be evenly divided.
func (m Money) Split(n int) ([]Money, Money, error) {
	if n <= 0 {
		return nil, Money{}, ErrDivisionByZero
	}

	quotient := m.Amount / int64(n)
	remainder := m.Amount % int64(n)

	parts := make([]Money, n)
	for i := 0; i < n; i++ {
		parts[i] = Money{Amount: quotient, Currency: m.Currency}
	}

	return parts, Money{Amount: remainder, Currency: m.Currency}, nil
}

// Allocate allocates the amount according to ratios.
func (m Money) Allocate(ratios []int) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, errors.New("ratios cannot be empty")
	}

	total := 0
	for _, r := range ratios {
		if r < 0 {
			return nil, errors.New("ratio cannot be negative")
		}
		total += r
	}

	if total == 0 {
		return nil, errors.New("total ratio cannot be zero")
	}

	results := make([]Money, len(ratios))
	remainder := m.Amount

	for i, ratio := range ratios {
		share := m.Amount * int64(ratio) / int64(total)
		results[i] = Money{Amount: share, Currency: m.Currency}
		remainder -= share
	}

	// Distribute remainder to first non-zero allocation
	for i := range results {
		if ratios[i] > 0 && remainder > 0 {
			results[i].Amount++
			remainder--
		}
		if remainder == 0 {
			break
		}
	}

	return results, nil
}

// Compare compares two money values.
// Returns -1 if m < other, 0 if equal, 1 if m > other.
func (m Money) Compare(other Money) (int, error) {
	if m.Currency != other.Currency {
		return 0, ErrCurrencyMismatch
	}

	if m.Amount < other.Amount {
		return -1, nil
	}
	if m.Amount > other.Amount {
		return 1, nil
	}
	return 0, nil
}

// Equals checks if two money values are equal.
func (m Money) Equals(other Money) bool {
	return m.Currency == other.Currency && m.Amount == other.Amount
}

// GreaterThan checks if m is greater than other.
func (m Money) GreaterThan(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	if err != nil {
		return false, err
	}
	return cmp > 0, nil
}

// GreaterThanOrEqual checks if m is greater than or equal to other.
func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	if err != nil {
		return false, err
	}
	return cmp >= 0, nil
}

// LessThan checks if m is less than other.
func (m Money) LessThan(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	if err != nil {
		return false, err
	}
	return cmp < 0, nil
}

// LessThanOrEqual checks if m is less than or equal to other.
func (m Money) LessThanOrEqual(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	if err != nil {
		return false, err
	}
	return cmp <= 0, nil
}

// Float returns the amount as a float.
func (m Money) Float() float64 {
	decimals := currencyDecimals[m.Currency]
	divisor := math.Pow(10, float64(decimals))
	return float64(m.Amount) / divisor
}

// Format formats the money value as a string.
func (m Money) Format() string {
	decimals := currencyDecimals[m.Currency]
	if decimals == 0 {
		return fmt.Sprintf("%s %d", m.Currency, m.Amount)
	}

	divisor := math.Pow(10, float64(decimals))
	value := float64(m.Amount) / divisor
	format := fmt.Sprintf("%s %%.%df", m.Currency, decimals)
	return fmt.Sprintf(format, value)
}

// String returns a string representation.
func (m Money) String() string {
	return m.Format()
}

// GetDecimals returns the number of decimal places for the currency.
func (m Money) GetDecimals() int {
	return currencyDecimals[m.Currency]
}

// Min returns the minimum of two money values.
func Min(a, b Money) (Money, error) {
	if a.Currency != b.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	if a.Amount < b.Amount {
		return a, nil
	}
	return b, nil
}

// Max returns the maximum of two money values.
func Max(a, b Money) (Money, error) {
	if a.Currency != b.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	if a.Amount > b.Amount {
		return a, nil
	}
	return b, nil
}

// Sum sums multiple money values.
func Sum(amounts ...Money) (Money, error) {
	if len(amounts) == 0 {
		return Money{}, nil
	}

	result := amounts[0]
	for i := 1; i < len(amounts); i++ {
		var err error
		result, err = result.Add(amounts[i])
		if err != nil {
			return Money{}, err
		}
	}
	return result, nil
}

// IsSupportedCurrency checks if a currency code is supported.
func IsSupportedCurrency(currency string) bool {
	_, ok := currencyDecimals[strings.ToUpper(currency)]
	return ok
}

// SupportedCurrencies returns a list of supported currency codes.
func SupportedCurrencies() []string {
	currencies := make([]string, 0, len(currencyDecimals))
	for c := range currencyDecimals {
		currencies = append(currencies, c)
	}
	return currencies
}
