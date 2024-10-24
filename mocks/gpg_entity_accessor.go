package mocks

import (
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/mock"
)

type GpgEntityAccessorMock struct {
	mock.Mock
}

func (ea *GpgEntityAccessorMock) GetEntity(keyPassphrase string, signingKey string) (*openpgp.Entity, error) {
	args := ea.Called(keyPassphrase, signingKey)
	return nil, args.Error(0)
}
