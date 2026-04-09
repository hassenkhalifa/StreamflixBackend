// Package cache fournit un cache en mémoire générique avec expiration par TTL.
//
// Ce package implémente un cache thread-safe utilisant les generics Go 1.18+.
// Les clés peuvent être de n'importe quel type comparable et les valeurs de
// n'importe quel type. Chaque entrée expire automatiquement après le TTL configuré.
//
// Ce cache est utilisé dans StreamFlix pour stocker temporairement les réponses
// des API externes (TMDB, Torrentio) et réduire la latence et le nombre d'appels réseau.
//
// Exemple d'utilisation :
//
//	c := cache.New[string, []Movie](30 * time.Minute)
//	c.Set("popular", movies)
//	if movies, ok := c.Get("popular"); ok {
//	    // Utiliser les données en cache
//	}
package cache

import (
	"sync"
	"time"
)

// entry représente une entrée du cache avec sa valeur et sa date d'expiration.
type entry[V any] struct {
	data      V         // Valeur stockée
	expiresAt time.Time // Date d'expiration de l'entrée
}

// Cache est un cache en mémoire générique thread-safe avec expiration par TTL.
//
// Il utilise un sync.RWMutex pour permettre des lectures concurrentes tout en
// protégeant les écritures. Les entrées expirées sont détectées à la lecture
// (lazy expiration) et ne sont pas supprimées proactivement.
//
// Paramètres de type :
//   - K : type de la clé (doit être comparable pour être utilisé comme clé de map)
//   - V : type de la valeur (peut être n'importe quel type)
type Cache[K comparable, V any] struct {
	mu    sync.RWMutex   // Mutex lecture/écriture pour la concurrence
	store map[K]entry[V] // Map interne stockant les entrées
	ttl   time.Duration  // Durée de vie des entrées
}

// New crée un nouveau cache avec le TTL spécifié.
//
// Paramètres :
//   - ttl : durée de vie des entrées en cache avant expiration
//
// Retourne un pointeur vers un Cache prêt à l'emploi.
//
// Exemple :
//
//	movieCache := cache.New[int, *Movie](60 * time.Minute)
func New[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	return &Cache[K, V]{
		store: make(map[K]entry[V]),
		ttl:   ttl,
	}
}

// Get récupère une valeur du cache par sa clé.
//
// Retourne la valeur et true si la clé existe et n'a pas expiré.
// Retourne la valeur zéro du type V et false si la clé n'existe pas
// ou si l'entrée a expiré (lazy expiration).
//
// Cette méthode utilise un RLock pour permettre des lectures concurrentes.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.store[key]
	if !ok || time.Now().After(e.expiresAt) {
		var zero V
		return zero, false
	}
	return e.data, true
}

// Set ajoute ou met à jour une entrée dans le cache.
//
// L'entrée expirera automatiquement après le TTL configuré lors de la création du cache.
// Si la clé existe déjà, la valeur et la date d'expiration sont remplacées.
//
// Cette méthode utilise un Lock exclusif pour protéger l'écriture.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = entry[V]{
		data:      value,
		expiresAt: time.Now().Add(c.ttl),
	}
}
