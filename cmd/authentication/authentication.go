package authentication

import (
	"context"
	"fmt"
	"github.com/healthcheck-exporter/cmd/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"net/http"
)

type AuthClient struct {
	client *http.Client
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
	}

	log.Info(fmt.Sprintf("Auth client registered on %s", config.Authentication.AuthUrl))

	return &authClient
}

func (ac *AuthClient) GetClient() *http.Client {
	return ac.client
}
