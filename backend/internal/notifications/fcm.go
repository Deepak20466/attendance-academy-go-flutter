package notifications

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// FCMClient sends push notifications via the Firebase Cloud Messaging HTTP
// v1 API, authenticating with a service account key using a self-signed
// JWT exchanged for a short-lived OAuth2 access token. This avoids pulling
// in the full Google Cloud SDK for what is otherwise a single HTTP call.
type FCMClient struct {
	projectID    string
	serviceAcct  *serviceAccountKey
	httpClient   *http.Client
	tokenMu      sync.Mutex
	cachedToken  string
	cachedExpiry time.Time
}

type serviceAccountKey struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

func NewFCMClient(projectID, credentialsFilePath string) (*FCMClient, error) {
	if projectID == "" || credentialsFilePath == "" {
		return nil, errors.New("FCM_PROJECT_ID and FCM_CREDENTIALS_FILE must be set")
	}
	raw, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, fmt.Errorf("read FCM credentials file: %w", err)
	}
	var key serviceAccountKey
	if err := json.Unmarshal(raw, &key); err != nil {
		return nil, fmt.Errorf("parse FCM credentials file: %w", err)
	}
	if key.TokenURI == "" {
		key.TokenURI = "https://oauth2.googleapis.com/token"
	}
	return &FCMClient{
		projectID:   projectID,
		serviceAcct: &key,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *FCMClient) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.cachedToken != "" && time.Now().Before(c.cachedExpiry) {
		return c.cachedToken, nil
	}

	block, _ := pem.Decode([]byte(c.serviceAcct.PrivateKey))
	if block == nil {
		return "", errors.New("invalid private key in FCM credentials")
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("FCM private key is not RSA")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   c.serviceAcct.ClientEmail,
		"scope": "https://www.googleapis.com/auth/firebase.messaging",
		"aud":   c.serviceAcct.TokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedJWT, err := token.SignedString(rsaKey)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	form := fmt.Sprintf("grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=%s", signedJWT)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serviceAcct.TokenURI, bytes.NewBufferString(form))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request access token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	c.cachedToken = tokenResp.AccessToken
	c.cachedExpiry = now.Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	return c.cachedToken, nil
}

// Send pushes a single notification to a device token. Errors are
// returned to the caller to log — a failed push must never block the
// database write that already recorded the notification.
func (c *FCMClient) Send(ctx context.Context, deviceToken, title, body string, data map[string]string) error {
	accessToken, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": deviceToken,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
			"data": data,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM send failed with %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
