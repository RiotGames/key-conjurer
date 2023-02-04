package onelogin

import (
	"context"
	"fmt"
)

// UserService handles communications with the authentication related methods on OneLogin.
type UserService service

// User represents a OneLogin user.
type User struct {
	ActivatedAt          string            `json:"activated_at"`
	CreatedAt            string            `json:"created_at"`
	Email                string            `json:"email"`
	Username             string            `json:"username"`
	FirstName            string            `json:"firstname"`
	GroupID              int64             `json:"group_id"`
	ID                   int64             `json:"id"`
	InvalidLoginAttempts int64             `json:"invalid_login_attempts"`
	InvitationSentAt     string            `json:"invitation_sent_at"`
	LastLogin            string            `json:"last_login"`
	LastName             string            `json:"lastname"`
	LockedUntil          string            `json:"locked_until"`
	Notes                string            `json:"notes"`
	OpenidName           string            `json:"openid_name"`
	LocaleCode           string            `json:"locale_code"`
	PasswordChangedAt    string            `json:"password_changed_at"`
	Phone                string            `json:"phone"`
	Status               int64             `json:"status"`
	UpdatedAt            string            `json:"updated_at"`
	DistinguishedName    string            `json:"distinguished_name"`
	ExternalID           string            `json:"external_id"`
	DirectoryID          int64             `json:"directory_id"`
	MemberOf             []string          `json:"member_of"`
	SamAccountName       string            `json:"samaccountname"`
	UserPrincipalName    string            `json:"userprincipalname"`
	ManagerAdID          int               `json:"manager_ad_id"`
	RoleIDs              []int64           `json:"role_id"`
	CustomAttributes     map[string]string `json:"custom_attributes"`
}

type getUserQuery struct {
	AfterCursor string `url:"after_cursor,omitempty"`
}

type App struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetUsers returns all the OneLogin users.
func (s *UserService) GetUsers(ctx context.Context) ([]*User, error) {
	u := "/api/1/users"

	var users []*User
	var afterCursor string

	for {
		uu, err := addOptions(u, &getUserQuery{AfterCursor: afterCursor})
		if err != nil {
			return nil, err
		}

		req, err := s.client.NewRequest("GET", uu, nil)
		if err != nil {
			return nil, err
		}

		if err := s.client.AddAuthorization(ctx, req); err != nil {
			return nil, err
		}

		var us []*User
		resp, err := s.client.Do(ctx, req, &us)
		if err != nil {
			return nil, err
		}
		users = append(users, us...)
		if resp.PaginationAfterCursor == nil {
			break
		}

		afterCursor = *resp.PaginationAfterCursor
	}

	return users, nil
}

// GetUser returns a OneLogin user.
func (s *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
	u := fmt.Sprintf("/api/1/users/%v", id)

	var users []*User

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	if err := s.client.AddAuthorization(ctx, req); err != nil {
		return nil, err
	}

	_, err = s.client.Do(ctx, req, &users)
	if err != nil {
		return nil, err
	}

	return users[0], nil
}

func (s *UserService) GetApps(ctx context.Context, id int64) (*[]App, error) {
	u := fmt.Sprintf("/api/1/users/%v/apps", id)

	var userApps []App

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	if err := s.client.AddAuthorization(ctx, req); err != nil {
		return nil, err
	}
	if _, err := s.client.Do(ctx, req, &userApps); err != nil {
		return nil, err
	}
	return &userApps, nil
}

// UpdateCustomAttributes returns a OneLogin user.
func (s *UserService) UpdateCustomAttributes(ctx context.Context, id int64, attributes map[string]string) error {
	u := fmt.Sprintf("/api/1/users/%v/set_custom_attributes", id)

	post := map[string]interface{}{
		"custom_attributes": attributes,
	}

	req, err := s.client.NewRequest("PUT", u, post)
	if err != nil {
		return err
	}

	if err := s.client.AddAuthorization(ctx, req); err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	if err != nil {
		return err
	}

	return nil
}
