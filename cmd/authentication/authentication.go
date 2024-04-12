package authentication

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/healthcheck-watchdog/cmd/common"
	"github.com/healthcheck-watchdog/cmd/model"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type AuthClient struct {
	clients map[string]*http.Client
	oauths  map[string]*clientcredentials.Config
	tokens  map[string]*oauth2.Token
	mutex   sync.Mutex
}

func NewAuthClient(config *model.Config) *AuthClient {
	clients, oauths := getClients(config)

	if len(clients) == 0 {
		err := errors.New("missing authentication parameters")
		log.Error(err.Error())
		panic(err)
	}

	authClient := AuthClient{
		clients: clients,
		oauths:  oauths,
		tokens:  make(map[string]*oauth2.Token),
	}

	log.Info(fmt.Sprintf("%d auth clients initialized", len(clients)))

	return &authClient
}

func getClients(config *model.Config) (clients map[string]*http.Client,
	oauths map[string]*clientcredentials.Config) {
	clients = make(map[string]*http.Client)
	oauths = make(map[string]*clientcredentials.Config)
	for id, clientConfig := range config.AuthenticationClients {
		oauth := &clientcredentials.Config{
			ClientID:     clientConfig.ClientId,
			ClientSecret: clientConfig.ClientSecret,
			TokenURL:     clientConfig.AuthUrl + "/protocol/openid-connect/token",
		}

		ctx := context.Background()
		client := oauth.Client(ctx)

		clients[id] = client
		oauths[id] = oauth
	}

	return clients, oauths
}

func (ac *AuthClient) GetClient(id string) *http.Client {
	client, ok := ac.clients[id]
	if !ok {
		log.Error(fmt.Sprintf("Auth client with id %s not found", id))
		client, ok = ac.clients[common.DefaultClientId]
		if !ok {
			log.Error(fmt.Sprintf("Auth client with id %s not found", common.DefaultClientId))
			client = nil
		}
	}
	return client
}

func (ac *AuthClient) getOauth(id string) *clientcredentials.Config {
	oauth, ok := ac.oauths[id]
	if !ok {
		log.Error(fmt.Sprintf("Auth oauth config with id %s not found", id))
		oauth, ok = ac.oauths[common.DefaultClientId]
		if !ok {
			log.Error(fmt.Sprintf("Auth oauth config with id %s not found", common.DefaultClientId))
			oauth = nil
		}
	}
	return oauth
}

func (ac *AuthClient) GetToken(id string) *oauth2.Token {
	ac.mutex.Lock()
	defer ac.mutex.Unlock()

	ctx := context.Background()
	var err error

	if ac.tokens[id] == nil || time.Now().After(ac.tokens[id].Expiry) {
		ac.tokens[id], err = ac.getOauth(id).Token(ctx)
		if err != nil {
			log.Error(fmt.Sprintf("Error while get client token: %s", err.Error()))
		}

		log.Info(fmt.Sprintf("Successfully obtained access token with lifetime until %s",
			ac.tokens[id].Expiry.String()))
	}

	return ac.tokens[id]
}
