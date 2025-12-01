package crypto

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var DefaultPasswordHasher *PasswordHasher

func init() {
	DefaultPasswordHasher = NewPasswordHasher(bcrypt.DefaultCost)
}

type PasswordHasher struct {
	cost int
}

func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < int(bcrypt.MinCost) || cost > int(bcrypt.MaxCost) {
		cost = int(bcrypt.DefaultCost)
	}
	return &PasswordHasher{cost: cost}
}

func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("empty password")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (ph *PasswordHasher) VerifyPassword(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, errors.New("empty password or hash")
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
