# Documentation API — StreamFlix Backend

## Format des reponses

Toutes les reponses suivent le format standardise :

```json
// Succes
{ "data": ..., "error": null }

// Erreur
{ "data": null, "error": { "code": "ERROR_CODE", "message": "..." } }
```

### Codes d'erreur

| Code HTTP | Code erreur | Description |
|---|---|---|
| 200 | - | Succes |
| 400 | `BAD_REQUEST` | Parametre invalide ou manquant |
| 404 | `NOT_FOUND` | Ressource non trouvee |
| 429 | `RATE_LIMITED` | Trop de requetes |
| 500 | `INTERNAL_ERROR` | Erreur interne du serveur |

---

## Health Check

### `GET /health`

Verifie l'etat du serveur.

**Reponse :**
```json
{
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": "2h30m15s"
  },
  "error": null
}
```

---

## Films

### `GET /api/v1/movies/popular`

Retourne les films populaires depuis TMDB.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `language` | query | `fr-FR` | Langue des resultats |

```bash
curl http://localhost:2000/api/v1/movies/popular
```

---

### `GET /api/v1/movies/top-rated`

Retourne les films les mieux notes.

```bash
curl http://localhost:2000/api/v1/movies/top-rated
```

---

### `GET /api/v1/movies/trending`

Retourne les films tendance.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `time_window` | query | `day` | `day` ou `week` |
| `language` | query | `fr-FR` | Langue |

```bash
curl "http://localhost:2000/api/v1/movies/trending?time_window=week&language=fr-FR"
```

---

### `GET /api/v1/movies/:id`

Retourne les details d'un film.

| Parametre | Type | Description |
|---|---|---|
| `id` | path | ID TMDB du film |

```bash
curl http://localhost:2000/api/v1/movies/550
```

**Erreurs possibles :**
- `400 BAD_REQUEST` : id invalide (non numerique)

---

### `GET /api/v1/movies/:id/credits`

Retourne les credits (casting, realisateur, producteur) d'un film.

| Parametre | Type | Description |
|---|---|---|
| `id` | path | ID TMDB du film (> 0) |

```bash
curl http://localhost:2000/api/v1/movies/550/credits
```

---

### `GET /api/v1/movies/:id/similar`

Retourne les films similaires.

| Parametre | Type | Description |
|---|---|---|
| `id` | path | ID TMDB du film (> 0) |

```bash
curl http://localhost:2000/api/v1/movies/550/similar
```

---

### `GET /api/v1/movies/search`

Recherche avancee de films.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `query` | query | - | Terme de recherche |
| `genres` | query | - | IDs de genres (CSV) |
| `years` | query | - | Annees (CSV) |
| `sort_by` | query | `popularity.desc` | Tri |
| `page` | query | `1` | Page |
| `language` | query | `fr-FR` | Langue |
| `rating` | query | `0` | Note minimale |

```bash
curl "http://localhost:2000/api/v1/movies/search?query=inception&language=fr-FR&page=1"
```

---

### `GET /api/v1/movies/by-genre`

Films filtres par genre.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `genre_id` | query | `28` | ID du genre |
| `page` | query | `1` | Page |
| `language` | query | `fr-FR` | Langue |

```bash
curl "http://localhost:2000/api/v1/movies/by-genre?genre_id=28&page=1"
```

---

### `GET /api/v1/movies/genres`

Liste des genres de films.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `language` | query | `fr-FR` | Langue |

```bash
curl "http://localhost:2000/api/v1/movies/genres?language=fr-FR"
```

---

### `GET /api/v1/movies/:id/imdb`

Retourne l'ID IMDB d'un film.

```bash
curl http://localhost:2000/api/v1/movies/550/imdb
```

---

## Series TV

### `GET /api/v1/tv/trending`

Series TV tendance.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `time_window` | query | `day` | `day` ou `week` |
| `language` | query | `fr-FR` | Langue |
| `page` | query | `1` | Page |

```bash
curl "http://localhost:2000/api/v1/tv/trending?time_window=day&language=fr-FR"
```

---

### `GET /api/v1/tv/popular`

Series TV populaires.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `language` | query | `fr-FR` | Langue |
| `page` | query | `1` | Page |

```bash
curl "http://localhost:2000/api/v1/tv/popular?language=fr-FR"
```

---

### `GET /api/v1/tv/by-genre`

Series filtrees par genre.

| Parametre | Type | Defaut | Description |
|---|---|---|---|
| `genre_id` | query | `18` | ID du genre |
| `page` | query | `1` | Page |
| `language` | query | `fr-FR` | Langue |

```bash
curl "http://localhost:2000/api/v1/tv/by-genre?genre_id=18&page=1"
```

---

### `GET /api/v1/tv/info`

Details complets d'une serie (saisons, episodes, casting).

