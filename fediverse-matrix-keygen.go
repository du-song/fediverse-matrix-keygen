package main

import (
	"context"
	"strings"
	"fmt"
	"log"
	"flag"
	"time"
	"strconv"
	"net/http"
	"math/rand"

	"github.com/mattn/go-mastodon"
	"maunium.net/go/mautrix"
	//_ "github.com/acobaugh/go-loghttp/global"
)

const appName           = "Fediverse Matrix KeyGen" 
const redirectURLSuffix = "/.verified"
var lettersForPassword  = []rune("!#%+23456789:=?@ABCDEFGHJKLMNPRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
var fediverseServer     = flag.String("f",  "", "Mastodon or Pleroma server name")
var matrixServer        = flag.String("m",  "", "Matrix server name")
var matrixAdminUser     = flag.String("mu", "", "Matrix server admin user")
var matrixAdminPassword = flag.String("mp", "", "Matrix server admin password")
var serverBaseURL       = flag.String("u",  "http://localhost:8848", "URL to access this service")
var serverPort          = flag.Int("p", 8848, "Port this service should listen on")
var mastodonApp *mastodon.Application
var matrixClient *mautrix.Client

func randomPassword(n int) string {
	b := make([]rune, n)
	for i := range b {
			b[i] = lettersForPassword[rand.Intn(len(lettersForPassword))]
	}
	return string(b)
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	if (!strings.HasSuffix(r.URL.Path, redirectURLSuffix)) {
		fmt.Fprintf(w, "<h1>Matrix Account for %s Users</h1>" + 
			"<p>Click <a href=\"%s\">here</a> to verify your username on %s, then create a new Matrix account <b>@username:%s</b> or reset its password if the account already exists.</p>" + 
			"<p><i>Please be sure to back up your Matrix encryption keys and passphrases, password reset will also logout your current Matrix sessions.</i></p>", *fediverseServer,
			mastodonApp.AuthURI, *fediverseServer, *matrixServer)
		return
	}

	ctx := r.Context()
	mastodonUserAccessToken := r.FormValue("code")
	c := mastodon.NewClient(&mastodon.Config{
		Server:       "https://" + *fediverseServer,
		ClientID:     mastodonApp.ClientID,
		ClientSecret: mastodonApp.ClientSecret,
	})
	err := c.AuthenticateToken(ctx, mastodonUserAccessToken, mastodonApp.RedirectURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	mastodonUserAccount, err := c.GetAccountCurrentUser(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	/*
		https://github.com/matrix-org/synapse/blob/develop/docs/admin_api/README.rst
		https://github.com/matrix-org/synapse/blob/develop/docs/admin_api/user_admin_api.rst#create-or-modify-account
	*/
	matrixUserID := strings.ToLower("@" + mastodonUserAccount.Username + ":" + *matrixServer)
	matrixUserPassword := randomPassword(9)
	urlCreateUser := matrixClient.BuildBaseURL("_synapse", "admin", "v2", "users", matrixUserID)
	s := struct {
		Password string `json:"password"`
		Deactivated bool `json:"deactivated"`
	}{matrixUserPassword, false}
	_, err = matrixClient.MakeRequest("PUT", urlCreateUser, &s, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "<h1>Success</h1><p>Created or updated Matrix account <b>%s</b> with password <b>%s</b></p>" +
		"<p><i>This password will not be shown again.</i></p>", matrixUserID, matrixUserPassword)
	log.Printf("Updated Matrix user %s \n", matrixUserID)
}


func main() {
	fmt.Println(appName)
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	if *fediverseServer == "" || *matrixServer == "" || *matrixAdminUser == "" || *matrixAdminPassword == "" {
		flag.PrintDefaults()
		return
	}

	log.Printf("Logging into Matrix %s as %s\n", *matrixServer, *matrixAdminUser)
	wellKnown, err := mautrix.DiscoverClientAPI(*matrixServer)
	if err != nil {
		log.Fatal(err)
	}
	matrixClient, err = mautrix.NewClient(wellKnown.Homeserver.BaseURL, "", "")
	if err != nil {
		log.Fatal(err)
	}
	_, err = matrixClient.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: *matrixAdminUser},
		Password:         *matrixAdminPassword,
		StoreCredentials: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Login successful\n")

	log.Printf("Connecting to fediverse %s \n", *fediverseServer)
	mastodonApp, err = mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:     "https://" + *fediverseServer,
		ClientName: appName,
		Scopes:     "read",
		Website:    *serverBaseURL,
		RedirectURIs: strings.TrimSuffix(*serverBaseURL, "/") + redirectURLSuffix,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected as client id %s\n", mastodonApp.ClientID)

	log.Printf("Listening on port %d with URL %s\n", *serverPort, *serverBaseURL)
	http.HandleFunc("/", webHandler)
	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(*serverPort), nil))
}
