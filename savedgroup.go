/*
Copyright 2026 The growthbook-go Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package growthbook

import (
	"context"
	"net/http"
	"net/url"
)

// SavedGroup is the GrowthBook saved group object as returned by the API.
type SavedGroup struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	DateCreated  string   `json:"dateCreated"`
	DateUpdated  string   `json:"dateUpdated"`
	Name         string   `json:"name"`
	Owner        string   `json:"owner,omitempty"`
	OwnerEmail   string   `json:"ownerEmail,omitempty"`
	Condition    string   `json:"condition,omitempty"`
	AttributeKey string   `json:"attributeKey,omitempty"`
	Values       []string `json:"values,omitempty"`
	Description  string   `json:"description,omitempty"`
	Projects     []string `json:"projects,omitempty"`
	Archived     *bool    `json:"archived,omitempty"`
	UseEmptyList *bool    `json:"useEmptyListGroup,omitempty"`
}

// SavedGroupRequest is the body for POST /v1/saved-groups (create) and
// POST /v1/saved-groups/{id} (update). Name is required on create; every
// field is optional on update. Type is inferred by the API from the other
// fields when omitted, so it is only sent on create.
//
// Values and Projects are *[]string, not []string: a nil pointer omits the
// field, while a pointer to an empty slice sends "[]" so an update can
// clear a previously-set list. encoding/json's omitempty treats a nil and
// an empty []string identically, which would make clearing impossible.
type SavedGroupRequest struct {
	Name           string    `json:"name,omitempty"`
	Type           string    `json:"type,omitempty"`
	Condition      *string   `json:"condition,omitempty"`
	AttributeKey   string    `json:"attributeKey,omitempty"`
	Values         *[]string `json:"values,omitempty"`
	Owner          *string   `json:"owner,omitempty"`
	Projects       *[]string `json:"projects,omitempty"`
	BypassApproval *bool     `json:"bypassApproval,omitempty"`
}

type savedGroupEnvelope struct {
	SavedGroup SavedGroup `json:"savedGroup"`
}

type savedGroupListEnvelope struct {
	SavedGroups []SavedGroup `json:"savedGroups"`
}

// GetSavedGroup fetches one saved group by id. A missing saved group returns
// an APIError satisfying IsNotFound.
func (c *Client) GetSavedGroup(ctx context.Context, id string) (*SavedGroup, error) {
	var out savedGroupEnvelope
	if err := c.do(ctx, http.MethodGet, "/v1/saved-groups/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out.SavedGroup, nil
}

// ListSavedGroups returns every saved group in the organization.
func (c *Client) ListSavedGroups(ctx context.Context) ([]SavedGroup, error) {
	var out savedGroupListEnvelope
	if err := c.do(ctx, http.MethodGet, "/v1/saved-groups", nil, &out); err != nil {
		return nil, err
	}
	return out.SavedGroups, nil
}

// CreateSavedGroup creates a saved group and returns the stored object.
func (c *Client) CreateSavedGroup(ctx context.Context, req SavedGroupRequest) (*SavedGroup, error) {
	var out savedGroupEnvelope
	if err := c.do(ctx, http.MethodPost, "/v1/saved-groups", req, &out); err != nil {
		return nil, err
	}
	return &out.SavedGroup, nil
}

// UpdateSavedGroup applies a partial update and returns the stored object.
// Note the API uses POST, not PUT, for this endpoint.
func (c *Client) UpdateSavedGroup(ctx context.Context, id string, req SavedGroupRequest) (*SavedGroup, error) {
	var out savedGroupEnvelope
	if err := c.do(ctx, http.MethodPost, "/v1/saved-groups/"+url.PathEscape(id), req, &out); err != nil {
		return nil, err
	}
	return &out.SavedGroup, nil
}

// DeleteSavedGroup deletes a saved group. Deleting a saved group that no
// longer exists returns an APIError satisfying IsNotFound. GrowthBook
// refuses to delete a saved group that isn't archived first (HTTP 400,
// "must be archived before it can be deleted"); call ArchiveSavedGroup
// before DeleteSavedGroup to avoid that.
func (c *Client) DeleteSavedGroup(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/v1/saved-groups/"+url.PathEscape(id), nil, nil)
}

// ArchiveSavedGroup archives a saved group, a required step before deleting
// it. Archiving a saved group still referenced by a feature, experiment, or
// another saved group returns an APIError with StatusCode 422.
func (c *Client) ArchiveSavedGroup(ctx context.Context, id string) (*SavedGroup, error) {
	var out savedGroupEnvelope
	if err := c.do(ctx, http.MethodPost, "/v1/saved-groups/"+url.PathEscape(id)+"/archive", struct{}{}, &out); err != nil {
		return nil, err
	}
	return &out.SavedGroup, nil
}
