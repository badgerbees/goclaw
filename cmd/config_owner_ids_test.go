package cmd

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/nextlevelbuilder/goclaw/internal/config"
)

func TestUpdateConfigOwnerIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := config.Default()
	original.Gateway.Port = 19999
	original.Gateway.Token = "gateway-token"
	original.Gateway.OwnerIDs = []string{"old-owner"}
	if err := config.Save(path, original); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if err := updateConfigOwnerIDs(path, []string{" alice ", "system", "bob", "alice"}); err != nil {
		t.Fatalf("update owner ids: %v", err)
	}

	updated, err := loadConfigFileForEdit(path)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if !reflect.DeepEqual(updated.Gateway.OwnerIDs, []string{"alice", "bob"}) {
		t.Fatalf("owner ids = %v, want [alice bob]", updated.Gateway.OwnerIDs)
	}
	if updated.Gateway.Port != 19999 {
		t.Fatalf("gateway port = %d, want 19999", updated.Gateway.Port)
	}
	if updated.Gateway.Token != "gateway-token" {
		t.Fatalf("gateway token = %q, want preserved token", updated.Gateway.Token)
	}
}

func TestUpdateConfigOwnerIDs_RejectsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	original := config.Default()
	original.Gateway.OwnerIDs = []string{"old-owner"}
	if err := config.Save(path, original); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if err := updateConfigOwnerIDs(path, []string{" ", ""}); err == nil {
		t.Fatal("expected error for empty owner list")
	}

	unchanged, err := loadConfigFileForEdit(path)
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	if !reflect.DeepEqual(unchanged.Gateway.OwnerIDs, []string{"old-owner"}) {
		t.Fatalf("owner ids changed unexpectedly: %v", unchanged.Gateway.OwnerIDs)
	}
}
