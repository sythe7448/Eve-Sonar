package ESI

import (
	"context"
	"encoding/json"
	"fmt"
	pkce "github.com/nirasan/go-oauth-pkce-code-verifier"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type CharacterInfo struct {
	CharacterID   int64  `json:"CharacterID"`
	CharacterName string `json:"CharacterName"`
	ExpiresOn     string `json:"ExpiresOn"`
}

type LocationInfo struct {
	ID int `json:"solar_system_id"`
}

const (
	ClientID     = "667fb5ef212f4fffb9383ad23ce4050e"
	LocalBaseURI = "http://localhost:8080"
	RedirectURI  = LocalBaseURI + "/callback"
	EveBaseURL   = "https://login.eveonline.com/oauth"
	AuthURL      = EveBaseURL + "/authorize"
	TokenURL     = EveBaseURL + "/token"
	VerifyURL    = EveBaseURL + "/verify"
	Scope        = "esi-location.read_location.v1"
	APIBaseURL   = "https://esi.evetech.net/latest"
)

var codeChallenge string
var codeVerifier string
var Character CharacterInfo
var Tokens TokenResponse
var server *http.Server

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", loginUsingOAuth)
	mux.HandleFunc("/callback", getCode)
	server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if !isServerRunning(server) {
		go func() {
			fmt.Println("server started")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Println("Error:", err)
			}
		}()

		// Once we have the access token shutdown server
		for len(Tokens.AccessToken) == 0 {
			time.Sleep(1 * time.Second)
		}

		fmt.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			fmt.Println("Error shutting down:", err)
		}
		fmt.Println("server shut down")
	}
}

func GetLocationId(accessToken string, characterID int64) (string, error) {
	if characterID == 0 {
		return "", nil
	}
	reqUrl := APIBaseURL + fmt.Sprintf("/characters/%d/location/", characterID)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var location LocationInfo
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return "", err
	}

	systemName := strconv.Itoa(location.ID)

	return systemName, nil
}

func loginUsingOAuth(w http.ResponseWriter, r *http.Request) {
	// Generate a random code verifier
	v, err := pkce.CreateCodeVerifier()
	if err != nil {
		fmt.Println("Error generating code verifier:", err)
		return
	}

	// Verifier String
	codeVerifier = v.String()
	// Create code_challenge with S256 method
	codeChallenge = v.CodeChallengeS256()

	conf := oauth2.Config{
		ClientID:    ClientID,
		RedirectURL: RedirectURI,
		Scopes:      []string{Scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}
	authURL := conf.AuthCodeURL(
		"state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

func refreshTokens() {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", Tokens.RefreshToken)
	data.Set("client_id", ClientID)
	req, err := http.NewRequest("POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "login.eveonline.com")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&Tokens); err != nil {
		panic(err)
	}
}

func getCode(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		// Exchange authorization code for access token
		setAccessTokens(code)
		setCharacterInformation(Tokens.AccessToken)
		fmt.Fprintln(w, "Access Token Granted you can close this tab")
	} else {
		fmt.Fprintln(w, "No authorization code received.")
	}
}

func setAccessTokens(code string) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", ClientID)
	data.Set("code_verifier", codeVerifier)

	req, err := http.NewRequest("POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Host", "login.eveonline.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&Tokens); err != nil {
		panic(err)
	}

	// subroutine to refresh access token every 19 minutes
	go func() {
		for range time.Tick(time.Minute * 19) {
			refreshTokens()
		}
	}()
}

// don't store this information in database
func setCharacterInformation(accessToken string) {
	req, err := http.NewRequest("GET", VerifyURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&Character); err != nil {
		panic(err)
	}
}

func isServerRunning(server *http.Server) bool {
	// Send a request to the server and check if it's responding
	client := &http.Client{Timeout: time.Second}
	_, err := client.Get("http://localhost" + server.Addr)
	return err == nil
}
