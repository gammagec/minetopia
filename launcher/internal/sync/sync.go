package sync

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"minetopia/launcher/internal/config"
	"minetopia/launcher/internal/modrinth"
)

// SyncClientMods downloads all client-side mods from cfg into modsDir.
func SyncClientMods(cfg *config.Config, modsDir string) error {
	if err := os.MkdirAll(modsDir, 0o755); err != nil {
		return fmt.Errorf("creating mods dir: %w", err)
	}

	mods := cfg.ClientMods()
	fmt.Printf("Syncing %d mod(s) for MC %s/%s to %s\n", len(mods), cfg.Minecraft.Version, cfg.Minecraft.Modloader, modsDir)

	for _, mod := range mods {
		if err := syncMod(mod, modsDir, cfg.Minecraft.Version, cfg.Minecraft.Modloader); err != nil {
			return fmt.Errorf("%s: %w", mod.Name, err)
		}
	}
	return nil
}

func syncMod(mod config.Mod, modsDir, mcVersion, loader string) error {
	switch mod.Source {
	case "modrinth":
		return syncModrinth(mod, modsDir, mcVersion, loader)
	case "url":
		return syncURL(mod, modsDir)
	default:
		return fmt.Errorf("unknown source %q", mod.Source)
	}
}

func syncModrinth(mod config.Mod, modsDir, mcVersion, loader string) error {
	ver, err := modrinth.GetVersion(mod.ProjectID, mcVersion, loader, mod.Version)
	if err != nil {
		return err
	}
	file, err := ver.PrimaryFile()
	if err != nil {
		return err
	}

	dest := filepath.Join(modsDir, file.Filename)
	expectedHash := file.Hashes["sha512"]

	if upToDate(dest, expectedHash) {
		fmt.Printf("  [OK] %s %s\n", mod.Name, ver.VersionNumber)
		return nil
	}

	fmt.Printf("  [DL] %s %s ...\n", mod.Name, ver.VersionNumber)
	data, err := download(file.URL)
	if err != nil {
		return err
	}

	if expectedHash != "" {
		if got := sha512Hex(data); got != expectedHash {
			return fmt.Errorf("hash mismatch: got %s, want %s", got[:16]+"…", expectedHash[:16]+"…")
		}
	}

	return os.WriteFile(dest, data, 0o644)
}

func syncURL(mod config.Mod, modsDir string) error {
	filename := mod.Filename
	if filename == "" {
		filename = path.Base(mod.URL)
	}
	dest := filepath.Join(modsDir, filename)
	if _, err := os.Stat(dest); err == nil {
		fmt.Printf("  [OK] %s\n", mod.Name)
		return nil
	}
	fmt.Printf("  [DL] %s ...\n", mod.Name)
	data, err := download(mod.URL)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}

func download(rawURL string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, rawURL, nil)
	req.Header.Set("User-Agent", "minetopia-launcher/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, rawURL)
	}
	return io.ReadAll(resp.Body)
}

func upToDate(path, expectedHash string) bool {
	if expectedHash == "" {
		_, err := os.Stat(path)
		return err == nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return sha512Hex(data) == expectedHash
}

func sha512Hex(data []byte) string {
	h := sha512.Sum512(data)
	return hex.EncodeToString(h[:])
}
