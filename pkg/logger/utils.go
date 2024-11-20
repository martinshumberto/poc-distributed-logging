package logger

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	counter      uint64
	counterMutex sync.Mutex
)

func GenerateRequestID(prefix string) string {
	timestamp := time.Now().UTC().Format("20060102T150405.999999999")

	counterMutex.Lock()
	counter++
	currentCounter := counter
	counterMutex.Unlock()

	return fmt.Sprintf("%s-%s-%06d", prefix, timestamp, currentCounter)
}

func ExtractErrorDetails(err error) (*string, *string) {
	errParts := strings.Split(err.Error(), ":")
	if len(errParts) < 2 {
		return nil, nil
	}
	return &errParts[0], &errParts[1]
}
