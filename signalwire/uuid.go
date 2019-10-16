package signalwire

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// GenUUIDv4 generates UUIDv4 (random), see: https://en.wikipedia.org/wiki/Universally_unique_identifier
func GenUUIDv4() (string, error) {
	u := make([]byte, 16)

	_, err := rand.Read(u)
	if err != nil {
		return "", err
	}

	// make sure that the 13th character is "4"
	u[6] = (u[6] | 0x40) & 0x4F
	// make sure that the 17th is "8", "9", "a", or "b"
	u[8] = (u[8] | 0x80) & 0xBF

	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])

	return strings.ToLower(uuid), nil
}
