package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/7StaSH7/practicum-diploma/internal/api"
)

func runSignup(args []string, stdout io.Writer) error {
	cfg, err := parseAuthFlags("signup", args)
	if err != nil {
		return err
	}
	client := api.New(cfg.serverURL, apiHTTPClientFactory())
	resp, err := client.Signup(context.Background(), cfg.login, cfg.password)
	if err != nil {
		return err
	}
	sess := session{
		ServerURL:    cfg.serverURL,
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		KDFSalt:      resp.KDFSalt,
	}
	if err := saveSession(sess); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "signup successful, user_id=%s\n", resp.UserID)
	return err
}

func runSignin(args []string, stdout io.Writer) error {
	cfg, err := parseAuthFlags("signin", args)
	if err != nil {
		return err
	}
	client := api.New(cfg.serverURL, apiHTTPClientFactory())
	resp, err := client.Signin(context.Background(), cfg.login, cfg.password)
	if err != nil {
		return err
	}
	sess := session{
		ServerURL:    cfg.serverURL,
		UserID:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		KDFSalt:      resp.KDFSalt,
	}
	if err := saveSession(sess); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "signin successful, user_id=%s\n", resp.UserID)
	return err
}

func runRefresh(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("refresh", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	sess, err := loadSession()
	if err != nil {
		return err
	}
	sess.ServerURL = effectiveServerURL(*serverURL, sess.ServerURL)
	if sess.ServerURL == "" {
		return errors.New("server URL is required (--server or SERVER_URL)")
	}
	if sess.RefreshToken == "" {
		return errors.New("refresh token is missing, run signin or signup")
	}

	client := api.New(sess.ServerURL, apiHTTPClientFactory())
	resp, err := client.Refresh(context.Background(), sess.RefreshToken)
	if err != nil {
		return err
	}
	sess.UserID = resp.UserID
	sess.AccessToken = resp.AccessToken
	sess.RefreshToken = resp.RefreshToken
	sess.KDFSalt = resp.KDFSalt
	if err := saveSession(sess); err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, "token refresh successful")
	return err
}

func parseAuthFlags(name string, args []string) (struct {
	serverURL string
	login     string
	password  string
}, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	login := fs.String("login", "", "User login")
	password := fs.String("password", "", "User password")
	if err := fs.Parse(args); err != nil {
		return struct {
			serverURL string
			login     string
			password  string
		}{}, err
	}
	resolvedServerURL := effectiveServerURL(*serverURL, "")
	if resolvedServerURL == "" {
		return struct {
			serverURL string
			login     string
			password  string
		}{}, errors.New("server URL is required (--server or SERVER_URL)")
	}
	if strings.TrimSpace(*login) == "" {
		return struct {
			serverURL string
			login     string
			password  string
		}{}, errors.New("--login is required")
	}
	if strings.TrimSpace(*password) == "" {
		return struct {
			serverURL string
			login     string
			password  string
		}{}, errors.New("--password is required")
	}
	return struct {
		serverURL string
		login     string
		password  string
	}{
		serverURL: resolvedServerURL,
		login:     strings.TrimSpace(*login),
		password:  *password,
	}, nil
}

func withAutoRefresh(sess session, client *api.API, fn func(accessToken string) error) (session, error) {
	err := fn(sess.AccessToken)
	if !api.IsHTTPStatus(err, http.StatusUnauthorized) {
		return sess, err
	}
	if sess.RefreshToken == "" {
		return sess, err
	}
	refreshed, refreshErr := client.Refresh(context.Background(), sess.RefreshToken)
	if refreshErr != nil {
		return sess, refreshErr
	}
	sess.UserID = refreshed.UserID
	sess.AccessToken = refreshed.AccessToken
	sess.RefreshToken = refreshed.RefreshToken
	sess.KDFSalt = refreshed.KDFSalt
	if requestErr := fn(sess.AccessToken); requestErr != nil {
		return sess, requestErr
	}
	return sess, nil
}

func loadSessionAndClient(overrideURL string) (session, *api.API, error) {
	sess, err := loadSession()
	if err != nil {
		return session{}, nil, err
	}
	sess.ServerURL = effectiveServerURL(overrideURL, sess.ServerURL)
	if sess.ServerURL == "" {
		return session{}, nil, errors.New("server URL is missing; use --server or set SERVER_URL")
	}
	return sess, api.New(sess.ServerURL, apiHTTPClientFactory()), nil
}
