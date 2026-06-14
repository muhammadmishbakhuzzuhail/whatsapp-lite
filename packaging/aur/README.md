# AUR packaging

`PKGBUILD` here builds **whatslite-git** (latest commit from source). It installs the binary, the
`.desktop` entry, and the icon, and declares its dependencies so an AUR helper resolves them
automatically.

## Try it locally (no AUR account needed)

```sh
cd packaging/aur
makepkg -si
```

This pulls `webkit2gtk-4.1` + `gtk3` (runtime) and `go`/`nodejs`/`npm`/`git` (build), compiles, and
installs `whatslite` to `/usr/bin`. Launch from your app menu or run `whatslite`.

## Publish to the AUR (so `yay -S whatslite-git` works for everyone)

You need an AUR account with an SSH key added (https://aur.archlinux.org → My Account → SSH keys).

```sh
# 1. Generate .SRCINFO (required by the AUR)
cd packaging/aur
makepkg --printsrcinfo > .SRCINFO

# 2. Clone the (empty) AUR repo for the package name
git clone ssh://aur@aur.archlinux.org/whatslite-git.git /tmp/aur-whatslite
cp PKGBUILD .SRCINFO /tmp/aur-whatslite/

# 3. Push
cd /tmp/aur-whatslite
git add PKGBUILD .SRCINFO
git commit -m "Initial import: whatslite-git"
git push
```

After that, users install with `yay -S whatslite-git` (or `paru -S whatslite-git`).

## A versioned package later

Once you cut tagged GitHub releases (e.g. `v0.1.0`), add a second package `whatslite` whose `source=()`
points at the release tarball and pins `pkgver`. `-git` tracks `main`; the versioned one tracks releases.
