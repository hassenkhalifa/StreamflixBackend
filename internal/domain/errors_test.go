// Package domain fournit les tests unitaires pour les erreurs sentinelles du domaine.
//
// Les tests vérifient que chaque erreur sentinelle est correctement définie et
// que le mécanisme d'encapsulation (wrapping) des erreurs fonctionne avec errors.Is.
package domain

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSentinelErrors vérifie que toutes les erreurs sentinelles du domaine
// (ErrNotFound, ErrBadRequest, ErrUnauthorized, ErrRateLimited, ErrInternal)
// sont non-nil et possèdent un message non vide. Utilise un pattern de tests
// pilotés par table pour itérer sur chaque erreur.
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

// TestWrappedErrors vérifie que les erreurs sentinelles peuvent être encapsulées
// avec fmt.Errorf et le verbe %w, et que errors.Is les détecte correctement.
// Confirme aussi qu'une erreur encapsulée avec ErrNotFound n'est pas confondue
// avec ErrBadRequest.
func TestWrappedErrors(t *testing.T) {
	wrapped := fmt.Errorf("movie 123: %w", ErrNotFound)
	assert.True(t, errors.Is(wrapped, ErrNotFound))
	assert.False(t, errors.Is(wrapped, ErrBadRequest))
}
