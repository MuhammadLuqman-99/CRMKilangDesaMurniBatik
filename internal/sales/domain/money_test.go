// Package domain contains the domain layer for the Sales Pipeline service.
package domain

import (
	"math"
	"testing"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency string
		wantErr  bool
	}{
		{"valid USD", 1000, "USD", false},
		{"valid EUR", 500, "EUR", false},
		{"valid JPY", 1000, "JPY", false},
		{"lowercase currency", 1000, "usd", false},
		{"currency with spaces", 1000, " USD ", false},
		{"invalid currency", 1000, "XYZ", true},
		{"empty currency", 1000, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := NewMoney(tt.amount, tt.currency)

			if tt.wantErr {
				if err == nil {
					t.Error("NewMoney() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewMoney() unexpected error: %v", err)
				return
			}

			if money.Amount != tt.amount {
				t.Errorf("Amount = %d, want %d", money.Amount, tt.amount)
			}
		})
	}
}

func TestNewMoneyFromFloat(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		currency string
		wantAmt  int64
		wantErr  bool
	}{
		{"USD 10.50", 10.50, "USD", 1050, false},
		{"EUR 99.99", 99.99, "EUR", 9999, false},
		{"JPY 100 (no decimals)", 100, "JPY", 100, false},
		{"USD rounding", 10.555, "USD", 1056, false},
		{"invalid currency", 10.00, "XYZ", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := NewMoneyFromFloat(tt.amount, tt.currency)

			if tt.wantErr {
				if err == nil {
					t.Error("NewMoneyFromFloat() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewMoneyFromFloat() unexpected error: %v", err)
				return
			}

			if money.Amount != tt.wantAmt {
				t.Errorf("Amount = %d, want %d", money.Amount, tt.wantAmt)
			}
		})
	}
}

func TestMustNewMoney(t *testing.T) {
	// Test valid
	t.Run("valid", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewMoney panicked unexpectedly: %v", r)
			}
		}()
		money := MustNewMoney(1000, "USD")
		if money.Amount != 1000 {
			t.Errorf("Amount = %d, want 1000", money.Amount)
		}
	})

	// Test invalid panics
	t.Run("invalid panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNewMoney should have panicked")
			}
		}()
		MustNewMoney(1000, "INVALID")
	})
}

func TestZero(t *testing.T) {
	money, err := Zero("USD")
	if err != nil {
		t.Errorf("Zero() error: %v", err)
	}
	if !money.IsZero() {
		t.Error("Zero() should return zero money")
	}
}

func TestMoney_IsZero(t *testing.T) {
	zero := MustNewMoney(0, "USD")
	positive := MustNewMoney(100, "USD")
	negative := Money{Amount: -100, Currency: "USD"}

	if !zero.IsZero() {
		t.Error("IsZero() should be true for 0")
	}
	if positive.IsZero() {
		t.Error("IsZero() should be false for positive")
	}
	if negative.IsZero() {
		t.Error("IsZero() should be false for negative")
	}
}

func TestMoney_IsPositive(t *testing.T) {
	positive := MustNewMoney(100, "USD")
	zero := MustNewMoney(0, "USD")
	negative := Money{Amount: -100, Currency: "USD"}

	if !positive.IsPositive() {
		t.Error("IsPositive() should be true for positive")
	}
	if zero.IsPositive() {
		t.Error("IsPositive() should be false for zero")
	}
	if negative.IsPositive() {
		t.Error("IsPositive() should be false for negative")
	}
}

func TestMoney_IsNegative(t *testing.T) {
	negative := Money{Amount: -100, Currency: "USD"}
	zero := MustNewMoney(0, "USD")
	positive := MustNewMoney(100, "USD")

	if !negative.IsNegative() {
		t.Error("IsNegative() should be true for negative")
	}
	if zero.IsNegative() {
		t.Error("IsNegative() should be false for zero")
	}
	if positive.IsNegative() {
		t.Error("IsNegative() should be false for positive")
	}
}