| Parametre | Type | Description |
|---|---|---|
| `series_id` | query | ID TMDB de la serie (**obligatoire**) |
| `language` | query | Langue (defaut: `fr-FR`) |

```bash
curl "http://localhost:2000/api/v1/tv/info?series_id=1399&language=fr-FR"
```

**Erreurs possibles :**
- `400 BAD_REQUEST` : series_id manquant ou invalide

---

### `GET /api/v1/tv/search`

Recherche de series.

| Parametre | Type | Description |
|---|---|---|
| `query` | query | Terme de recherche (**obligatoire**) |
| `language` | query | Langue (defaut: `fr-FR`) |
| `page` | query | Page (defaut: `1`) |

```bash
curl "http://localhost:2000/api/v1/tv/search?query=breaking+bad&language=fr-FR"
```

**Erreurs possibles :**
- `400 BAD_REQUEST` : query manquant

---

## Lecteur Video

### `GET /api/v1/player/movie/:id`

Lance le workflow complet pour obtenir un stream de film :
TMDB -> IMDB ID -> Torrentio -> Real-Debrid -> Stream URL

| Parametre | Type | Description |
|---|---|---|
| `id` | path | ID TMDB du film |

```bash
curl http://localhost:2000/api/v1/player/movie/550
```

---

### `GET /api/v1/player/series/:id/:season/:episode`

Lance le workflow pour un episode de serie.

| Parametre | Type | Description |
|---|---|---|
| `id` | path | ID TMDB de la serie |
| `season` | path | Numero de saison |
| `episode` | path | Numero d'episode |

```bash
curl http://localhost:2000/api/v1/player/series/1399/1/1
```

---

## Zone Telechargement

### `GET /api/v1/zt/search`

Recherche sur Zone Telechargement.

| Parametre | Type | Description |
|---|---|---|
| `category` | query | Categorie (`films`, `series`) (**obligatoire**) |
| `query` | query | Terme de recherche (**obligatoire**) |
| `page` | query | Page (defaut: `1`) |

```bash
curl "http://localhost:2000/api/v1/zt/search?category=films&query=inception&page=1"
```

---

### `GET /api/v1/zt/search-all`

Recherche sur toutes les pages.

| Parametre | Type | Description |
|---|---|---|
| `category` | query | Categorie (**obligatoire**) |
| `query` | query | Terme de recherche (**obligatoire**) |

```bash
curl "http://localhost:2000/api/v1/zt/search-all?category=films&query=inception"
```

---

### `GET /api/v1/zt/basic/:category/:id`

Infos basiques d'un contenu ZT.

```bash
curl http://localhost:2000/api/v1/zt/basic/films/12345
```

---

### `GET /api/v1/zt/:category/:id`

Details complets d'un contenu ZT.

```bash
curl http://localhost:2000/api/v1/zt/films/12345
```

---

## Donnees utilisateur

### `GET /api/v1/user/list`

Retourne la liste des elements utilisateur (favoris, historique, watchlist).

```bash
curl http://localhost:2000/api/v1/user/list
```

---

## Endpoints Legacy

Les anciens endpoints restent disponibles pour la retrocompatibilite :

| Ancien | Nouveau equivalent |
|---|---|
| `/popularMovies` | `/api/v1/movies/popular` |
| `/getTopRatedMovies` | `/api/v1/movies/top-rated` |
| `/getTrendingMovies` | `/api/v1/movies/trending` |
| `/movie/:id` | `/api/v1/movies/:id` |
| `/movie/:id/credits` | `/api/v1/movies/:id/credits` |
| `/movie/:id/similar` | `/api/v1/movies/:id/similar` |
| `/moviesbygenre` | `/api/v1/movies/by-genre` |
| `/getMovieGenreList` | `/api/v1/movies/genres` |
| `/searchMovies` | `/api/v1/movies/search` |
| `/getTrendingTV` | `/api/v1/tv/trending` |
| `/getTVShowsByGenre` | `/api/v1/tv/by-genre` |
| `/getPopularTVShows` | `/api/v1/tv/popular` |
| `/getTVInfo` | `/api/v1/tv/info` |
| `/searchTV` | `/api/v1/tv/search` |
| `/videoPlayer/:id` | `/api/v1/player/movie/:id` |
| `/videoPlayer/:id/:s/:e` | `/api/v1/player/series/:id/:s/:e` |
| `/zt/*` | `/api/v1/zt/*` |

---

## Headers de securite

Chaque reponse inclut :

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

## Rate Limiting

- Limite configurable via `RATE_LIMIT_PER_MINUTE` (defaut: 60/min par IP)
- Headers de reponse : `X-RateLimit-Limit`, `X-RateLimit-Remaining`
- Reponse 429 avec header `Retry-After` si la limite est depassee
- Le endpoint `/health` est exempt du rate limiting
