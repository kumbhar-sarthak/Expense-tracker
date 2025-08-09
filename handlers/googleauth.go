package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)
var Store *sessions.CookieStore
var GoogleAuth *oauth2.Config

func init() {
	
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	secret := os.Getenv("STORE_SEC")
	if secret == "" {
		log.Fatal("SESSION_KEY not set in .env")
	}

	Store = sessions.NewCookieStore([]byte(secret))
 	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
	}

	GoogleAuth = &oauth2.Config{
		ClientID:     os.Getenv("API_KEY"),
		ClientSecret: os.Getenv("API_KEY_SEC"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}


func HandleGoogleLogin(w http.ResponseWriter, r *http.Request){
	url := GoogleAuth.AuthCodeURL("state-token",oauth2.AccessTypeOffline)
	http.Redirect(w,r,url,http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request)  {
	code := r.URL.Query().Get("code")

	token, err:= GoogleAuth.Exchange(context.Background(),code)
	if err != nil {
		log.Fatal("Error while reciving the token")
	}

	client := GoogleAuth.Client(context.Background(),token)
	res, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")

	if err != nil {
		log.Fatal("Error while getting client info",http.StatusInternalServerError)
	}
	defer res.Body.Close()

	var user struct {
		GoogleID string `json:"id"`
		Email string `json:"email"`
		Name string `json:"name"`
		Profile string `json:"picture"`
	}

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		http.Error(w,"Faild to decode the user infor",http.StatusInternalServerError)
		return
	}

	session, err := Store.Get(r, "user-session")
	if err != nil {
		log.Fatal("unable to laod the sesion data",http.StatusInternalServerError)
	}

	session.Values["google_id"] = user.GoogleID
	session.Values["name"] = user.Name
	session.Values["email"] = user.Email
	session.Values["picture"] = user.Profile
	err = session.Save(r, w)

	if err != nil {
		log.Printf("Session save error: %v", err)
	}

	http.Redirect(w,r,"/dashboard",http.StatusSeeOther)
}