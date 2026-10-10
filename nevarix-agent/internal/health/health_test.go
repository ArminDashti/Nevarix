package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthReturnsIdentity(t *testing.T) {
	srv := httptest.NewServer(Handler(Identity{Agent: "nevarix-agent", AgentID: "a1", Hostname: "h1"}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var id Identity
	if err := json.NewDecoder(res.Body).Decode(&id); err != nil {
		t.Fatal(err)
	}
	if id.Agent != "nevarix-agent" || id.AgentID != "a1" {
		t.Fatalf("unexpected identity: %+v", id)
	}
}