func TestMoney_Abs(t *testing.T) {
	positive := MustNewMoney(100, "USD")
	negative := Money{Amount: -100, Currency: "USD"}

	absPos := positive.Abs()
	absNeg := negative.Abs()

	if absPos.Amount != 100 {
		t.Errorf("Abs() of positive = %d, want 100", absPos.Amount)
	}
	if absNeg.Amount != 100 {
		t.Errorf("Abs() of negative = %d, want 100", absNeg.Amount)
	}
}

func TestMoney_Negate(t *testing.T) {
	positive := MustNewMoney(100, "USD")
	negative := Money{Amount: -100, Currency: "USD"}

	negatedPos := positive.Negate()
	negatedNeg := negative.Negate()

	if negatedPos.Amount != -100 {
		t.Errorf("Negate() of positive = %d, want -100", negatedPos.Amount)
	}
	if negatedNeg.Amount != 100 {
		t.Errorf("Negate() of negative = %d, want 100", negatedNeg.Amount)
	}
}

func TestMoney_Add(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(500, "USD")
	m3 := MustNewMoney(100, "EUR")

	// Same currency
	result, err := m1.Add(m2)
	if err != nil {
		t.Errorf("Add() error: %v", err)
	}
	if result.Amount != 1500 {
		t.Errorf("Add() = %d, want 1500", result.Amount)
	}

	// Different currency
	_, err = m1.Add(m3)
	if err == nil {
		t.Error("Add() should error with different currencies")
	}
}

func TestMoney_Add_Overflow(t *testing.T) {
	m1 := Money{Amount: math.MaxInt64 - 100, Currency: "USD"}
	m2 := Money{Amount: 200, Currency: "USD"}

	_, err := m1.Add(m2)
	if err == nil {
		t.Error("Add() should error on overflow")
	}
}

func TestMoney_Subtract(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(300, "USD")
	m3 := MustNewMoney(100, "EUR")

	// Same currency
	result, err := m1.Subtract(m2)
	if err != nil {
		t.Errorf("Subtract() error: %v", err)
	}
	if result.Amount != 700 {
		t.Errorf("Subtract() = %d, want 700", result.Amount)
	}

	// Different currency
	_, err = m1.Subtract(m3)
	if err == nil {
		t.Error("Subtract() should error with different currencies")
	}
}

func TestMoney_Multiply(t *testing.T) {
	money := MustNewMoney(1000, "USD")

	result := money.Multiply(1.5)
	if result.Amount != 1500 {
		t.Errorf("Multiply(1.5) = %d, want 1500", result.Amount)
	}

	result = money.Multiply(0.5)
	if result.Amount != 500 {
		t.Errorf("Multiply(0.5) = %d, want 500", result.Amount)
	}

	result = money.Multiply(0)
	if result.Amount != 0 {
		t.Errorf("Multiply(0) = %d, want 0", result.Amount)
	}
}

func TestMoney_MultiplyInt(t *testing.T) {
	money := MustNewMoney(1000, "USD")

	result, err := money.MultiplyInt(3)
	if err != nil {
		t.Errorf("MultiplyInt() error: %v", err)
	}
	if result.Amount != 3000 {
		t.Errorf("MultiplyInt(3) = %d, want 3000", result.Amount)
	}
}

func TestMoney_Divide(t *testing.T) {
	money := MustNewMoney(1000, "USD")

	result, err := money.Divide(4)
	if err != nil {
		t.Errorf("Divide() error: %v", err)
	}
	if result.Amount != 250 {
		t.Errorf("Divide(4) = %d, want 250", result.Amount)
	}

	// Division by zero
	_, err = money.Divide(0)
	if err == nil {
		t.Error("Divide(0) should error")
	}
}

