package keyconjurer

type STSTokenResponse struct {
	AccessKeyID     *string `json:"accessKeyId"`
	SecretAccessKey *string `json:"secretAccessKey"`
	SessionToken    *string `json:"sessionToken"`
	Expiration      string  `json:"expiration"`
}
