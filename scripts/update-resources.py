#!/usr/bin/env python3
"""Refresh the resources internal/install embeds, and record where they came from.

The Brewfiles and the wallpaper cask list are owned upstream but compiled into
every bluefin-cli binary, and they are executable package policy rather than
cosmetic data. Identifying them by a mutable branch URL alone meant a reviewer
could see the copied diff but could not tell which upstream state produced it,
and re-running the update at the same bluefin-cli commit could quietly yield a
different binary (tuna-os/bluefin-cli#258).

So this resolves each upstream branch to the commit SHA that is current at
fetch time, downloads the files at that immutable revision, and writes
PROVENANCE.json recording the repo, path, commit and a sha256 of the bytes
actually committed. internal/install/provenance_test.go then holds the tree to
that manifest.
"""

import datetime
import hashlib
import json
import os
import sys
import urllib.error
import urllib.request

COMMON_REPO = "projectbluefin/common"
COMMON_REF = "main"
TAP_REPO = "ublue-os/homebrew-tap"
TAP_REF = "main"

DEFAULT_PATH = "system_files/shared/usr/share/ublue-os/homebrew"
BLUEFIN_PATH = "system_files/bluefin/usr/share/ublue-os/homebrew"

SHARED_FILES = [
    "ai-tools.Brewfile",
    "cli.Brewfile",
    "cncf.Brewfile",
    "experimental-ide.Brewfile",
    "fonts.Brewfile",
    "ide.Brewfile",
    "k8s-tools.Brewfile",
]
BLUEFIN_FILES = ["full-desktop.Brewfile"]

RESOURCES = "internal/install/resources"
BREWFILES = f"{RESOURCES}/brewfiles"
MANIFEST = f"{RESOURCES}/PROVENANCE.json"


def get(url, accept="application/vnd.github+json"):
    request = urllib.request.Request(url, headers={
        "Accept": accept,
        "User-Agent": "bluefin-cli-update-resources",
    })
    token = os.environ.get("GITHUB_TOKEN")
    if token:
        request.add_header("Authorization", f"Bearer {token}")
    with urllib.request.urlopen(request, timeout=60) as response:
        return response.read()


def resolve_commit(repo, ref):
    """Pin a branch to the SHA it points at right now.

    A failure here is not fatal: the update is still worth having without the
    SHA, and the digests remain exact either way. It is reported rather than
    swallowed so a run that produced weaker provenance is visible.
    """
    try:
        data = json.loads(get(f"https://api.github.com/repos/{repo}/commits/{ref}"))
        return data["sha"]
    except (urllib.error.URLError, KeyError, json.JSONDecodeError) as err:
        print(f"  ! could not resolve {repo}@{ref} to a commit: {err}", file=sys.stderr)
        return ""


def digest(data):
    return "sha256:" + hashlib.sha256(data).hexdigest()


def main():
    os.makedirs(BREWFILES, exist_ok=True)
    resources = {}

    common_sha = resolve_commit(COMMON_REPO, COMMON_REF)
    # Fetching at the resolved SHA rather than at the branch is what makes the
    # download reproducible: the branch can move between two files in one run.
    common_rev = common_sha or COMMON_REF
    print(f"Updating Brewfiles from {COMMON_REPO}@{common_rev[:12]}...")

    for name, upstream_dir in (
        [(f, DEFAULT_PATH) for f in SHARED_FILES] + [(f, BLUEFIN_PATH) for f in BLUEFIN_FILES]
    ):
        path = f"{upstream_dir}/{name}"
        url = f"https://raw.githubusercontent.com/{COMMON_REPO}/{common_rev}/{path}"
        print(f"  -> {name}")
        data = get(url, accept="text/plain")
        with open(f"{BREWFILES}/{name}", "wb") as f:
            f.write(data)
        resources[f"brewfiles/{name}"] = {
            "source_repo": COMMON_REPO,
            "source_ref": COMMON_REF,
            "source_commit": common_sha,
            "source_path": path,
            "sha256": digest(data),
        }

    tap_sha = resolve_commit(TAP_REPO, TAP_REF)
    print(f"Updating wallpaper cask list from {TAP_REPO}@{(tap_sha or TAP_REF)[:12]}...")
    listing = json.loads(get(f"https://api.github.com/repos/{TAP_REPO}/contents/Casks?ref={tap_sha or TAP_REF}"))
    casks = sorted(
        entry["name"].removesuffix(".rb")
        for entry in listing
        if entry["type"] == "file"
        and entry["name"].endswith(".rb")
        and "wallpaper" in entry["name"].lower()
    )
    if not casks:
        # An empty list would silently remove every wallpaper from the CLI, and
        # is far more likely to mean the upstream layout moved than that the
        # taps really carry no wallpapers.
        print("  ! upstream listing produced no wallpaper casks; refusing to write an empty list", file=sys.stderr)
        return 1
    payload = (json.dumps(casks, indent=2) + "\n").encode()
    with open(f"{RESOURCES}/wallpaper-casks.json", "wb") as f:
        f.write(payload)
    print(f"  -> wallpaper-casks.json ({len(casks)} casks)")
    resources["wallpaper-casks.json"] = {
        "source_repo": TAP_REPO,
        "source_ref": TAP_REF,
        "source_commit": tap_sha,
        "source_path": "Casks",
        "derivation": "filtered listing of Casks/*.rb whose name contains 'wallpaper'",
        "sha256": digest(payload),
    }

    existing = {}
    if os.path.exists(MANIFEST):
        with open(MANIFEST) as f:
            existing = json.load(f)

    manifest = {
        "_comment": existing.get("_comment", []),
        "generated_at": datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "resources": dict(sorted(resources.items())),
    }
    with open(MANIFEST, "w") as f:
        json.dump(manifest, f, indent=2)
        f.write("\n")
    print(f"  -> PROVENANCE.json ({len(resources)} entries)")
    print("Update complete!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
