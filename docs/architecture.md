# Architecture — StreamFlix Backend

## Vue d'ensemble

StreamFlix Backend est une API REST construite avec Go et le framework Gin.
L'architecture suit les conventions Go standard avec une separation claire des responsabilites.

## Structure du projet

```
StreamflixBackend/
├── cmd/
│   └── api/
│       └── main.go                    # Point d'entree (wiring + graceful shutdown)
├── internal/
│   ├── config/
│   │   ├── config.go                  # Chargement config depuis .env via godotenv
│   │   └── config_test.go             # Tests unitaires
│   ├── domain/
│   │   ├── errors.go                  # Erreurs metier sentinelles
│   │   └── errors_test.go             # Tests
│   ├── http/
│   │   ├── router.go                  # Definition routes + groupes + handler factories
│   │   ├── router_test.go             # Tests d'integration des endpoints
│   │   ├── handlers/
│   │   │   ├── movies.go              # Logique TMDB films (Popular, Trending, Search...)
│   │   │   ├── series.go              # Logique TMDB series TV
│   │   │   ├── player.go              # Construction du VideoPlayer
│   │   │   ├── realdebrid.go          # Integration Real-Debrid (Unrestrict, MediaInfo...)
│   │   │   ├── real_debrid.go         # Client Real-Debrid (AddMagnet, SelectFiles...)
│   │   │   ├── streaming.go           # Service de streaming (workflow complet)
│   │   │   ├── torrentio.go           # Client Torrentio
│   │   │   ├── trancodevideo.go       # Transcodage FFmpeg/DASH
│   │   │   ├── zt.go                  # Parser Zone Telechargement
│   │   │   ├── categories.go          # Generation categories aleatoires
│   │   │   └── user.go                # Generation donnees utilisateur
│   │   └── middleware/
│   │       ├── cors.go                # CORS configurable par origines
│   │       ├── security.go            # Headers de securite HTTP
│   │       ├── ratelimit.go           # Rate limiter in-memory par IP
│   │       ├── logger.go              # Logging structure JSON (slog)
│   │       ├── recovery.go            # Recovery panic + log stack trace
│   │       └── middleware_test.go      # Tests unitaires middleware
│   ├── models/
│   │   ├── content_models.go          # DTOs films, series, recherche
│   │   ├── rd_models.go               # Modeles Real-Debrid et Torrentio
│   │   ├── player_models.go           # Modeles lecteur video
│   │   ├── zt_models.go               # Modeles Zone Telechargement
│   │   ├── user_models.go             # Modele liste utilisateur
│   │   ├── constants_models.go        # Genres, gradients, caches globaux
│   │   └── ffmpegs_models.go          # Modeles FFmpeg/ffprobe
│   ├── services/                      # Couche service (a developper)
│   ├── cache/
│   │   ├── cache.go                   # Cache generique type-safe avec TTL
│   │   └── cache_test.go              # Tests unitaires
│   ├── clients/
│   │   └── realdebrid/
│   │       ├── client.go              # Client HTTP Real-Debrid
│   │       └── types.go               # Types de reponse
│   └── utils/
│       ├── errors.go                  # Reponses JSON standardisees
│       └── errors_test.go             # Tests
├── docs/
│   ├── api-endpoints.md               # Documentation API
│   └── architecture.md                # Ce fichier
├── .env.example                       # Template variables d'environnement
├── .gitignore
├── go.mod
└── go.sum
```

## Flux d'une requete

```
Client HTTP
    │
    ▼
┌─────────────────┐
│   Gin Engine     │
│  (main.go)       │
└────────┬────────┘
         │
    ┌────▼────┐
    │ Middleware│  Recovery → Logger → Security → CORS → RateLimit
    └────┬────┘
         │
    ┌────▼────┐
    │  Router  │  /api/v1/movies/... , /api/v1/tv/... , /health
    └────┬────┘
         │
    ┌────▼────┐
    │ Handler  │  Validation params + appel service + reponse JSON
    └────┬────┘
         │
    ┌────▼────┐
    │ Service  │  Logique metier (TMDB, Real-Debrid, scraping)
    └────┬────┘
         │
    ┌────▼────┐
    │  Cache   │  Cache in-memory TTL (evite les appels API repetitifs)
    └─────────┘
```

## Middleware Stack

L'ordre des middleware est important :

1. **Recovery** — Capture les panics, log la stack trace, retourne 500 generique
2. **Logger** — Log structure JSON de chaque requete (method, path, status, latency)
3. **SecurityHeaders** — Ajoute X-Content-Type-Options, X-Frame-Options, etc.
4. **CORS** — Verifie l'origine et gere les preflight OPTIONS
5. **RateLimit** — Token bucket par IP, bypass pour /health et OPTIONS

## Cache

Le systeme de cache est generique et type-safe grace aux generics Go :

```go
cache.New[KeyType, ValueType](ttl time.Duration)
```

Chaque endpoint TMDB a son propre cache avec un TTL adapte :
- Films populaires : 30 min
- Films tendance : 15 min
- Details film : 60 min
- Genres : 24h

## Gestion d'erreurs

### Reponses standardisees

Toutes les reponses suivent le format :
```json
{ "data": ..., "error": null }
{ "data": null, "error": { "code": "NOT_FOUND", "message": "..." } }
```

### Erreurs sentinelles

Le package `domain` definit des erreurs metier :
- `ErrNotFound` — Ressource non trouvee
- `ErrBadRequest` — Requete invalide
- `ErrUnauthorized` — Non autorise
- `ErrRateLimited` — Limite depassee
- `ErrInternal` — Erreur interne

### Principe : ne jamais exposer les details internes

Les erreurs techniques (stack traces, messages d'API externe) sont loggees
cote serveur mais jamais exposees au client.

## Securite

| Mesure | Implementation |
|---|---|
| CORS | Origines configurables via `CORS_ORIGINS` |
| Rate Limiting | Token bucket par IP, configurable |
| Headers HTTP | X-Content-Type-Options, X-Frame-Options, X-XSS-Protection |
| Secrets | Variables d'environnement, jamais dans le code |
| Panic Recovery | Middleware custom avec log JSON |
| Graceful Shutdown | Signal SIGINT/SIGTERM, timeout 30s |

## Tests

Couverture par package :
- `internal/cache` : 100%
- `internal/config` : 85%
- `internal/utils` : 100%
- `internal/http/middleware` : 82%
- `internal/http` (router) : 38% (limite par les appels API externes)
- `internal/domain` : 100%

Commande :
```bash
go test ./... -v -cover
```
