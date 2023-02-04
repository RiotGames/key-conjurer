package onelogin

import "golang.org/x/net/context"

// GroupService deals with OneLogin groups.
type GroupService service

type Group struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// GetGroups returns all the OneLogin groups.
func (s *GroupService) GetGroups(ctx context.Context) ([]*Group, error) {
	u := "/api/1/groups"

	var groups []*Group
	var afterCursor string

	for {
		uu, err := addOptions(u, &urlQuery{AfterCursor: afterCursor})
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

		var gs []*Group
		resp, err := s.client.Do(ctx, req, &gs)
		if err != nil {
			return nil, err
		}
		groups = append(groups, gs...)
		if resp.PaginationAfterCursor == nil {
			break
		}

		afterCursor = *resp.PaginationAfterCursor
	}

	return groups, nil
}
