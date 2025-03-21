package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"k8s.io/client-go/util/flowcontrol"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
)

const (
	// Max QPS to allow through to the token URL.
	tokenURLQPS = .05 // back off to once every 20 seconds when failing
	// Maximum burst of requests to token URL before limiting.
	tokenURLBurst = 3
)

// TODO(#276) add metrics around token requests once the driver integrates with Prometheus.

// AltTokenSource is the structure holding the data for the functionality needed to generates tokens
type AltTokenSource struct {
	oauthClient *http.Client
	tokenURL    string
	tokenBody   string
	throttle    flowcontrol.RateLimiter
}

// Token returns a token which may be used for authentication
func (a *AltTokenSource) Token() (*oauth2.Token, error) {
	a.throttle.Accept()
	return a.token()
}

func (a *AltTokenSource) token() (*oauth2.Token, error) {
	req, err := http.NewRequest("POST", a.tokenURL, strings.NewReader(a.tokenBody))
	if err != nil {
		return nil, err
	}
	res, err := a.oauthClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	var tok struct {
		AccessToken string    `json:"accessToken"`
		ExpireTime  time.Time `json:"expireTime"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tok); err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken: tok.AccessToken,
		Expiry:      tok.ExpireTime,
	}, nil
}

// NewAltTokenSource constructs a new alternate token source for generating tokens.
func NewAltTokenSource(tokenURL, tokenBody string) oauth2.TokenSource {
	client := oauth2.NewClient(oauth2.NoContext, google.ComputeTokenSource(""))
	a := &AltTokenSource{
		oauthClient: client,
		tokenURL:    tokenURL,
		tokenBody:   tokenBody,
		throttle:    flowcontrol.NewTokenBucketRateLimiter(tokenURLQPS, tokenURLBurst),
	}
	return oauth2.ReuseTokenSource(nil, a)
}
