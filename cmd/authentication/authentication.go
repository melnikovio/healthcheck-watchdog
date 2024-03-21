package authentication

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type AuthClient struct {
	client *http.Client
	oauth  *clientcredentials.Config
	token  *oauth2.Token
}

func NewAuthClient(config *model.Config) *AuthClient {
	if config.Authentication.ClientId == "" || config.Authentication.ClientSecret == "" ||
		config.Authentication.AuthUrl == "" {
		err := errors.New("missing authentication parameters")
		log.Error(err.Error())
		panic(err)
	}

	oauth := &clientcredentials.Config{
		ClientID:     config.Authentication.ClientId,
		ClientSecret: config.Authentication.ClientSecret,
		TokenURL:     config.Authentication.AuthUrl + "/protocol/openid-connect/token",
	}

	ctx := context.Background()
	client := oauth.Client(ctx)

	authClient := AuthClient{
		client: client,
		oauth:  oauth,
	}

	log.Info(fmt.Sprintf("Auth client initialized on %s", config.Authentication.AuthUrl))

	return &authClient
}

func (ac *AuthClient) GetClient() *http.Client {
	return ac.client
}

func (ac *AuthClient) GetToken() *oauth2.Token {
	ctx := context.Background()
	var err error

	if ac.token == nil || time.Now().After(ac.token.Expiry) {
		ac.token, err = ac.oauth.Token(ctx)
		if err != nil {
			log.Error(fmt.Sprintf("Error while get client token: %s", err.Error()))
		}

		log.Info(fmt.Sprintf("Successfully obtained access token with lifetime until %s",
			ac.token.Expiry.String()))
	}

	return ac.token
}
