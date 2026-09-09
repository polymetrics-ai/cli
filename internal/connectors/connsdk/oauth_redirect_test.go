package connsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestOAuthTokenExchangesRefuse307And308DestinationReplay(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var destinationHits atomic.Int32
			destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				destinationHits.Add(1)
				if r.FormValue("client_secret") != "" || r.FormValue("refresh_token") != "" {
					t.Fatal("redirect destination received OAuth credential material")
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer destination.Close()
			origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, destination.URL, status)
			}))
			defer origin.Close()

			for _, auth := range []Authenticator{
				&OAuth2ClientCredentials{TokenURL: origin.URL, ClientID: "client", ClientSecret: "secret"},
				&OAuth2RefreshToken{TokenURL: origin.URL, ClientID: "client", ClientSecret: "secret", RefreshToken: "refresh"},
			} {
				req, err := http.NewRequest(http.MethodGet, "https://api.example.test/resource", nil)
				if err != nil {
					t.Fatal(err)
				}
				if err := auth.Apply(context.Background(), req); err == nil {
					t.Fatal("OAuth redirect was accepted")
				}
			}
			if got := destinationHits.Load(); got != 0 {
				t.Fatalf("OAuth redirect destination hits = %d, want zero", got)
			}
		})
	}
}
