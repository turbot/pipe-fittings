package utils

import (
	"crypto/hmac"
	"crypto/md5" //nolint:gosec // we are using md5 for hashing and distributing the string, not for security
	"encoding/hex"
	"strconv"
)

func DistributedStringIndex(message string, secret string, numberRange int64) (int64, error) {
	key := []byte(secret)
	h := hmac.New(md5.New, key)
	h.Write([]byte(message))
	hexString := hex.EncodeToString(h.Sum(nil))
	last3charsHex := hexString[len(hexString)-3:]
	intValue, err := strconv.ParseInt(last3charsHex, 16, 64)
	if err != nil {
		return 0, err
	}

	index := numberRange * intValue / 4096

	return index, nil
}