func TestMoney_DivideInt(t *testing.T) {
	money := MustNewMoney(1000, "USD")

	result, err := money.DivideInt(4)
	if err != nil {
		t.Errorf("DivideInt() error: %v", err)
	}
	if result.Amount != 250 {
		t.Errorf("DivideInt(4) = %d, want 250", result.Amount)
	}

	// Division by zero
	_, err = money.DivideInt(0)
	if err == nil {
		t.Error("DivideInt(0) should error")
	}
}

func TestMoney_Percentage(t *testing.T) {
	money := MustNewMoney(10000, "USD") // $100.00

	// 10%
	result, err := money.Percentage(10)
	if err != nil {
		t.Errorf("Percentage() error: %v", err)
	}
	if result.Amount != 1000 {
		t.Errorf("Percentage(10) = %d, want 1000", result.Amount)
	}

	// 50%
	result, err = money.Percentage(50)
	if err != nil {
		t.Errorf("Percentage() error: %v", err)
	}
	if result.Amount != 5000 {
		t.Errorf("Percentage(50) = %d, want 5000", result.Amount)
	}

	// Invalid percentage
	_, err = money.Percentage(-10)
	if err == nil {
		t.Error("Percentage(-10) should error")
	}

	_, err = money.Percentage(110)
	if err == nil {
		t.Error("Percentage(110) should error")
	}
}

func TestMoney_Split(t *testing.T) {
	money := MustNewMoney(1000, "USD") // $10.00

	parts, remainder, err := money.Split(3)
	if err != nil {
		t.Errorf("Split() error: %v", err)
	}

	if len(parts) != 3 {
		t.Errorf("Split() returned %d parts, want 3", len(parts))
	}

	// 1000 / 3 = 333 per part, remainder = 1
	for i, part := range parts {
		if part.Amount != 333 {
			t.Errorf("parts[%d].Amount = %d, want 333", i, part.Amount)
		}
	}
	if remainder.Amount != 1 {
		t.Errorf("remainder.Amount = %d, want 1", remainder.Amount)
	}

	// Split into zero parts
	_, _, err = money.Split(0)
	if err == nil {
		t.Error("Split(0) should error")
	}
}

func TestMoney_Allocate(t *testing.T) {
	money := MustNewMoney(100, "USD") // $1.00

	// Allocate 50/30/20
	results, err := money.Allocate([]int{50, 30, 20})
	if err != nil {
		t.Errorf("Allocate() error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Allocate() returned %d results, want 3", len(results))
	}

	total := int64(0)
	for _, r := range results {
		total += r.Amount
	}
	if total != 100 {
		t.Errorf("Allocate() total = %d, want 100", total)
	}

	// Empty ratios
	_, err = money.Allocate([]int{})
	if err == nil {
		t.Error("Allocate() with empty ratios should error")
	}

	// Negative ratio
	_, err = money.Allocate([]int{50, -10, 60})
	if err == nil {
		t.Error("Allocate() with negative ratio should error")
	}
}

func TestMoney_Compare(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(500, "USD")
	m3 := MustNewMoney(1000, "USD")
	m4 := MustNewMoney(1000, "EUR")

	// Greater
	cmp, err := m1.Compare(m2)
	if err != nil || cmp != 1 {
		t.Errorf("Compare() = %d, want 1", cmp)
	}

	// Less
	cmp, err = m2.Compare(m1)
	if err != nil || cmp != -1 {
		t.Errorf("Compare() = %d, want -1", cmp)
	}

	// Equal
	cmp, err = m1.Compare(m3)
	if err != nil || cmp != 0 {
		t.Errorf("Compare() = %d, want 0", cmp)
	}

	// Different currency
	_, err = m1.Compare(m4)
	if err == nil {
		t.Error("Compare() should error with different currencies")
	}
}

func TestMoney_Equals(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(1000, "USD")
	m3 := MustNewMoney(500, "USD")
	m4 := MustNewMoney(1000, "EUR")

	if !m1.Equals(m2) {
		t.Error("Same amount and currency should be equal")
	}
	if m1.Equals(m3) {
		t.Error("Different amount should not be equal")
	}
	if m1.Equals(m4) {
		t.Error("Different currency should not be equal")
	}
}

