package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func useTempInventoryState(t *testing.T) {
	t.Helper()
	orig := inventoryStatePath
	inventoryStatePath = filepath.Join(t.TempDir(), "inventory.state")
	t.Cleanup(func() { inventoryStatePath = orig })
}

func TestInventoryIsSentOnceThenOnlyWhenChangedOrStale(t *testing.T) {
	useTempInventoryState(t)
	now := time.Date(2026, 11, 22, 4, 0, 0, 0, time.UTC)
	inv := []InstalledPackage{
		{Name: "openssl", Version: "3.0.2-0ubuntu1.18", Source: "openssl", Arch: "amd64"},
		{Name: "bash", Version: "5.1-6ubuntu1.1", Source: "bash", Arch: "amd64"},
	}

	if got := inventoryToSend(inv, now, false); len(got) != 2 {
		t.Fatalf("first run must send the inventory, got %d entries", len(got))
	}
	markInventorySent(inv, now)

	// Five minutes later, unchanged (and listed in another order): not sent.
	reordered := []InstalledPackage{inv[1], inv[0]}
	if got := inventoryToSend(reordered, now.Add(5*time.Minute), false); got != nil {
		t.Fatalf("an unchanged inventory must not be sent again, got %d entries", len(got))
	}

	// A package was upgraded: sent at once.
	changed := []InstalledPackage{inv[0], {Name: "bash", Version: "5.1-6ubuntu1.2", Source: "bash", Arch: "amd64"}}
	if got := inventoryToSend(changed, now.Add(10*time.Minute), false); len(got) != 2 {
		t.Fatalf("a changed inventory must be sent")
	}

	// Unchanged for the whole interval: sent again so a lost check-in heals.
	if got := inventoryToSend(inv, now.Add(inventoryResendInterval), false); len(got) != 2 {
		t.Fatalf("an inventory older than the resend interval must be sent")
	}
	if got := inventoryToSend(inv, now.Add(inventoryResendInterval-time.Minute), false); got != nil {
		t.Fatalf("just inside the interval the inventory must not be sent")
	}

	// The install run always sends.
	if got := inventoryToSend(inv, now.Add(time.Minute), true); len(got) != 2 {
		t.Fatalf("-install must always send the inventory")
	}

	// The clock went backwards: send.
	if got := inventoryToSend(inv, now.Add(-time.Hour), false); len(got) != 2 {
		t.Fatalf("a state from the future must not suppress the inventory")
	}
}

func TestInventoryStateProblemsMeanSend(t *testing.T) {
	useTempInventoryState(t)
	now := time.Now()
	inv := []InstalledPackage{{Name: "bash", Version: "5.1"}}

	for _, content := range []string{"", "garbage", "abc notanumber", "one two three"} {
		if err := os.WriteFile(inventoryStatePath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if got := inventoryToSend(inv, now, false); len(got) != 1 {
			t.Errorf("state %q must not suppress the inventory", content)
		}
	}

	// An empty inventory (the probe failed) is never sent and never recorded,
	// so the server keeps what it has.
	if got := inventoryToSend(nil, now, false); got != nil {
		t.Errorf("an empty inventory must not be sent")
	}
	_ = os.Remove(inventoryStatePath)
	markInventorySent(nil, now)
	if _, err := os.Stat(inventoryStatePath); err == nil {
		t.Errorf("an empty inventory must not be recorded as sent")
	}

	// An unwritable state path must not break anything.
	inventoryStatePath = filepath.Join(t.TempDir(), "missing-dir", "inventory.state")
	markInventorySent(inv, now)
	if got := inventoryToSend(inv, now, false); len(got) != 1 {
		t.Errorf("without a state file the inventory is sent")
	}
}

// A stalled server must not hang the agent: later cron runs refuse to start
// while one is still running.
func TestHttpClientHasATimeout(t *testing.T) {
	c := GetHttpClient()
	if c.Timeout != httpTimeout || c.Timeout <= 0 || c.Timeout > 5*time.Minute {
		t.Fatalf("http client timeout = %s", c.Timeout)
	}
}
