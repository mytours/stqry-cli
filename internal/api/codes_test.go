package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/codes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"codes": []interface{}{
				map[string]interface{}{"id": 1, "coupon_code": "SAVE10"},
				map[string]interface{}{"id": 2, "coupon_code": "SAVE20"},
			},
			"meta": map[string]interface{}{
				"page": 1, "pages": 1, "per_page": 30, "count": 2,
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	codes, meta, err := ListCodes(c, nil)
	if err != nil {
		t.Fatalf("ListCodes: %v", err)
	}
	if len(codes) != 2 {
		t.Errorf("expected 2 codes, got %d", len(codes))
	}
	if meta == nil {
		t.Error("expected non-nil meta")
	} else if meta.Count != 2 {
		t.Errorf("expected meta.Count=2, got %d", meta.Count)
	}
}

func TestGetCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/codes/5" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": map[string]interface{}{"id": 5, "coupon_code": "SAVE10"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	code, err := GetCode(c, "5")
	if err != nil {
		t.Fatalf("GetCode: %v", err)
	}
	if code["coupon_code"] != "SAVE10" {
		t.Errorf("expected coupon_code=SAVE10, got %v", code["coupon_code"])
	}
}

func TestCreateCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/codes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["code"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"code\": %v", body)
		}
		if body["coupon_code"] != "NEWCODE" {
			t.Errorf("expected coupon_code=NEWCODE, got %v", body["coupon_code"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": map[string]interface{}{"id": 5, "coupon_code": "NEWCODE"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	code, err := CreateCode(c, map[string]interface{}{"coupon_code": "NEWCODE"})
	if err != nil {
		t.Fatalf("CreateCode: %v", err)
	}
	if code["coupon_code"] != "NEWCODE" {
		t.Errorf("expected coupon_code=NEWCODE, got %v", code["coupon_code"])
	}
}

func TestUpdateCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/codes/5" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if _, wrapped := body["code"]; wrapped {
			t.Errorf("expected flat body, got body wrapped under \"code\": %v", body)
		}
		if body["coupon_code"] != "UPDATED" {
			t.Errorf("expected coupon_code=UPDATED, got %v", body["coupon_code"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": map[string]interface{}{"id": 5, "coupon_code": "UPDATED"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	code, err := UpdateCode(c, "5", map[string]interface{}{"coupon_code": "UPDATED"})
	if err != nil {
		t.Fatalf("UpdateCode: %v", err)
	}
	if code["coupon_code"] != "UPDATED" {
		t.Errorf("expected coupon_code=UPDATED, got %v", code["coupon_code"])
	}
}

func TestDeleteCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/public/codes/5" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-token")
	if err := DeleteCode(c, "5"); err != nil {
		t.Fatalf("DeleteCode: %v", err)
	}
}
