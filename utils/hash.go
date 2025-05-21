package utils

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher şifre hashleme ve doğrulama işlemleri için interface
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) error
}

type BcryptHasher struct {
	cost int // bcrypt cost parametresi (4-31 arası)
}

func NewBcryptHasher(cost int) (*BcryptHasher, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, errors.New("geçersiz bcrypt cost değeri")
	}
	return &BcryptHasher{cost: cost}, nil
}

// HashPassword şifreyi bcrypt ile hashler
func (h *BcryptHasher) HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", errors.New("şifre boş olamaz")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("şifre hashlenirken hata: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash hashlenmiş şifre ile düz metni karşılaştırır
func (h *BcryptHasher) CheckPasswordHash(password, hash string) error {
	if password == "" || hash == "" {
		return errors.New("şifre veya hash boş olamaz")
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
