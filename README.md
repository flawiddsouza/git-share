# git-share

Securely share git patches with end-to-end encryption. It's like [croc](https://github.com/schollz/croc), but specifically for git diffs.

- **E2E Encrypted**: XChaCha20-Poly1305, with keys derived via HKDF.
- **One-time use**: Patches are deleted immediately after the first download.
- **Auto-expiry**: Default 1h TTL (configurable).
- **Zero knowledge**: The relay server only sees ciphertext.
- **Single binary**: Written in Go, no runtime dependencies.

## Installation

### Go
```bash
go install github.com/flawiddsouza/git-share@latest
```

### Windows (Scoop)
```powershell
scoop install https://raw.githubusercontent.com/flawiddsouza/git-share/main/git-share.json
```

### Linux & macOS
```bash
curl -sSf https://raw.githubusercontent.com/flawiddsouza/git-share/main/install.sh | sh
```

## Quick Start

```bash
# Build (optional)
go build -o git-share .

# Start a local relay
./git-share serve

# Send changes from your repo
./git-share send
# Output: git-share receive k7Xm9pQ2wR-alpha-bravo-charlie-delta

# Receive and apply in another repo
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
git-share serve                       # default port 3141
git-share serve --port 8080           # custom port
git-share serve --max-ttl 2h          # max allowed TTL
git-share serve --max-size 50MB       # max blob size (default: 10MB)

# Use your own relay
git-share send --server https://my-relay.example.com
```

## How it works

1. **Sender** collects changes via `git diff` or `git format-patch`.
2. A random code is generated: `<codeId>-<passphrase>`.
3. An encryption key is derived from the passphrase using HKDF-SHA256.
4. The patch is encrypted with XChaCha20-Poly1305.
5. The encrypted blob is uploaded to the relay, keyed by `codeId`.
6. **Receiver** downloads the blob and decrypts it locally using the passphrase.

The relay server never sees the passphrase or the encryption key. Blobs are deleted immediately after the first successful download.

## Security

| Property | Implementation |
|----------|---------------|
| Encryption | XChaCha20-Poly1305 |
| Key Derivation | HKDF-SHA256 |
| Passphrase | 4 random words (diceware) |
| Server Trust | Zero-knowledge (ciphertext only) |
| Persistence | One-time use + TTL expiry |
