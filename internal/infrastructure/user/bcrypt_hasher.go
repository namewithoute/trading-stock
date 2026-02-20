package user

import "golang.org/x/crypto/bcrypt"

// BcryptHasher implements domain/user.PasswordHasher using the bcrypt algorithm.
type BcryptHasher struct{}

// NewBcryptHasher creates a new BcryptHasher.
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

// Hash generates a bcrypt hash from the given plain-text password.
func (b *BcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare checks whether the plain-text password matches the bcrypt hash.
func (b *BcryptHasher) Compare(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
