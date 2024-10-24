package mocks

import (
	"bytes"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

type ReleaseGetterMock struct {
	mock.Mock
}

func (rg *ReleaseGetterMock) Get(url string) (*http.Response, error) {
	args := rg.Called(url)
	resp := &http.Response{
		StatusCode: args.Int(0),
		Body:       io.NopCloser(bytes.NewReader([]byte(args.String(1)))),
	}
	return resp, args.Error(2)
}
