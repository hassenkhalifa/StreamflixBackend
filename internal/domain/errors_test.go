package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrBadRequest", ErrBadRequest},
		{"ErrUnauthorized", ErrUnauthorized},
		{"ErrRateLimited", ErrRateLimited},
		{"ErrInternal", ErrInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
			assert.NotEmpty(t, tt.err.Error())
		})
	}
}

func TestWrappedErrors(t *testing.T) {
	wrapped := fmt.Errorf("movie 123: %w", ErrNotFound)
	assert.True(t, errors.Is(wrapped, ErrNotFound))
	assert.False(t, errors.Is(wrapped, ErrBadRequest))
}
