package gpg_test

import (
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/joe-at-startupmedia/version-bump/v2/gpg"
	"github.com/joe-at-startupmedia/version-bump/v2/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGpg_GetPrivateKeyFails(t *testing.T) {
	erm := new(mocks.GpgEntityReader)
	defer erm.AssertExpectations(t)
	ea := &gpg.EntityAccessor{
		Reader: erm,
	}
	erm.On("GetPrivateKey", mock.Anything, mock.Anything).Return("", errors.New("gpg_test_get_private_key_error"))
	_, err := ea.GetEntity("", "")
	assert.ErrorContains(t, err, "gpg_test_get_private_key_error")
}

func TestGpg_ReadArmoredKeyRingPasses(t *testing.T) {
	erm := new(mocks.GpgEntityReader)
	defer erm.AssertExpectations(t)
	ea := &gpg.EntityAccessor{
		Reader: erm,
	}
	erm.On("GetPrivateKey", mock.Anything, mock.Anything).Return("my_private_key", nil)
	erm.On("ReadArmoredKeyRing", "my_private_key").Return(mockEntityList(), nil)
	_, err := ea.GetEntity("", "")
	assert.Nil(t, err)
}

func mockEntityList() openpgp.EntityList {
	entity, _ := openpgp.NewEntity("", "", "", nil)
	return openpgp.EntityList{entity}
}
