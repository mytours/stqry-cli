# Releasing

## Quick release

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

Pushing a tag triggers the [release workflow](.github/workflows/release.yml), which builds binaries for all platforms, signs and notarizes the macOS binaries, and publishes a GitHub Release.

## What happens

1. All tests run
2. `quill` is installed on the CI runner (supports signing macOS binaries from Linux — no Mac needed)
3. GoReleaser builds binaries for linux, darwin, and windows × amd64/arm64
4. After each darwin binary is built, a hook calls `quill sign-and-notarize` which:
   - Signs the binary with the Developer ID Application certificate
   - Submits it to Apple's notary service and waits for acceptance
5. Signed, notarized binaries are archived and uploaded to the GitHub Release
6. `checksums.txt` is generated and attached

## Versioning

Follow [semver](https://semver.org/). Pre-1.0: minor bumps for features, patch bumps for fixes.

## CI secrets

These must be set before the first signed release. Go to **Settings → Environments → release** (create the `release` environment if it doesn't exist), then add:

| Secret | How to get it |
|--------|--------------|
| `MACOS_SIGN_P12` | Base64-encoded Developer ID Application certificate (.p12) — see below |
| `MACOS_SIGN_PASSWORD` | Password used when exporting the .p12 |
| `MACOS_NOTARY_KEY` | Base64-encoded App Store Connect API key (.p8) — see below |
| `MACOS_NOTARY_KEY_ID` | 10-character key ID shown in App Store Connect |
| `MACOS_NOTARY_ISSUER_ID` | Issuer UUID shown in App Store Connect (under the key) |

## One-time Apple setup

You need a paid Apple Developer account ($99/year at [developer.apple.com](https://developer.apple.com)).

### 1. Create a Developer ID Application certificate

1. Open **Keychain Access** → Certificate Assistant → Request a Certificate from a Certificate Authority
   - Save the CSR to disk
2. Go to [developer.apple.com/account](https://developer.apple.com/account) → Certificates, Identifiers & Profiles → Certificates → +
3. Choose **Developer ID Application** → continue → upload the CSR → download the `.cer`
4. Double-click the `.cer` to install it in Keychain Access
5. In Keychain Access, find the "Developer ID Application: Your Name (TEAMID)" cert, right-click → Export → save as `cert.p12`, set a strong password
6. Base64-encode it:
   ```bash
   base64 -i cert.p12 | pbcopy
   ```
7. Paste the result as the `MACOS_SIGN_P12` secret; use the export password for `MACOS_SIGN_PASSWORD`

### 2. Create an App Store Connect API key

1. Go to [App Store Connect](https://appstoreconnect.apple.com) → Users and Access → Integrations → App Store Connect API
2. Click **+** to generate a new key; role: **Developer** is sufficient
3. Download the `.p8` key file (you can only download it once)
4. Note the **Key ID** (10 chars) and **Issuer ID** (UUID) shown on that page
5. Base64-encode the key:
   ```bash
   base64 -i AuthKey_XXXXXXXXXX.p8 | pbcopy
   ```
6. Paste the result as `MACOS_NOTARY_KEY`; set `MACOS_NOTARY_KEY_ID` and `MACOS_NOTARY_ISSUER_ID` from the values on the page

## Verifying a signed binary locally

```bash
# Check the signature
codesign -dv --verbose=4 ./stqry

# Check notarization (requires internet — Apple's servers)
spctl -a -t exec -vv ./stqry
```

The output should show `source=Notarized Developer ID` for a properly signed and notarized binary.

## How signing works (no Mac required)

[`anchore/quill`](https://github.com/anchore/quill) signs macOS Mach-O binaries using the raw certificate and key material — no macOS keychain needed. It runs on Linux, which means our Ubuntu CI runner can sign and notarize darwin binaries as part of the normal GoReleaser build. The signing hook in `.goreleaser.yaml` runs after each darwin binary is compiled, before it's archived.
