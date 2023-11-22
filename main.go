package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	flag.Usage = usage
}

func main() {
	ctx := context.Background()

	flag.Parse()

	if flag.NArg() != 3 {
		flag.Usage()
		os.Exit(1)
	}

	regurl := flag.Arg(0)
	username := flag.Arg(1)
	password := flag.Arg(2)

	authUrl, authService, err := getAuthData(ctx, regurl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting auth url: %v\n", err)
		os.Exit(2)
	}

	token, err := getAuthToken(ctx, authUrl, authService, username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting auth token: %v\n", err)
		os.Exit(3)
	}

	json.NewEncoder(os.Stdout).Encode(token)
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s <REGISTRY_URL> <USERNAME> <PASSWORD>\n", os.Args[0])
}

func getAuthData(ctx context.Context, regurl string) (realm string, service string, err error) {
	u := &url.URL{}
	u, err = u.Parse(regurl)
	if err != nil {
		err = fmt.Errorf("invalid registry url: %w", err)
		return
	}

	u.Path += "/v2/"
	resp, err := http.Get(u.String())
	if err != nil {
		return
	}

	h := resp.Header.Get("WWW-Authenticate")
	if h == "" {
		err = fmt.Errorf("no WWW-Authenticate header")
		return
	}

	for _, sub := range strings.Split(h, ",") {
		parts := strings.SplitN(sub, "=", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid WWW-Authenticate header component: %s", parts)
		}
		if parts[0] == "Bearer realm" {
			realm = strings.Trim(parts[1], "\"")
			continue
		}
		if parts[0] == "service" {
			service = strings.Trim(parts[1], "\"")
			continue
		}
	}

	if realm == "" || service == "" {
		err = fmt.Errorf("invalid WWW-Authenticate header")
	}
	return
}

func getAuthToken(ctx context.Context, authUrl, authService, username, password string) (*jwt.Token, error) {
	u := &url.URL{}
	u, err := u.Parse(authUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid auth url: %w", err)
	}

	u.User = url.UserPassword(username, password)
	q := u.Query()
	q.Add("service", authService)
	q.Add("scope", "")
	u.RawQuery = q.Encode()

	req := http.Request{
		Method: http.MethodGet,
		URL:    u,
	}

	resp, err := http.DefaultClient.Do(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform auth request: %w", err)
	}

	data := make(map[string]any)
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth request failed (status %s): %s", resp.Status, data["details"])
	}

	tokenString, ok := data["token"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid auth response")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	if !errors.Is(err, jwt.ErrInvalidKeyType) {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return token, nil
}
