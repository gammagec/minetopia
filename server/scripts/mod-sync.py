#!/usr/bin/env python3
"""Download mods listed in mods.yml to the target directory."""

import hashlib
import os
import sys
import requests
import yaml


MODRINTH_API = "https://api.modrinth.com/v2"


def sync_mods(config_path: str, mods_dir: str, side: str = "server") -> None:
    with open(config_path) as f:
        config = yaml.safe_load(f)

    os.makedirs(mods_dir, exist_ok=True)

    mods = config.get("mods", [])
    print(f"Syncing {len(mods)} mod(s) for side={side} -> {mods_dir}")

    for mod in mods:
        mod_side = mod.get("side", "both")
        if mod_side == "client" and side == "server":
            continue
        if mod_side == "server" and side == "client":
            continue

        source = mod.get("source")
        if source == "modrinth":
            sync_modrinth(mod, mods_dir)
        elif source == "url":
            sync_url(mod, mods_dir)
        else:
            print(f"  [WARN] Unknown source '{source}' for {mod.get('name')}, skipping")


def sync_modrinth(mod: dict, mods_dir: str) -> None:
    project_id = mod["project_id"]
    version_num = mod["version"]

    resp = requests.get(
        f"{MODRINTH_API}/project/{project_id}/version",
        params={"version_number": version_num},
        headers={"User-Agent": "minetopia-server/1.0"},
    )
    resp.raise_for_status()
    versions = resp.json()

    if not versions:
        raise RuntimeError(f"Version {version_num} not found for project {project_id}")

    files = versions[0]["files"]
    file = next((f for f in files if f.get("primary")), files[0])

    dest = os.path.join(mods_dir, file["filename"])
    expected_hash = file.get("hashes", {}).get("sha512")

    if os.path.exists(dest) and expected_hash:
        with open(dest, "rb") as f:
            if hashlib.sha512(f.read()).hexdigest() == expected_hash:
                print(f"  [OK] {mod['name']} already up to date")
                return

    print(f"  [DL] {mod['name']} -> {file['filename']}")
    r = requests.get(file["url"], headers={"User-Agent": "minetopia-server/1.0"})
    r.raise_for_status()

    if expected_hash and hashlib.sha512(r.content).hexdigest() != expected_hash:
        raise RuntimeError(f"Hash mismatch for {file['filename']}")

    with open(dest, "wb") as f:
        f.write(r.content)
    print(f"  [OK] {mod['name']}")


def sync_url(mod: dict, mods_dir: str) -> None:
    url = mod["url"]
    filename = mod.get("filename", url.split("/")[-1])
    dest = os.path.join(mods_dir, filename)

    if os.path.exists(dest):
        print(f"  [OK] {mod['name']} already present")
        return

    print(f"  [DL] {mod['name']} -> {filename}")
    r = requests.get(url)
    r.raise_for_status()
    with open(dest, "wb") as f:
        f.write(r.content)
    print(f"  [OK] {mod['name']}")


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <mods.yml> <mods-dir> [server|client]")
        sys.exit(1)

    side_arg = sys.argv[3] if len(sys.argv) > 3 else "server"
    sync_mods(sys.argv[1], sys.argv[2], side_arg)
    print("Mod sync complete.")
