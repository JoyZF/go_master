package main

import (
	"fmt"
	"net/http"

	"github.com/emanoelxavier/openid2go/openid"
)

func AuthenticatedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "The user was authenticated!")
}

func AuthenticatedHandlerWithUser(u *openid.User, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "The user was authenticated! The token was issued by %v and the user is %+v.", u.Issuer, u)
}

func main() {
	configuration, err := openid.NewConfiguration(openid.ProvidersGetter(myGetProviders))

	if err != nil {
		panic(err)
	}

	http.Handle("/user", openid.AuthenticateUser(configuration, openid.UserHandlerFunc(AuthenticatedHandlerWithUser)))
	http.Handle("/authn", openid.Authenticate(configuration, http.HandlerFunc(AuthenticatedHandler)))

	http.ListenAndServe(":5100", nil)
}

func myGetProviders() ([]openid.Provider, error) {
	provider, err := openid.NewProvider("https://steamcommunity.com/openid/", []string{"1969870"})

	if err != nil {
		return nil, err
	}

	return []openid.Provider{provider}, nil
}
