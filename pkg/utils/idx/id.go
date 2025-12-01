package idx

import (
	"github.com/google/uuid"
)

func UUIDv4() string {
	return uuid.New().String()
}

func UUIDv7() (id string, err error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func UUIDv4Bytes() (id []byte, err error) {
	u := uuid.New()
	return u.MarshalBinary()
}

func UUIDv7Bytes() (id []byte, err error) {
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return u.MarshalBinary()
}

func UUIDFromBytes(b []byte) (id string, err error) {
	u, err := uuid.FromBytes(b)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func ValidateUUID(id string) error {
	return uuid.Validate(id)
}
