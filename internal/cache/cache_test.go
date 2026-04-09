// Package cache fournit les tests unitaires pour le cache générique en mémoire.
//
// Les tests couvrent les opérations fondamentales du cache : insertion, récupération,
// expiration automatique, écrasement de valeurs, gestion de clés multiples,
// ainsi que le support de différents types de clés (string, int, struct).
package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCache_SetAndGet vérifie qu'une valeur insérée dans le cache peut être
// récupérée correctement avec sa clé correspondante.
func TestCache_SetAndGet(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("key1", 42)
	val, ok := c.Get("key1")

	assert.True(t, ok)
	assert.Equal(t, 42, val)
}

// TestCache_GetMissing vérifie que la récupération d'une clé inexistante retourne
// la valeur zéro du type et false pour indiquer l'absence de la clé.
func TestCache_GetMissing(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	val, ok := c.Get("nonexistent")

	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

// TestCache_Expiration vérifie que les entrées du cache expirent automatiquement
// après la durée de vie (TTL) configurée. Le test utilise un TTL court (50ms)
// puis attend 100ms pour confirmer l'expiration.
func TestCache_Expiration(t *testing.T) {
	c := New[string, string](50 * time.Millisecond)

	c.Set("key1", "value1")

	// Should be present immediately
	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", val)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	val, ok = c.Get("key1")
	assert.False(t, ok)
	assert.Equal(t, "", val)
}

// TestCache_Overwrite vérifie qu'une insertion avec une clé existante écrase
// correctement l'ancienne valeur avec la nouvelle.
func TestCache_Overwrite(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("key1", 1)
	c.Set("key1", 2)

	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 2, val)
}

// TestCache_MultipleKeys vérifie le stockage et la récupération de plusieurs
// entrées simultanées dans le cache. Utilise un pattern de tests pilotés par
// table (table-driven tests) pour vérifier chaque paire clé-valeur.
func TestCache_MultipleKeys(t *testing.T) {
	c := New[string, string](5 * time.Minute)

	c.Set("a", "alpha")
	c.Set("b", "beta")
	c.Set("c", "gamma")

	tests := []struct {
		key      string
		expected string
	}{
		{"a", "alpha"},
		{"b", "beta"},
		{"c", "gamma"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := c.Get(tt.key)
			assert.True(t, ok)
			assert.Equal(t, tt.expected, val)
		})
	}
}

// TestCache_IntKeys vérifie que le cache fonctionne correctement avec des clés
// de type entier (int), démontrant la généricité du cache.
func TestCache_IntKeys(t *testing.T) {
	c := New[int, string](5 * time.Minute)

	c.Set(1, "one")
	c.Set(2, "two")

	val, ok := c.Get(1)
	assert.True(t, ok)
	assert.Equal(t, "one", val)

	val, ok = c.Get(2)
	assert.True(t, ok)
	assert.Equal(t, "two", val)
}

// structKey est une structure de test utilisée comme clé composite pour vérifier
// que le cache supporte les clés de type struct (comparable).
type structKey struct {
	id   int
	lang string
}

// TestCache_StructKeys vérifie que le cache fonctionne avec des clés de type struct.
// Ce test est important car il valide le support des clés composites (id + langue),
// un cas d'usage courant pour le cache de contenu multilingue de StreamFlix.
func TestCache_StructKeys(t *testing.T) {
	c := New[structKey, string](5 * time.Minute)

	c.Set(structKey{1, "fr"}, "bonjour")
	c.Set(structKey{1, "en"}, "hello")

	val, ok := c.Get(structKey{1, "fr"})
	assert.True(t, ok)
	assert.Equal(t, "bonjour", val)

	val, ok = c.Get(structKey{1, "en"})
	assert.True(t, ok)
	assert.Equal(t, "hello", val)

	_, ok = c.Get(structKey{2, "fr"})
	assert.False(t, ok)
}
