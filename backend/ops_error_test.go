package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpsWrapKeepsSentinel(t *testing.T) {
	err := wrapOps("store.get", "ops-404", ErrOpsNotFound)
	if !errors.Is(err, ErrOpsNotFound) {
		t.Fatalf("wrapped error lost sentinel: %v", err)
	}
}

func TestOpsCodeClassifiesWrapped(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{fmt.Errorf("db: %w", ErrOpsNotFound), "not_found"},
		{fmt.Errorf("db: %w", ErrOpsConflict), "conflict"},
		{fmt.Errorf("db: %w", ErrOpsInvalid), "invalid"},
		{fmt.Errorf("db: %w", ErrOpsTransition), "transition"},
		{fmt.Errorf("db: %w", ErrOpsPolicy), "policy"},
	}
	for _, c := range cases {
		if got := opsCode(c.err); got != c.want {
			t.Fatalf("opsCode(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}

func newOpsTestRouter() http.Handler {
	return newOpsRouter(newOpsService(seedOpsRecords()))
}

func TestOpsHTTPTransitionMissing404(t *testing.T) {
	router := newOpsTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ops/records/nope/status", strings.NewReader(`{"status":"active"}`))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing record transition: got %d, want 404", rec.Code)
	}
}

func TestOpsHTTPStaleRevision409(t *testing.T) {
	router := newOpsTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ops/records/ops-201/status", strings.NewReader(`{"status":"paused","expectedRevision":99}`))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("stale revision transition: got %d, want 409", rec.Code)
	}
}

func TestOpsHTTPDuplicateCreate409(t *testing.T) {
	router := newOpsTestRouter()
	body := `{"id":"ops-201","subject":"dup","owner":"zhang","priority":"high","labels":{"site":"north"}}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ops/records", strings.NewReader(body))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate create: got %d, want 409", rec.Code)
	}
}

func TestOpsHTTPInvalidBody400(t *testing.T) {
	router := newOpsTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ops/records", strings.NewReader(`{bad json`))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid body: got %d, want 400", rec.Code)
	}
}

func TestOpsHTTPRecordInvalidStatus400(t *testing.T) {
	router := newOpsTestRouter()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/ops/records/ops-201/status", strings.NewReader(`{"status":"bogus"}`))
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid status: got %d, want 400", rec.Code)
	}
}
