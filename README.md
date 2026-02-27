# git-share

Securely share git patches with E2E encryption. Like [croc](https://github.com/schollz/croc), but for git diffs.

- 🔒 **E2E Encrypted** — XChaCha20-Poly1305, key derived from passphrase via HKDF
- 💥 **One-time use** — patch is destroyed after the first download
- ⏰ **Auto-expiry** — patches expire after a configurable TTL (default 1h)
- 🚀 **Single binary** — no runtime dependencies
- 🔐 **Zero knowledge** — the relay server only sees ciphertext

## Installation

### 🌐 Cross-Platform (Go)
The easiest way to install `git-share` globally if you have Go installed:
```bash
go install github.com/flawiddsouza/git-share@latest
```

---

### 🪟 Windows (Scoop)
If you use [Scoop](https://scoop.sh/):
```powershell
scoop install https://raw.githubusercontent.com/flawiddsouza/git-share/main/git-share.json
```

---

### 🐧 Linux & 🍎 macOS
Use the universal install script:
```bash
curl -sSf https://raw.githubusercontent.com/flawiddsouza/git-share/main/install.sh | sh
```

## Quick Start

```bash
# Build
go build -o git-share .

# Start the relay server (in one terminal)
./git-share serve

# Send uncommitted changes (in your repo)
./git-share send
# Output: git-share receive k7Xm9pQ2wR-alpha-bravo-charlie-delta

# Receive and apply (in another repo)
./git-share receive k7Xm9pQ2wR-alpha-bravo-charlie-delta
```

## Usage

### Sending

```bash
git-share send                   # uncommitted changes
git-share send --staged          # staged changes only
git-share send <commit-ref>      # specific commit (e.g. abc1234)
git-share send <range>           # commit range (e.g. HEAD~3.. or main..feature)
git-share send --ttl 15m         # custom expiry (default: 1h)
```

### Receiving

```bash
git-share receive <code>          # download, decrypt, and apply to working tree
git-share receive <code> --commit # apply as a commit (git am style)
```

### Self-hosting the relay

```bash
git-share serve                   # default port 3141
git-share serve --port 8080       # custom port
git-share serve --max-ttl 2h      # max allowed TTL

# Point clients at your relay (default: http://localhost:3141)
git-share send --server https://my-relay.example.com
```

## How It Works

1. **Sender** collects git changes (`git diff` / `git format-patch`)
2. Generates a random code: `<codeId>-<passphrase>` (e.g. `k7Xm9pQ2wR-alpha-bravo-charlie-delta`)
3. Derives an encryption key from the passphrase using HKDF-SHA256
4. Encrypts the patch with XChaCha20-Poly1305
5. Uploads the encrypted blob to the relay (keyed by `codeId`)
6. **Receiver** downloads the blob, derives the same key, decrypts, and applies

The relay server **never sees the passphrase** — only the `codeId` and opaque ciphertext. The blob is deleted immediately after download.

## Security

| Property | Implementation |
|----------|---------------|
| Encryption | XChaCha20-Poly1305 (24-byte nonce) |
| Key derivation | HKDF-SHA256 with fixed salt |
| Passphrase | 4 random words from 256-word diceware list |
| Server trust | Zero-knowledge — server stores only ciphertext |
| One-time use | Blob deleted after first GET |
| Expiry | Configurable TTL, default 1 hour |