func TestMoney_ComparisonMethods(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(500, "USD")

	gt, _ := m1.GreaterThan(m2)
	if !gt {
		t.Error("1000 should be greater than 500")
	}

	gte, _ := m1.GreaterThanOrEqual(m2)
	if !gte {
		t.Error("1000 should be greater than or equal to 500")
	}

	lt, _ := m2.LessThan(m1)
	if !lt {
		t.Error("500 should be less than 1000")
	}

	lte, _ := m2.LessThanOrEqual(m1)
	if !lte {
		t.Error("500 should be less than or equal to 1000")
	}
}

func TestMoney_Float(t *testing.T) {
	m1 := MustNewMoney(1050, "USD")
	if m1.Float() != 10.50 {
		t.Errorf("Float() = %f, want 10.50", m1.Float())
	}

	m2 := MustNewMoney(100, "JPY")
	if m2.Float() != 100 {
		t.Errorf("Float() for JPY = %f, want 100", m2.Float())
	}
}

func TestMoney_Format(t *testing.T) {
	m1 := MustNewMoney(1050, "USD")
	if m1.Format() != "USD 10.50" {
		t.Errorf("Format() = %s, want USD 10.50", m1.Format())
	}

	m2 := MustNewMoney(100, "JPY")
	if m2.Format() != "JPY 100" {
		t.Errorf("Format() = %s, want JPY 100", m2.Format())
	}
}

func TestMoney_String(t *testing.T) {
	money := MustNewMoney(1050, "USD")
	if money.String() != "USD 10.50" {
		t.Errorf("String() = %s, want USD 10.50", money.String())
	}
}

func TestMin(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(500, "USD")

	min, err := Min(m1, m2)
	if err != nil {
		t.Errorf("Min() error: %v", err)
	}
	if min.Amount != 500 {
		t.Errorf("Min() = %d, want 500", min.Amount)
	}

	// Different currencies
	m3 := MustNewMoney(100, "EUR")
	_, err = Min(m1, m3)
	if err == nil {
		t.Error("Min() should error with different currencies")
	}
}

func TestMax(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(500, "USD")

	max, err := Max(m1, m2)
	if err != nil {
		t.Errorf("Max() error: %v", err)
	}
	if max.Amount != 1000 {
		t.Errorf("Max() = %d, want 1000", max.Amount)
	}
}

func TestSum(t *testing.T) {
	m1 := MustNewMoney(1000, "USD")
	m2 := MustNewMoney(500, "USD")
	m3 := MustNewMoney(300, "USD")

	sum, err := Sum(m1, m2, m3)
	if err != nil {
		t.Errorf("Sum() error: %v", err)
	}
	if sum.Amount != 1800 {
		t.Errorf("Sum() = %d, want 1800", sum.Amount)
	}

	// Empty sum
	sum, err = Sum()
	if err != nil {
		t.Errorf("Sum() error: %v", err)
	}
	if sum.Amount != 0 {
		t.Error("Empty Sum() should return 0")
	}
}

func TestIsSupportedCurrency(t *testing.T) {
	if !IsSupportedCurrency("USD") {
		t.Error("USD should be supported")
	}
	if !IsSupportedCurrency("usd") {
		t.Error("usd (lowercase) should be supported")
	}
	if IsSupportedCurrency("XYZ") {
		t.Error("XYZ should not be supported")
	}
}

func TestSupportedCurrencies(t *testing.T) {
	currencies := SupportedCurrencies()
	if len(currencies) == 0 {
		t.Error("Should return supported currencies")
	}

	// Check some common currencies are in the list
	found := make(map[string]bool)
	for _, c := range currencies {
		found[c] = true
	}

	if !found["USD"] {
		t.Error("USD should be in supported currencies")
	}
	if !found["EUR"] {
		t.Error("EUR should be in supported currencies")
	}
}
