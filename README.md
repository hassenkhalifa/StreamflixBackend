# StreamFlix Backend

Backend API pour l'application StreamFlix, construit avec Go et Gin.

## Description

StreamFlix Backend fournit une API REST pour :
- Recherche et consultation de films et series TV (via TMDB)
- Streaming video via Real-Debrid (torrent debriding)
- Scraping de contenu depuis Zone Telechargement
- Generation de donnees mock pour le developpement

## Prerequis

- **Go 1.22+**
- Un token **TMDB** (The Movie Database)
- Un token **Real-Debrid** (optionnel, pour le streaming)

## Installation

```bash
git clone https://github.com/hassenkhalifa/StreamflixBackend.git
cd StreamflixBackend

# Copier et configurer les variables d'environnement
cp .env.example .env.dev
# Editer .env.dev avec vos tokens

# Installer les dependances
go mod tidy

# Lancer le serveur
go run cmd/api/main.go
```

## Variables d'environnement

| Variable | Description | Defaut |
|---|---|---|
| `PORT` | Port du serveur HTTP | `2000` |
| `GIN_MODE` | Mode Gin (debug/release/test) | `debug` |
| `ENVIRONMENT` | Environnement (development/production) | `development` |
| `REALDEBRID_TOKEN` | Token API Real-Debrid | - |
| `TMDB_TOKEN` | Bearer token TMDB | - |
| `CORS_ORIGINS` | Origines CORS autorisees (separees par ,) | `http://localhost:3000` |
| `RATE_LIMIT_PER_MINUTE` | Limite de requetes par minute par IP | `60` |
| `LOG_LEVEL` | Niveau de log (debug/info/warn/error) | `info` |
| `USER_AGENT` | User-Agent pour les requetes HTTP sortantes | `StreamFlix/1.0` |
| `HTTP_TIMEOUT_SECONDS` | Timeout HTTP sortant en secondes | `10` |
| `CACHE_TTL_MINUTES` | TTL du cache en minutes | `5` |

## Architecture

```
StreamflixBackend/
├── cmd/api/main.go              # Point d'entree (< 50 lignes, graceful shutdown)
├── internal/
│   ├── config/                  # Chargement configuration depuis .env
│   ├── domain/                  # Erreurs metier (ErrNotFound, ErrBadRequest...)
│   ├── http/
│   │   ├── router.go            # Definition des routes et groupes /api/v1
│   │   ├── handlers/            # Logique HTTP (TMDB, Real-Debrid, ZT, Player)
│   │   └── middleware/          # CORS, Security Headers, Rate Limiting, Recovery, Logger
│   ├── models/                  # DTOs et structures de donnees
│   ├── services/                # Couche service (a developper)
│   ├── cache/                   # Cache generique in-memory avec TTL
│   ├── clients/                 # Clients API externes (Real-Debrid)
│   └── utils/                   # Reponses JSON standardisees
├── docs/                        # Documentation API
├── .env.example                 # Template des variables d'environnement
├── go.mod
└── go.sum
```

## Endpoints API

### Health Check
```bash
curl http://localhost:2000/health
```

### Films (API v1)
```bash
# Films populaires
curl http://localhost:2000/api/v1/movies/popular

# Films tendance
curl http://localhost:2000/api/v1/movies/trending?time_window=day&language=fr-FR

# Films les mieux notes
curl http://localhost:2000/api/v1/movies/top-rated

# Details d'un film
curl http://localhost:2000/api/v1/movies/550

# Credits d'un film
curl http://localhost:2000/api/v1/movies/550/credits

# Films similaires
curl http://localhost:2000/api/v1/movies/550/similar

# Recherche de films
curl "http://localhost:2000/api/v1/movies/search?query=inception&language=fr-FR"

# Films par genre
curl "http://localhost:2000/api/v1/movies/by-genre?genre_id=28&page=1&language=fr-FR"

# Liste des genres
curl "http://localhost:2000/api/v1/movies/genres?language=fr-FR"
```

### Series TV (API v1)
```bash
# Series tendance
curl http://localhost:2000/api/v1/tv/trending?time_window=day&language=fr-FR

# Series populaires
curl http://localhost:2000/api/v1/tv/popular?language=fr-FR

# Series par genre
curl "http://localhost:2000/api/v1/tv/by-genre?genre_id=18&language=fr-FR"

# Details d'une serie
curl "http://localhost:2000/api/v1/tv/info?series_id=1399&language=fr-FR"

# Recherche de series
curl "http://localhost:2000/api/v1/tv/search?query=breaking+bad&language=fr-FR"
```

### Lecteur Video (API v1)
```bash
# Lancer un film
curl http://localhost:2000/api/v1/player/movie/550

# Lancer un episode de serie
curl http://localhost:2000/api/v1/player/series/1399/1/1
```

### Zone Telechargement (API v1)
```bash
# Recherche
curl "http://localhost:2000/api/v1/zt/search?category=films&query=inception&page=1"

# Recherche complete
curl "http://localhost:2000/api/v1/zt/search-all?category=films&query=inception"

# Infos basiques
curl http://localhost:2000/api/v1/zt/basic/films/12345

# Details complets
curl http://localhost:2000/api/v1/zt/films/12345
```

> Les anciens endpoints (sans prefixe `/api/v1`) restent disponibles pour la retrocompatibilite.

## Tests

```bash
# Lancer tous les tests
go test ./... -v

# Avec couverture
go test ./... -cover

# Linting
go vet ./...
```

## Format des reponses JSON

Toutes les reponses suivent le format standardise :

**Succes :**
```json
{
  "data": { ... },
  "error": null
}
```

**Erreur :**
```json
{
  "data": null,
  "error": {
    "code": "BAD_REQUEST",
    "message": "id invalide"
  }
}
```

## Securite

- CORS configurable via `CORS_ORIGINS` (pas de wildcard `*` en production)
- Rate limiting par IP (configurable via `RATE_LIMIT_PER_MINUTE`)
- Headers de securite : `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`
- Graceful shutdown avec timeout de 30s
- Logging structure JSON via `slog`
- Recovery middleware avec stack trace (jamais expose au client)
