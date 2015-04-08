package rest

import (
	"fmt"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/ably/ably-go/Godeps/_workspace/src/github.com/flynn/flynn/pkg/random"
	"github.com/ably/ably-go/config"
)

type Capability map[string][]string

func (c Capability) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(b)
}

type Token struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Capability string `json:"capability"`
}

type tokenResponse struct {
	AccessToken *Token `json:"access_token"`
}

type TokenRequest struct {
	ID         string `json:"id"`
	TTL        int    `json:"ttl"`
	Capability string `json:"capability"`
	ClientID   string `json:"client_id"`
	Timestamp  int64  `json:"timestamp"`
	Nonce      string `json:"nonce"`
	Mac        string `json:"mac"`
}

func (t *TokenRequest) Sign(secret string) {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintln(mac, t.ID)
	fmt.Fprintln(mac, t.TTL)
	fmt.Fprintln(mac, t.Capability)
	fmt.Fprintln(mac, t.ClientID)
	fmt.Fprintln(mac, t.Timestamp)
	fmt.Fprintln(mac, t.Nonce)
	t.Mac = base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

type Auth struct {
	config.Params
	client *Client
}

func NewAuth(params config.Params, client *Client) *Auth {
	return &Auth{
		Params: params,
		client: client,
	}
}

func (a *Auth) CreateTokenRequest(ttl int, capability Capability) *TokenRequest {
	req := &TokenRequest{
		ID:         a.AppID,
		TTL:        ttl,
		Capability: capability.String(),
		ClientID:   a.ClientID,
		Timestamp:  time.Now().Unix(),
		Nonce:      random.String(32),
	}

	req.Sign(a.AppSecret)

	return req
}

func (a *Auth) RequestToken(ttl int, capability Capability) (*Token, error) {
	req := a.CreateTokenRequest(ttl, capability)

	res := &tokenResponse{}
	_, err := a.client.Post("/keys/"+a.AppID+"/requestToken", req, res)
	if err != nil {
		return nil, err
	}
	return res.AccessToken, nil
}
