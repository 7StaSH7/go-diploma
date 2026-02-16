package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	dtoauth "github.com/7StaSH7/practicum-diploma/internal/dto/auth"
	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	"github.com/7StaSH7/practicum-diploma/pkg/apiclient"
)

type API struct {
	client *apiclient.Client
}

func New(baseURL string, httpClient *http.Client) *API {
	return &API{
		client: apiclient.New(baseURL, httpClient),
	}
}

func (a *API) Signup(ctx context.Context, login, password string) (dtoauth.AuthResponse, error) {
	var out dtoauth.AuthResponse
	err := a.client.DoJSON(ctx, http.MethodPost, "/auth/signup", nil, dtoauth.AuthRequest{
		Login:    login,
		Password: password,
	}, &out)
	if err != nil {
		return dtoauth.AuthResponse{}, err
	}
	return out, nil
}

func (a *API) Signin(ctx context.Context, login, password string) (dtoauth.AuthResponse, error) {
	var out dtoauth.AuthResponse
	err := a.client.DoJSON(ctx, http.MethodPost, "/auth/signin", nil, dtoauth.AuthRequest{
		Login:    login,
		Password: password,
	}, &out)
	if err != nil {
		return dtoauth.AuthResponse{}, err
	}
	return out, nil
}

func (a *API) Refresh(ctx context.Context, refreshToken string) (dtoauth.AuthResponse, error) {
	var out dtoauth.AuthResponse
	err := a.client.DoJSON(ctx, http.MethodPost, "/auth/refresh", nil, dtoauth.RefreshRequest{
		RefreshToken: refreshToken,
	}, &out)
	if err != nil {
		return dtoauth.AuthResponse{}, err
	}
	return out, nil
}

func (a *API) ListSecrets(ctx context.Context, accessToken, since string) ([]dtosecret.SecretResponse, error) {
	path := "/secrets"
	if strings.TrimSpace(since) != "" {
		path += "?since=" + url.QueryEscape(since)
	}

	var out []dtosecret.SecretResponse
	err := a.client.DoJSON(ctx, http.MethodGet, path, authHeader(accessToken), nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (a *API) GetSecret(ctx context.Context, accessToken, id string) (dtosecret.SecretResponse, error) {
	var out dtosecret.SecretResponse
	err := a.client.DoJSON(ctx, http.MethodGet, "/secrets/"+id, authHeader(accessToken), nil, &out)
	if err != nil {
		return dtosecret.SecretResponse{}, err
	}
	return out, nil
}

func (a *API) CreateSecret(ctx context.Context, accessToken string, payload dtosecret.SecretPayload) (dtosecret.SecretResponse, error) {
	var out dtosecret.SecretResponse
	err := a.client.DoJSON(ctx, http.MethodPost, "/secrets", authHeader(accessToken), payload, &out)
	if err != nil {
		return dtosecret.SecretResponse{}, err
	}
	return out, nil
}

func (a *API) UpdateSecret(ctx context.Context, accessToken, id string, payload dtosecret.SecretPayload) (dtosecret.SecretResponse, error) {
	var out dtosecret.SecretResponse
	err := a.client.DoJSON(ctx, http.MethodPut, "/secrets/"+id, authHeader(accessToken), payload, &out)
	if err != nil {
		return dtosecret.SecretResponse{}, err
	}
	return out, nil
}

func (a *API) DeleteSecret(ctx context.Context, accessToken, id string) error {
	return a.client.DoJSON(ctx, http.MethodDelete, "/secrets/"+id, authHeader(accessToken), nil, nil)
}

func IsHTTPStatus(err error, statusCode int) bool {
	var httpErr *apiclient.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	return httpErr.StatusCode == statusCode
}

func authHeader(accessToken string) map[string]string {
	if strings.TrimSpace(accessToken) == "" {
		return nil
	}
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
}
