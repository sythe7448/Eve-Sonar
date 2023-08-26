package ESI

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToke string `json:"refresh_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type CharacterInfo struct {
	CharacterID   int64  `json:"CharacterID"`
	CharacterName string `json:"CharacterName"`
	ExpiresOn     string `json:"ExpiresOn"`
}

type LocationInfo struct {
	ID int64 `json:"solar_system_id"`
}

const (
	LocalBaseURI = "http://localhost:8080"
	RedirectURI  = LocalBaseURI + "/callback"
	EveBaseURL   = "https://login.eveonline.com/oauth"
	AuthURL      = EveBaseURL + "/authorize"
	TokenURL     = EveBaseURL + "/token"
	VerifyURL    = EveBaseURL + "/verify"
	Scope        = "esi-location.read_location.v1"
	APIBaseURL   = "https://esi.evetech.net/latest"
)

var ClientID string
var ClientSecret string
var Character CharacterInfo
var Tokens TokenResponse
var Server *http.Server

func init() {
	err := godotenv.Load("ESI/.env") // Load environment variables from a .env file
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file:%s\n", err))
	}
	ClientID = os.Getenv("Client_ID")
	ClientSecret = os.Getenv("Client_SECRET")
}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", LoginUsingOAuth)
	mux.HandleFunc("/callback", getCode)
	Server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	if !isServerRunning(Server) {
		go func() {
			fmt.Println("Server started")
			if err := Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
		if err := Server.Shutdown(ctx); err != nil {
			fmt.Println("Error shutting down:", err)
		}
		fmt.Println("Server shut down")
	}
}

func LoginUsingOAuth(w http.ResponseWriter, r *http.Request) {
	conf := oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		RedirectURL:  RedirectURI,
		Scopes:       []string{Scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}
	authURL := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, authURL, http.StatusSeeOther)
}

func getCode(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		// Exchange authorization code for access token
		setAccessTokens(code)
		setCharacterInformationFromToken(Tokens.AccessToken)
		fmt.Fprintln(w, "Access Token Granted you can close this tab")
	} else {
		fmt.Fprintln(w, "No authorization code received.")
	}
}

func setAccessTokens(code string) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", RedirectURI)

	req, err := http.NewRequest("POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(ClientID, ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

func setCharacterInformationFromToken(accessToken string) {
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

func GetLocationId(accessToken string, characterID int64) (int64, error) {
	reqUrl := APIBaseURL + fmt.Sprintf("/characters/%d/location/", characterID)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var location LocationInfo
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		return 0, err
	}

	systemName := location.ID

	return systemName, nil
}
