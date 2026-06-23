package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenStoresKeyInExecutableDataFolder(t *testing.T) {
	oldExecutablePath := executablePath
	defer func() {
		executablePath = oldExecutablePath
	}()

	root := t.TempDir()
	exeDir := filepath.Join(root, "app")
	executablePath = func() (string, error) {
		return filepath.Join(exeDir, "RestDeck.exe"), nil
	}

	vault, err := Open()
	if err != nil {
		t.Fatalf("open vault: %v", err)
	}
	if vault == nil {
		t.Fatal("vault is nil")
	}

	keyPath := filepath.Join(exeDir, "Data", "secret.key")
	key, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("read key: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("key length = %d", len(key))
	}
}
