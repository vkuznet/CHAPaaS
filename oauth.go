package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dghubble/gologin/v2/github"
	oauth2Login "github.com/dghubble/gologin/v2/oauth2"
	sessions "github.com/dghubble/sessions"
)

const (
	// here we keep names of cookies in our OAuth session
	sessionName     = "CHAP-App"
	sessionSecret   = "CHAP-Secret"
	sessionUserID   = "CHAP-UserID"
	sessionUserName = "CHAP-UserName"
	sessionToken    = "CHAP-Token"
	sessionProvider = "CHAP-Provider"
)

// sessionStore encodes and decodes session data stored in signed cookies
var sessionStore = sessions.NewCookieStore[any](sessions.DebugCookieConfig, []byte(sessionSecret), nil)

// issueSession issues a cookie session after successful provider login
func issueSession(provider string) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		var userName, userID, token string
		session := sessionStore.New(sessionName)
		ctx := req.Context()
		if t, err := oauth2Login.TokenFromContext(ctx); err == nil {
			token = t.AccessToken
		} else {
			log.Println("ERROR: fail to obtain OAuth2 token", err)
		}
		if provider == "github" {
			if user, err := github.UserFromContext(ctx); err == nil {
				userID = fmt.Sprintf("%v", *user.ID)
				userName = fmt.Sprintf("%v", *user.Login)
			} else {
				log.Println("ERROR: fail to obtain github credentials", err)
			}
		}
		session.Set(sessionProvider, provider)
		session.Set(sessionToken, token)
		session.Set(sessionUserID, userID)
		session.Set(sessionUserName, userName)
		if Config.Verbose > 0 {
			log.Printf("OAuth: provider %s user %s userID %s token %s", provider, userName, userID, token)
		}
		if err := session.Save(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// by default we will redirect to access end-point
		rpath := "/access"
		if req.URL != nil {
			// but if we get redirect query parameter we'll use it
			// to change redirect path
			redirect := req.URL.Query().Get("redirect")
			if redirect != "" {
				rpath = redirect
			}
		}
		if Config.Verbose > 0 {
			log.Printf("session redirect to '%s', request %+v", rpath, req)
		}
		http.Redirect(w, req, rpath, http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

/*
To check github token we can use the following API call:
curl -v -H "Authorization: Bearer $token" https://api.github.com/user
it will return something like this:
{
  "login": "UserName",
  "id": UserID,
  "type": "User",
  "name": "First Last name",
  "company": "Company Name",
  "location": "City, State",
  "bio": "Title associated with user",
}
*/

// UserData represents meta-data information about user
type UserData struct {
	Login    string
	ID       int
	Name     string
	Company  string
	Location string
	Bio      string
}

// helper function to get user data info
func githubTokenInfo(token string) (UserData, error) {
	var userData UserData
	// make HTTP call to github
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	uri := "https://api.github.com/user"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return userData, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rsp, err := client.Do(req)
	if err != nil {
		return userData, err
	}
	defer rsp.Body.Close()
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return userData, err
	}
	err = json.Unmarshal(data, &userData)
	return userData, err
}
