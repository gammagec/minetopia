package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"minetopia/launcher/internal/config"
	modsync "minetopia/launcher/internal/sync"
)

var Version = "dev"

func main() {
	var (
		configPath = flag.String("config", "mods.yml", "path to mods.yml")
		modsDir    = flag.String("mods-dir", "", "override mods directory (default: auto-detect)")
		showVer    = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("minetopia-launcher %s\n", Version)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fatalf("load config: %v", err)
	}

	dir := *modsDir
	if dir == "" {
		dir = defaultModsDir()
	}

	fmt.Printf("minetopia-launcher %s\n", Version)
	fmt.Printf("Minecraft %s (%s %s)\n", cfg.Minecraft.Version, cfg.Minecraft.Modloader, cfg.Minecraft.ModloaderVersion)

	if err := modsync.SyncClientMods(cfg, dir); err != nil {
		fatalf("sync mods: %v", err)
	}

	fmt.Println("Done! Launch Minecraft with the Fabric profile to play.")
}

func defaultModsDir() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), ".minecraft", "mods")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "minecraft", "mods")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".minecraft", "mods")
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
