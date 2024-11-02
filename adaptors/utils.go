package adaptors

import (
	"fmt"
	"junoplugin/models"
	"math/big"
)

func CombineFeltToBigInt(highFelt, lowFelt [32]byte) models.BigInt {
	combinedBytes := make([]byte, 64) // 32 bytes for highFelt and 32 bytes for lowFelt

	// Copy highFelt into the first 32 bytes
	copy(combinedBytes[0:32], highFelt[:])

	// Copy lowFelt into the next 32 bytes
	copy(combinedBytes[32:64], lowFelt[:])

	// Convert the combined bytes to a big.Int
	combinedInt := models.BigInt{Int: new(big.Int).SetBytes(combinedBytes)}

	return combinedInt
}

func FeltToBigInt(felt [32]byte) models.BigInt {

	byteData := make([]byte, 32)
	copy(byteData[:], felt[:])
	return models.BigInt{Int: new(big.Int).SetBytes(byteData)}
}
func FeltToHexString(felt [32]byte) string {

	combinedInt := models.BigInt{Int: new(big.Int).SetBytes(felt[:])}
	// Assuming `f.Value` holds the *big.Int representation of the felt
	return "0x" + combinedInt.Text(16)
}

func BigIntToHexString(f big.Int) string {

	// Assuming `f.Value` holds the *big.Int representation of the felt
	return "0x" + f.Text(16)
}

func DecimalStringToHexString(decimalString string) (string, error) {
	// Create a new big.Int and set its value from the decimal string
	num := new(big.Int)
	_, success := num.SetString(decimalString, 10)
	if !success {
		return "", fmt.Errorf("invalid decimal string")
	}

	// Convert the big.Int to a hexadecimal string
	hexString := num.Text(16)

	// Add "0x" prefix for clarity
	return "0x" + hexString, nil
}
