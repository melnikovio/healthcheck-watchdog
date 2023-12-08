package authentication

import (
	"context"
	"fmt"
	"net/http"

	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type AuthClient struct {
	client *http.Client
	oauth *clientcredentials.Config
}

func NewAuthClient(config *model.Config) *AuthClient {
	oauth := &clientcredentials.Config{
		ClientID:     config.Authentication.ClientId,
		ClientSecret: config.Authentication.ClientSecret,
		TokenURL:     config.Authentication.AuthUrl + "/protocol/openid-connect/token",
	}

	ctx := context.Background()
	client := oauth.Client(ctx)

	authClient := AuthClient{
		client: client,
		oauth: oauth,
	}

	log.Info(fmt.Sprintf("Auth client registered on %s", config.Authentication.AuthUrl))

	return &authClient
}

func (ac *AuthClient) GetClient() *http.Client {
	return ac.client
}

func (ac *AuthClient) GetToken() *oauth2.Token {
	ctx := context.Background()
	token, err := ac.oauth.Token(ctx)
	if err != nil {
		log.Error(fmt.Sprintf("Error while get client token: %s", err.Error()))
	}
	return token
}

