vault-go-client
=====
# Under Development
This is a Golang client for Vault. It is currently under development.  v1.0.0 will be the first official release.

# Supported Auth Methods
- :heavy_check_mark: IAM
- :heavy_check_mark: AppRole
- :heavy_check_mark: LDAP
- :heavy_check_mark: Token
- k8s (coming soon)

# Supported Secret Stores
- :heavy_check_mark: KV2

# Usage
To retrieve this package run:
```
go get github.com/riotgames/vault-go-client
```

## Creating a Client
The following will create a client with default configuration:
```
import vault "github.com/riotgames/vault-go-client"
...

// Uses VAULT_ADDR env var to set the clients URL
client, err := vault.NewClient(vault.DefaultConfig())

if err != nil {
    log.Fatal(err.Error())
}
...
```

## Putting a Secret into Vault
The following will put a secret into Vault:
```
secretMap := map[string]interface{}{
    "hello": "world",
}

if _, err = client.KV2.Put(vault.KV2PutOptions{
	MountPath:  secretMountPath,
	SecretPath: secretPath,
	Secrets:    secretMap,
}); err != nil {
	log.Fatal(err.Error())
}
```

## Retrieving a Secret from Vault
### Unmarshaling Approach
This approach unmarshals the secret from Vault into the provided struct. 
The embedded struct `vault.SecretMetadata` is optional.
```
type Secret struct {
	Hello string `json:"hello"`
	vault.SecretMetadata
}
...
secret := &Secret{}

if _, err = client.KV2.Get(vault.KV2GetOptions{
	MountPath:     secretMountPath,
	SecretPath:    secretPath,
	UnmarshalInto: secret,
}); err != nil {
	log.Fatal(err.Error())
}
fmt.Printf("%v\n", secret)
```
### Raw Secret Approach
This approach returns a `Secret` defined in `github.com/hashicorp/vault/api`.
```
secret, err := client.KV2.Get(vault.KV2GetOptions{
	MountPath:  secretMountPath,
	SecretPath: secretPath,
})

if err != nil {
	log.Fatal(err.Error())
}
```
