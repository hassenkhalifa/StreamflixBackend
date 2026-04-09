// Package realdebrid fournit un client HTTP pour l'API Real-Debrid.
//
// Ce package implémente un client bas-niveau pour interagir avec l'API REST
// Real-Debrid (https://api.real-debrid.com/rest/1.0). Il gère l'authentification
// par token Bearer, les timeouts HTTP et le parsing des réponses JSON.
//
// Le client est configuré avec des paramètres de transport optimisés :
//   - Timeout global de 20 secondes
//   - Connection pooling (100 max idle, 10 par host)
//   - Keep-alive de 30 secondes
//
// Exemple d'utilisation :
//
//	client := realdebrid.NewClient("mon_token_api")
//	resp, err := client.UnrestrictLink("https://example.com/file", "", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(resp.Download)
package realdebrid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// baseURL est l'URL de base de l'API REST Real-Debrid (version 1.0).
const baseURL = "https://api.real-debrid.com/rest/1.0"

// Client est le client HTTP pour interagir avec l'API Real-Debrid.
//
// Il encapsule un http.Client configuré avec des paramètres de transport optimisés
// et un token d'authentification Bearer. Le client est thread-safe et peut être
// partagé entre plusieurs goroutines.
type Client struct {
	httpClient *http.Client
	token      string
}

// NewClient crée un nouveau client Real-Debrid avec le token d'API fourni.
//
// Le client HTTP interne est configuré avec :
//   - Un timeout global de 20 secondes.
//   - Un timeout de connexion de 10 secondes.
//   - Un keep-alive de 30 secondes.
//   - Un pool de connexions idle (100 max, 10 par host, timeout 90s).
//   - Le support du proxy depuis les variables d'environnement.
//
// Paramètres :
//   - token : clé API Real-Debrid pour l'authentification Bearer.
//
// Retourne un pointeur vers le Client initialisé.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// UnrestrictLink dé-restreint un lien hébergeur via l'API Real-Debrid.
//
// Elle envoie une requête POST vers /unrestrict/link avec le lien à dé-restreindre.
// La réponse contient l'URL de téléchargement direct, le nom du fichier, sa taille, etc.
// Le corps de la réponse est limité à 2 Mo pour éviter les dépassements mémoire.
//
// Paramètres :
//   - link : URL du fichier hébergé à dé-restreindre.
//   - password : mot de passe du lien (chaîne vide si aucun).
//   - remote : si non nil, indique l'identifiant de téléchargement distant Real-Debrid.
//
// Retourne un UnrestrictResponse avec l'URL de téléchargement direct, ou une erreur.
// En cas d'erreur HTTP, tente de parser un ErrorResponse Real-Debrid avant de
// retourner une erreur générique.
func (c *Client) UnrestrictLink(link string, password string, remote *int) (*UnrestrictResponse, error) {
	form := url.Values{}
	form.Set("link", link)
	if password != "" {
		form.Set("password", password)
	}
	if remote != nil {
		form.Set("remote", fmt.Sprintf("%d", *remote))
	}

	req, err := http.NewRequest("POST", baseURL+"/unrestrict/link", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB cap

	if resp.StatusCode != http.StatusOK {
		var rdErr ErrorResponse
		if json.Unmarshal(body, &rdErr) == nil && rdErr.Message != "" {
			return nil, rdErr
		}
		return nil, fmt.Errorf("realdebrid: http %d: %s", resp.StatusCode, string(body))
	}

	var out UnrestrictResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if out.Download == "" {
		return nil, errors.New("realdebrid: missing download url")
	}

	return &out, nil
}

// DownloadFile télécharge un fichier depuis une URL de téléchargement direct.
//
// Cette méthode effectue une requête GET vers l'URL fournie (typiquement obtenue
// via UnrestrictLink) et retourne la réponse HTTP brute dont le Body contient
// le flux de données du fichier.
//
// IMPORTANT : l'appelant est responsable de fermer resp.Body après utilisation.
//
// Paramètres :
//   - downloadURL : URL de téléchargement direct (obtenue via UnrestrictResponse.Download).
//
// Retourne la réponse HTTP (status 200) ou une erreur si le téléchargement échoue.
func (c *Client) DownloadFile(downloadURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: http %d", resp.StatusCode)
	}

	return resp, nil
}
