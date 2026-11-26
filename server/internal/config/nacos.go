package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const maxNacosConfigBytes = 2 * 1024 * 1024

type nacosSettings struct {
	baseURL                  string
	namespace, group, dataID string
	username, password       string
	timeout                  time.Duration
}

func configurationSource() (string, error) {
	source := strings.ToLower(strings.TrimSpace(os.Getenv("CASINO_CONFIG_SOURCE")))
	if source == "" || source == "auto" {
		if strings.TrimSpace(os.Getenv("NACOS_SERVER_ADDR")) != "" {
			return "nacos", nil
		}
		return "local", nil
	}
	if source != "local" && source != "nacos" {
		return "", errors.New("CASINO_CONFIG_SOURCE must be auto, local or nacos")
	}
	return source, nil
}

func nacosFromEnvironment() (nacosSettings, error) {
	s := nacosSettings{
		namespace: strings.TrimSpace(os.Getenv("NACOS_NAMESPACE_ID")),
		group:     strings.TrimSpace(os.Getenv("NACOS_GROUP")),
		dataID:    strings.TrimSpace(os.Getenv("NACOS_DATA_ID")),
		username:  os.Getenv("NACOS_USERNAME"), password: os.Getenv("NACOS_PASSWORD"),
		timeout: 15 * time.Second,
	}
	address := strings.TrimSpace(os.Getenv("NACOS_SERVER_ADDR"))
	if address == "" {
		return s, errors.New("NACOS_SERVER_ADDR is required for Nacos configuration")
	}
	if !strings.Contains(address, "://") {
		address = "http://" + address
	}
	u, err := url.Parse(address)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return s, errors.New("NACOS_SERVER_ADDR must be an HTTP(S) address without credentials or query parameters")
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if u.Path == "" {
		u.Path = "/nacos"
	}
	s.baseURL = strings.TrimRight(u.String(), "/")
	if s.group == "" {
		s.group = "DEFAULT_GROUP"
	}
	if s.dataID == "" {
		return s, errors.New("NACOS_DATA_ID is required for Nacos configuration")
	}
	if (s.username == "") != (s.password == "") {
		return s, errors.New("NACOS_USERNAME and NACOS_PASSWORD must both be supplied")
	}
	if raw := strings.TrimSpace(os.Getenv("NACOS_TIMEOUT_SECONDS")); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds < 1 || seconds > 120 {
			return s, errors.New("NACOS_TIMEOUT_SECONDS must be between 1 and 120")
		}
		s.timeout = time.Duration(seconds) * time.Second
	}
	return s, nil
}

// Read only the selected document using the Nacos 2.x compatible HTTP API.
// Never include response bodies, request URLs, tokens or credentials in errors.
func readNacosConfig(ctx context.Context, s nacosSettings) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	client := &http.Client{
		Timeout: s.timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("Nacos redirects are not allowed")
		},
	}
	token := ""
	if s.username != "" {
		body := url.Values{"username": {s.username}, "password": {s.password}}.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/auth/login", strings.NewReader(body))
		if err != nil {
			return nil, errors.New("could not construct Nacos authentication request")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		data, err := nacosRequest(client, req, "authentication", 64*1024)
		if err != nil {
			return nil, err
		}
		var login struct {
			AccessToken string `json:"accessToken"`
		}
		if json.Unmarshal(data, &login) != nil || login.AccessToken == "" {
			return nil, errors.New("Nacos authentication did not return a valid access token")
		}
		token = login.AccessToken
	}
	query := url.Values{"dataId": {s.dataID}, "group": {s.group}}
	// In Nacos 2.3 public is the empty namespace; omit its HTTP tenant parameter.
	if s.namespace != "" {
		query.Set("tenant", s.namespace)
	}
	if token != "" {
		query.Set("accessToken", token)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/v1/cs/configs?"+query.Encode(), nil)
	if err != nil {
		return nil, errors.New("could not construct Nacos configuration request")
	}
	return nacosRequest(client, req, "configuration read", maxNacosConfigBytes)
}

func nacosRequest(client *http.Client, req *http.Request, operation string, limit int64) ([]byte, error) {
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Nacos %s failed; check connectivity, TLS and timeout", operation)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Nacos %s returned HTTP %d", operation, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("Nacos %s response could not be read", operation)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("Nacos %s response exceeds the size limit", operation)
	}
	return data, nil
}

func parseNacosYAML(data []byte, target *Config) error {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return errors.New("Nacos configuration is not valid YAML")
	}
	for _, section := range []string{"app", "http", "database"} {
		if len(v.GetStringMap(section)) == 0 {
			return fmt.Errorf("Nacos YAML must include the %s section", section)
		}
	}
	if err := v.Unmarshal(target); err != nil {
		return errors.New("Nacos YAML does not match the application configuration structure")
	}
	return nil
}
