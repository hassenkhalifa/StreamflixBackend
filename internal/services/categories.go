// Package services est destiné à contenir la logique métier de l'application StreamFlix.
//
// Ce package est actuellement en cours de développement. La logique métier se trouve
// temporairement dans le package handlers et sera progressivement migrée ici
// pour respecter l'architecture en couches (handlers → services → clients).
//
// Structure prévue :
//   - categories.go : service de gestion des catégories de contenu
//   - movies.go : service de gestion des films (TMDB, cache, transformation)
//   - realdebrid.go : service de débridage (Real-Debrid, workflow torrent → stream)
//   - user.go : service de gestion des données utilisateur
package services
