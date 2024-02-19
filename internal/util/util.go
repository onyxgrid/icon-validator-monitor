package util

import (
	"math/big"
	"strconv"
)

// FormatIconNumber formats a 0xBignumber, divides it by 10^18, converts it to a float with 2 decimal places, and returns it as a string
func FormatIconNumber(n *big.Int) string {
	// Remove the "0x" prefix
	// n = n[2:]

	// Convert the hex string to a big.Int
	// bigInt, _ := new(big.Int).SetString(n, 16)

	// Divide by 10^18
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	result := new(big.Float).Quo(new(big.Float).SetInt(n), new(big.Float).SetInt(divisor))

	// Convert to float with 2 decimal places
	resultFloat, _ := result.Float64()
	resultFloatString := strconv.FormatFloat(resultFloat, 'f', 2, 64)

	return resultFloatString
}