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
	testSavedGroupName         = "internal-users"
	testSavedGroupType         = "list"
	testSavedGroupAttributeKey = "userId"
	testSavedGroupID           = "sg_1"
)

func TestSavedGroupRoundTrip(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody SavedGroupRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody = SavedGroupRequest{}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/saved-groups":
			_ = json.NewEncoder(w).Encode(map[string]any{"savedGroups": []SavedGroup{
				{ID: testSavedGroupID, Name: testSavedGroupName, Type: testSavedGroupType},
			}})
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"savedGroup": SavedGroup{
				ID: testSavedGroupID, Name: testSavedGroupName, Type: testSavedGroupType,
				AttributeKey: testSavedGroupAttributeKey, Values: []string{"a", "b"},
			}})
		case r.Method == http.MethodDelete:
			_, _ = w.Write([]byte(`{"deletedId":"sg_1"}`))
		default:
			var values []string
			if gotBody.Values != nil {
				values = *gotBody.Values
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"savedGroup": SavedGroup{
				ID: testSavedGroupID, Name: gotBody.Name, Type: testSavedGroupType,
				AttributeKey: gotBody.AttributeKey, Values: values,
			}})
		}
	}))
	defer srv.Close()

	c, err := NewFromSecret([]byte(`{"apiKey":"secret_1","apiUrl":"` + srv.URL + `/api"}`))
	if err != nil {
		t.Fatalf("NewFromSecret() error = %v", err)
	}
	ctx := context.Background()

	groups, err := c.ListSavedGroups(ctx)
	if err != nil || gotPath != "/api/v1/saved-groups" || len(groups) != 1 || groups[0].ID != testSavedGroupID {
		t.Errorf("ListSavedGroups() = %+v, path=%q, err %v", groups, gotPath, err)
	}

	got, err := c.GetSavedGroup(ctx, testSavedGroupID)
	if err != nil || gotPath != "/api/v1/saved-groups/sg_1" || got.Name != testSavedGroupName || len(got.Values) != 2 {
		t.Errorf("GetSavedGroup() = %+v, path=%q, err %v", got, gotPath, err)
	}

	created, err := c.CreateSavedGroup(ctx, SavedGroupRequest{
		Name: testSavedGroupName, Type: testSavedGroupType,
		AttributeKey: testSavedGroupAttributeKey, Values: savedGroupStrSlicePtr([]string{"x", "y"}),
	})
	if err != nil || gotMethod != http.MethodPost || gotPath != "/api/v1/saved-groups" {
		t.Fatalf("CreateSavedGroup() err=%v method=%q path=%q", err, gotMethod, gotPath)
	}
	if created.ID != testSavedGroupID || len(created.Values) != 2 || created.Values[0] != "x" {
		t.Errorf("CreateSavedGroup() = %+v, want values echoed back", created)
	}

	updated, err := c.UpdateSavedGroup(ctx, testSavedGroupID, SavedGroupRequest{Values: savedGroupStrSlicePtr([]string{"z"})})
	if err != nil || gotMethod != http.MethodPost || gotPath != "/api/v1/saved-groups/sg_1" || len(updated.Values) != 1 || updated.Values[0] != "z" {
		t.Errorf("UpdateSavedGroup() = %+v, method=%q, path=%q, err %v", updated, gotMethod, gotPath, err)
	}

	// Clearing values/projects to empty must send "[]", not omit the field
	// (a nil *[]string omits; a pointer to an empty slice sends "[]").
	cleared, err := c.UpdateSavedGroup(ctx, testSavedGroupID, SavedGroupRequest{Values: savedGroupStrSlicePtr([]string{})})
	if err != nil || len(cleared.Values) != 0 {
		t.Errorf("UpdateSavedGroup(empty values) = %+v, err %v, want values cleared", cleared, err)
	}

	if err := c.DeleteSavedGroup(ctx, testSavedGroupID); err != nil || gotMethod != http.MethodDelete || gotPath != "/api/v1/saved-groups/sg_1" {
		t.Errorf("DeleteSavedGroup() err=%v method=%q path=%q", err, gotMethod, gotPath)
	}
}

func TestSavedGroupRequestOmitsUnset(t *testing.T) {
	b, err := json.Marshal(SavedGroupRequest{Name: testSavedGroupName, Type: testSavedGroupType, AttributeKey: testSavedGroupAttributeKey})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"name":"internal-users","type":"list","attributeKey":"userId"}`, string(b)); diff != "" {
		t.Errorf("unset fields must be omitted: -want +got\n%s", diff)
	}
}

func TestSavedGroupRequestValuesProjectsNilVsEmpty(t *testing.T) {
	// nil *[]string omits the field entirely; a pointer to an empty slice
	// must marshal to "[]" so an update can clear a previously-set list.
	nilBody, err := json.Marshal(SavedGroupRequest{Name: testSavedGroupName})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"name":"internal-users"}`, string(nilBody)); diff != "" {
		t.Errorf("nil Values/Projects must be omitted: -want +got\n%s", diff)
	}

	emptyBody, err := json.Marshal(SavedGroupRequest{
		Name:     testSavedGroupName,
		Values:   savedGroupStrSlicePtr([]string{}),
		Projects: savedGroupStrSlicePtr([]string{}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"name":"internal-users","values":[],"projects":[]}`, string(emptyBody)); diff != "" {
		t.Errorf("empty-slice Values/Projects must marshal to []: -want +got\n%s", diff)
	}
}

func savedGroupStrSlicePtr(s []string) *[]string { return &s }

func TestArchiveSavedGroup(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		archived := true
		_ = json.NewEncoder(w).Encode(map[string]any{"savedGroup": SavedGroup{
			ID: testSavedGroupID, Name: testSavedGroupName, Type: testSavedGroupType, Archived: &archived,
		}})
	}))
	defer srv.Close()

	c, err := NewFromSecret([]byte(`{"apiKey":"secret_1","apiUrl":"` + srv.URL + `/api"}`))
	if err != nil {
		t.Fatalf("NewFromSecret() error = %v", err)
	}

	got, err := c.ArchiveSavedGroup(context.Background(), testSavedGroupID)
	if err != nil || gotMethod != http.MethodPost || gotPath != "/api/v1/saved-groups/sg_1/archive" {
		t.Fatalf("ArchiveSavedGroup() err=%v method=%q path=%q", err, gotMethod, gotPath)
	}
	if got.Archived == nil || !*got.Archived {
		t.Errorf("ArchiveSavedGroup() Archived = %v, want true", got.Archived)
	}
}

func TestSavedGroupNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Could not find savedGroup with id sg_missing"})
	}))
	defer srv.Close()

	c, err := NewFromSecret([]byte(`{"apiKey":"secret_1","apiUrl":"` + srv.URL + `/api"}`))
	if err != nil {
		t.Fatalf("NewFromSecret() error = %v", err)
	}

	if _, err := c.GetSavedGroup(context.Background(), "sg_missing"); !IsNotFound(err) {
		t.Errorf("GetSavedGroup(missing) err = %v, want not found", err)
	}
}
