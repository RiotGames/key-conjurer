package vault

import hashivault "github.com/hashicorp/vault/api"

func DefaultConfig() *hashivault.Config {
	return hashivault.DefaultConfig()
}

type Client struct {
	client *hashivault.Client
	Auth   *Auth
	KV2    *KV2
}

func NewClient(config *hashivault.Config) (*Client, error) {
	client, err := hashivault.NewClient(config)

	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		Auth:   NewAuth(client),
		KV2:    &KV2{client: client}}, nil
}
