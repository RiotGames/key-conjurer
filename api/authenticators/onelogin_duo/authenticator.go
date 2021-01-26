package oneloginduo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/riotgames/key-conjurer/api/core"

	"github.com/riotgames/key-conjurer/api/authenticators/duo"
	"github.com/riotgames/key-conjurer/api/settings"

	"github.com/rnikoopour/onelogin"
)

type Authenticator struct {
	settings *settings.Settings
	mfa      duo.Duo
}

// New creates a new OneLogin authenticator with the given MFA instance
func New(settings *settings.Settings, mfa duo.Duo) *Authenticator {
	return &Authenticator{
		settings: settings,
		mfa:      mfa,
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, creds core.Credentials) (core.User, error) {
	o := newOneLogin(a.settings)
	user, err := o.SamlClient.Oauth.Authenticate(ctx, creds.Username, creds.Password)
	if err != nil {
		return core.User{}, err
	}

	return core.User{ID: strconv.FormatInt(user.ID, 10)}, nil
}

func (a *Authenticator) ListRoles(context.Context, core.User) ([]core.Role, error) {
	// HACK: We return an error because in Riot's implementation of OneLogin, we do not use multiple roles.
	// If you are an open source contributor and want to extend OneLogin to support multiple roles, we will welcome PRs.
	return nil, errors.New("provider does not support listing roles")
}

func (a *Authenticator) ListApplications(ctx context.Context, user core.User) ([]core.Application, error) {
	o := newOneLogin(a.settings)
	id, err := strconv.ParseInt(user.ID, 10, 64)
	if err != nil {
		return nil, err
	}

	apps, err := o.ReadUserClient.User.GetApps(ctx, id)
	if err != nil {
		return nil, err
	}

	applications := []core.Application{}
	for _, app := range *apps {
		if strings.HasPrefix(app.Name, "AWS") {
			applications = append(applications, core.Application{
				LegacyID: uint(app.ID),
				ID:       strconv.FormatInt(app.ID, 10),
				Name:     app.Name,
			})
		}
	}

	return applications, nil
}

func (ola *Authenticator) GenerateSAMLAssertion(ctx context.Context, creds core.Credentials, appID string) (*core.SAMLResponse, error) {
	o := newOneLogin(ola.settings)
	stateTokenResponse, err := o.SamlClient.SAML.SamlAssertion(ctx, creds.Username, creds.Password, appID)
	if err != nil {
		return nil, err
	}

	device := &onelogin.Device{}
	for i, aDevice := range stateTokenResponse.Devices {
		if aDevice.DeviceType == "Duo Duo Security" {
			device = &stateTokenResponse.Devices[i]
		}
	}
	signatures := strings.Split(device.SignatureRequest, ":")
	txSignature := signatures[0]
	appSignature := signatures[1]

	mfaCookie, err := ola.mfa.SendPush(txSignature, stateTokenResponse.StateToken, stateTokenResponse.CallbackUrl, device.ApiHostName)
	if err != nil {
		return nil, err
	}

	mfaToken := fmt.Sprintf("%v:%v", mfaCookie, appSignature)

	samlString, err := o.GetSamlAssertion(ctx, mfaToken, stateTokenResponse.StateToken, appID, fmt.Sprint(device.Id))
	if err != nil {
		return nil, err
	}

	return core.ParseEncodedResponse(samlString)
}

var _ core.AuthenticationProvider = &Authenticator{}
