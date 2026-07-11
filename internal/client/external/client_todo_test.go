//go:build todo

package external_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go_sql_mid_trainer_v2/internal/client/external"
	"go_sql_mid_trainer_v2/internal/domain"
)

func TestGetRiskProfileTODO(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/external/risk/1" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"user_id":1,"level":"low","score":10}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	client := external.New(srv.URL, srv.Client())
	profile, err := client.GetRiskProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetRiskProfile: %v", err)
	}
	if profile.UserID != 1 || profile.Level != "low" || profile.Score != 10 {
		t.Fatalf("unexpected profile: %+v", profile)
	}

	_, err = client.GetRiskProfile(context.Background(), 0)
	if !errors.Is(err, domain.ErrWrongID) {
		t.Fatalf("expected ErrWrongID, got %v", err)
	}

	_, err = client.GetRiskProfile(context.Background(), 404)
	if err == nil {
		t.Fatalf("expected bad status error")
	}
}
