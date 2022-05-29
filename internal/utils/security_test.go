package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptPassword(t *testing.T) {
	hash, err := EncryptPassword("123456789")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 60)
}

func TestVerifyPassword(t *testing.T) {
	password := "123456789"
	hash, err := EncryptPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 60)

	assert.NoError(t, VerifyPassword(hash, password))

	assert.Error(t, VerifyPassword(hash, "12345678"))
	assert.Error(t, VerifyPassword(password, password))
	assert.Error(t, VerifyPassword(hash, hash))
}
