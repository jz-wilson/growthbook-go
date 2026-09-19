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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	testAttrProperty = "plan_tier"
	testAttrDatatype = "string"
)

func TestAttributeRoundTrip(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody AttributeRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = AttributeRequest{}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"attributes": []Attribute{
				{Property: testAttrProperty, Datatype: testAttrDatatype},
			}})
		case http.MethodDelete:
			_ = json.NewEncoder(w).Encode(map[string]any{"deletedProperty": testAttrProperty})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"attribute": Attribute{
				Property: testAttrProperty, Datatype: testAttrDatatype,
			}})
		}
	}))
	defer srv.Close()

	c, err := NewFromSecret([]byte(`{"apiKey":"secret_1","apiUrl":"` + srv.URL + `/api"}`))
	if err != nil {
		t.Fatalf("NewFromSecret() error = %v", err)
	}
	ctx := context.Background()

	attrs, err := c.ListAttributes(ctx)
	if err != nil || gotPath != "/api/v1/attributes" || len(attrs) != 1 || attrs[0].Property != testAttrProperty {
		t.Errorf("ListAttributes() = %+v, path=%q, err %v", attrs, gotPath, err)
	}

	created, err := c.CreateAttribute(ctx, AttributeRequest{Property: testAttrProperty, Datatype: testAttrDatatype})
	if err != nil || gotMethod != http.MethodPost || gotPath != "/api/v1/attributes" || created.Property != testAttrProperty {
		t.Fatalf("CreateAttribute() = %+v, method=%q, path=%q, err %v", created, gotMethod, gotPath, err)
	}

	desc := "The customer's plan tier"
	updated, err := c.UpdateAttribute(ctx, testAttrProperty, AttributeRequest{Description: &desc})
	if err != nil || gotMethod != http.MethodPut || gotPath != "/api/v1/attributes/plan_tier" {
		t.Errorf("UpdateAttribute() = %+v, method=%q, path=%q, err %v", updated, gotMethod, gotPath, err)
	}
	if gotBody.Property != "" {
		t.Errorf("UpdateAttribute() sent property=%q, want unset (PUT schema rejects it)", gotBody.Property)
	}

	if err := c.DeleteAttribute(ctx, testAttrProperty); err != nil || gotMethod != http.MethodDelete || gotPath != "/api/v1/attributes/plan_tier" {
		t.Errorf("DeleteAttribute() err=%v method=%q path=%q", err, gotMethod, gotPath)
	}
}

func TestAttributeRequestOmitsUnset(t *testing.T) {
	b, err := json.Marshal(AttributeRequest{Property: testAttrProperty, Datatype: testAttrDatatype})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"property":"plan_tier","datatype":"string"}`, string(b)); diff != "" {
		t.Errorf("unset fields must be omitted: -want +got\n%s", diff)
	}
}

func TestAttributeRequestClearsListFields(t *testing.T) {
	empty := []string{}
	b, err := json.Marshal(AttributeRequest{Projects: &empty, Tags: &empty})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"projects":[],"tags":[]}`, string(b)); diff != "" {
		t.Errorf("a pointer to an empty slice must send []: -want +got\n%s", diff)
	}

	b, err = json.Marshal(AttributeRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{}`, string(b)); diff != "" {
		t.Errorf("nil Projects/Tags must be omitted, not sent as null: -want +got\n%s", diff)
	}
}

func TestAttributeDeleteNotFound(t *testing.T) {
	// GrowthBook sometimes returns 400 with a "Could not find ..." message
	// for a missing resource instead of a clean 404; IsNotFound already
	// handles both shapes for every resource.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Could not find attribute with property missing_attr"})
	}))
	defer srv.Close()

	c, err := NewFromSecret([]byte(`{"apiKey":"secret_1","apiUrl":"` + srv.URL + `/api"}`))
	if err != nil {
		t.Fatalf("NewFromSecret() error = %v", err)
	}

	if err := c.DeleteAttribute(context.Background(), "missing_attr"); !IsNotFound(err) {
		t.Errorf("DeleteAttribute(missing) err = %v, want not found", err)
	}
}
