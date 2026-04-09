// Package domain contient les types et erreurs du domaine métier de StreamFlix.
//
// Ce package définit les erreurs sentinelles utilisées dans toute l'application
// pour représenter les cas d'erreur courants de manière cohérente. Ces erreurs
// permettent aux handlers HTTP de déterminer le code de statut HTTP approprié
// via errors.Is().
//
// Exemple d'utilisation :
//
//	if errors.Is(err, domain.ErrNotFound) {
//	    utils.NotFound(c, "film non trouvé")
//	}
package domain

import "errors"

// Erreurs sentinelles pour la logique métier.
//
// Ces erreurs sont utilisées comme base pour le wrapping d'erreurs dans les services
// et permettent aux handlers de déterminer le code HTTP approprié.
//
// Exemple de wrapping :
//
//	return fmt.Errorf("film %d: %w", movieID, domain.ErrNotFound)
var (
	// ErrNotFound indique qu'une ressource demandée n'existe pas (HTTP 404).
	ErrNotFound = errors.New("resource not found")
	// ErrBadRequest indique que la requête est invalide ou malformée (HTTP 400).
	ErrBadRequest = errors.New("bad request")
	// ErrUnauthorized indique que l'authentification est requise ou invalide (HTTP 401).
	ErrUnauthorized = errors.New("unauthorized")
	// ErrRateLimited indique que le client a dépassé le nombre de requêtes autorisées (HTTP 429).
	ErrRateLimited = errors.New("rate limited")
	// ErrInternal indique une erreur serveur interne inattendue (HTTP 500).
	ErrInternal = errors.New("internal server error")
)
