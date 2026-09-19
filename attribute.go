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

// Attribute is a GrowthBook SDK targeting attribute as returned by the API.
// Property is the attribute's identifier; the API has no GET-by-id endpoint,
// only list, so callers needing one attribute must list and filter.
type Attribute struct {
	Property      string `json:"property"`
	Datatype      string `json:"datatype"`
	Description   string `json:"description,omitempty"`
	HashAttribute bool   `json:"hashAttribute,omitempty"`
	Archived      bool   `json:"archived,omitempty"`
	// Enum is a comma-separated list of allowed values. Required for the
	// "enum" datatype; optionally restricts string[]/number[]/secureString[].
	Enum     string   `json:"enum,omitempty"`
	Format   string   `json:"format,omitempty"`
	Projects []string `json:"projects,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// AttributeRequest is the body for POST /v1/attributes and
// PUT /v1/attributes/{property}. Property and Datatype are required on
// create; PUT's schema does not accept Property at all (the API rejects
// unknown fields), so callers must leave it unset when updating.
type AttributeRequest struct {
	Property      string  `json:"property,omitempty"`
	Datatype      string  `json:"datatype,omitempty"`
	Description   *string `json:"description,omitempty"`
	Archived      *bool   `json:"archived,omitempty"`
	HashAttribute *bool   `json:"hashAttribute,omitempty"`
	Enum          *string `json:"enum,omitempty"`
	Format        *string `json:"format,omitempty"`
	// Projects and Tags are *[]string, not []string: a plain slice's
	// omitempty also drops an empty (non-nil) slice, so there would be no
	// way to send "[]" and clear the list. nil means omit the field; a
	// pointer to an empty slice means send "[]".
	Projects *[]string `json:"projects,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
}

type attributeEnvelope struct {
	Attribute Attribute `json:"attribute"`
}

type attributeListEnvelope struct {
	Attributes []Attribute `json:"attributes"`
}

// ListAttributes returns every SDK targeting attribute in the organization.
func (c *Client) ListAttributes(ctx context.Context) ([]Attribute, error) {
	var out attributeListEnvelope
	if err := c.do(ctx, http.MethodGet, "/v1/attributes", nil, &out); err != nil {
		return nil, err
	}
	return out.Attributes, nil
}

// CreateAttribute creates an SDK targeting attribute and returns the stored
// object.
func (c *Client) CreateAttribute(ctx context.Context, req AttributeRequest) (*Attribute, error) {
	var out attributeEnvelope
	if err := c.do(ctx, http.MethodPost, "/v1/attributes", req, &out); err != nil {
		return nil, err
	}
	return &out.Attribute, nil
}

// UpdateAttribute applies a partial update, keyed by property, and returns
// the stored object.
func (c *Client) UpdateAttribute(ctx context.Context, property string, req AttributeRequest) (*Attribute, error) {
	var out attributeEnvelope
	if err := c.do(ctx, http.MethodPut, "/v1/attributes/"+url.PathEscape(property), req, &out); err != nil {
		return nil, err
	}
	return &out.Attribute, nil
}

// DeleteAttribute deletes an attribute by property. Deleting an attribute
// that no longer exists returns an APIError satisfying IsNotFound.
func (c *Client) DeleteAttribute(ctx context.Context, property string) error {
	return c.do(ctx, http.MethodDelete, "/v1/attributes/"+url.PathEscape(property), nil, nil)
}
