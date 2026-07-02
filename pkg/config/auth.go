package config

import (
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v88/github"
	"github.com/sirupsen/logrus"
)

type GithubAuth interface {
	ClientOptions() ([]github.ClientOptionsFunc, error)
}

type GithubToken string

func (t GithubToken) ClientOptions() ([]github.ClientOptionsFunc, error) {
	return []github.ClientOptionsFunc{
		github.WithAuthToken(string(t)),
	}, nil
}

type GithubApp struct {
	ID             int64
	InstallationID int64
	PrivateKey     string
}

func (a GithubApp) ClientOptions() ([]github.ClientOptionsFunc, error) {
	var transport http.RoundTripper
	var err error
	if _, serr := os.Stat(a.PrivateKey); serr == nil {
		logrus.Debugf("Loading GitHub App Private Key from %s", a.PrivateKey)
		// PrivateKey is path to a file that exists and can be read
		transport, err = ghinstallation.NewKeyFromFile(httpClient.Transport, a.ID, a.InstallationID, a.PrivateKey)
	} else {
		logrus.Debug("Loading GitHub App Private Key from byte value")
		// Try to load PrivateKey value as literal key bytes
		transport, err = ghinstallation.New(httpClient.Transport, a.ID, a.InstallationID, []byte(a.PrivateKey))
	}
	if err != nil {
		return nil, err
	}
	return []github.ClientOptionsFunc{
		github.WithHTTPClient(&http.Client{Timeout: httpClient.Timeout, Transport: transport}),
	}, nil
}
