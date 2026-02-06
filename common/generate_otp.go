package common

import (
	"crypto/rand"
	"io"

	"github.com/cockroachdb/errors"
)

func GenerateOTP(length int) (string, error) {
	const chars = "0123456789"

	buffer := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, buffer)
	if err != nil {
		return "", errors.Wrap(err, "failed to read random bytes")
	}

	for i := range length {
		buffer[i] = chars[int(buffer[i])%len(chars)]
	}

	return string(buffer), nil
}
