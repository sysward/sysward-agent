package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The installed inventory is ~80 bytes per package, so a few hundred kilobytes
// for a typical server, and it rarely changes between two check-ins five
// minutes apart. It is sent only when it changed since the last time it was
// sent, or when inventoryResendInterval has passed. The server keeps the last
// inventory it received when a check-in carries none.
//
// The periodic resend covers everything the agent cannot see: a check-in that
// never arrived (CheckIn does not report failures), a server that dropped the
// inventory, or a host that was deleted and registered again.
const inventoryResendInterval = 6 * time.Hour

// inventoryStatePath holds "<sha256> <unix seconds>" of the last inventory
// sent. A variable so tests can use a temporary file.
var inventoryStatePath = "/opt/sysward/bin/inventory.state"

// inventoryHash identifies an inventory regardless of the order the package
// manager listed it in.
func inventoryHash(inv []InstalledPackage) string {
	lines := make([]string, 0, len(inv))
	for _, p := range inv {
		lines = append(lines, p.Name+"\t"+p.Version+"\t"+p.Source+"\t"+p.Arch)
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

// inventoryToSend returns inv when it has to be sent with this check-in and
// nil when the server already has it. force (the -install run) always sends.
func inventoryToSend(inv []InstalledPackage, now time.Time, force bool) []InstalledPackage {
	if len(inv) == 0 {
		return nil
	}
	if force {
		return inv
	}
	raw, err := os.ReadFile(inventoryStatePath)
	if err != nil {
		return inv
	}
	fields := strings.Fields(string(raw))
	if len(fields) != 2 || fields[0] != inventoryHash(inv) {
		return inv
	}
	sentAt, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return inv
	}
	age := now.Sub(time.Unix(sentAt, 0))
	// A timestamp in the future means the clock moved: send again.
	if age < 0 || age >= inventoryResendInterval {
		return inv
	}
	return nil
}

// markInventorySent records that inv went out at now. Failing to write the
// state only means the inventory is sent again next time.
func markInventorySent(inv []InstalledPackage, now time.Time) {
	if len(inv) == 0 {
		return
	}
	state := inventoryHash(inv) + " " + strconv.FormatInt(now.Unix(), 10) + "\n"
	_ = os.WriteFile(inventoryStatePath, []byte(state), 0600)
}
