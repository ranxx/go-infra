package utils

import (
	"github.com/gofrs/uuid/v5"
)

// GenerateUUID generate a new UUIDv7
func GenerateUUID() string {
	uuid, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return uuid.String()
}
