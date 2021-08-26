package switchblade

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
)

func RandomName() (string, error) {
	now := time.Now()
	timestamp := ulid.Timestamp(now)
	entropy := ulid.Monotonic(rand.Reader, 0)

	guid, err := ulid.New(timestamp, entropy)
	if err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("switchblade-%s", guid)), nil
}
