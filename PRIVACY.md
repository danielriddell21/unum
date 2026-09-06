# Privacy

How unum handles your data, what it stores, and what it sends over the network.

## Summary

* **Your file contents never leave your machine.** No tool uploads, proxies, or
  round-trips the data you point it at. Every parser, differ, renderer, and
  encoder runs in-process.
* **Anonymous usage telemetry is on by default** and sends counts and sizes
  only — never file contents, filenames, paths, or arguments. Turn it off with
  `--no-telemetry`, `DO_NOT_TRACK=1`, or a config setting.
* **Two files are written to `~/.config/unum/`**, both `0600` inside a `0700`
  directory. One of them, the `hash` history, stores the strings you passed to
  `unum hash` in plaintext.

## Where your data goes

### Content stays local

Every tool processes input in memory and writes only what you ask for:

| Tool | Processing |
|---|---|
| `unum json` | parsed in-process; browser UI served from the embedded filesystem |
| `unum diff` | both sides read locally, diffed in-process |
| `unum hash` | SHA256 derivation, no network |
| `unum diagram` | mermaid and d2 rendered by vendored Go libraries, not a remote renderer |
| `unum image` | decoded, resized, and re-encoded locally |

The browser UIs (`--web`) are served by a local HTTP server from assets compiled
into the binary. There are no CDN references, remote fonts, or third-party
scripts, so opening a unum web UI makes no outbound requests.

The web APIs take content in the request body, never a server-side file path —
no endpoint will read an arbitrary file off disk for a caller.

### Files written to disk

Both live in `os.UserConfigDir()/unum` (`~/.config/unum` on Linux,
`~/Library/Application Support/unum` on macOS), mode `0600` in a `0700`
directory:

* **`config.json`** — theme preferences, your telemetry preference, and a
  randomly generated `client_id` (a UUID, created on first run).
* **`hash-history.json`** — the **50 most recent inputs to `unum hash`, stored
  in plaintext** alongside timestamps, so the TUI and web UI can offer history.
  Every `unum hash <text>` invocation appends to this file, including plain
  one-shot CLI runs.

  `unum hash` is a deterministic deriver — it maps a name like `staging-api` to
  a stable port, UUID, and colour. It is **not** a password tool and is not
  intended for secrets. If you do pass something sensitive, it lands in this
  file; delete `hash-history.json` to clear it.

Delete either file at any time; both are regenerated with defaults.

### What telemetry sends

Telemetry is OpenTelemetry traces and metrics sent over OTLP/HTTP to an endpoint
baked in at release build time. Each invocation reports:

* tool name and output mode (`static`, `ui`, `web`)
* `GOOS`, `GOARCH`, and unum version
* the **names** of the flags you used — never their values
* input size in bytes, node/line counts, and processing duration
* an error *category* (`read`, `parse`, `render`, `query`) on failure
* the `client_id` from your config file, and `deployment.environment`

It does **not** send file contents, filenames, paths, argument values, query
expressions, hostnames, usernames, or environment variables.

Note that `client_id` is a stable identifier: it makes the data pseudonymous
rather than strictly anonymous, since invocations from one machine can be
correlated with each other. It is not linked to any account or identity.

### Turning telemetry off

Any one of these disables it:

```bash
unum json data.json --no-telemetry   # per invocation
export DO_NOT_TRACK=1                # honours the consoledonottrack.com convention
export UNUM_NO_TELEMETRY=1
```

Or set it permanently in `~/.config/unum/config.json`:

```json
{ "telemetry": false }
```

Set `UNUM_TELEMETRY_DEBUG=1` to log to stderr exactly what would be sent.

## Web UI mode

`--web` binds to `localhost` on a random free port and opens your browser. The
data you loaded stays in that process's memory for its lifetime.

**If the `PORT` environment variable is set, the server binds to `0.0.0.0`
instead** — every interface, with no authentication. This is intended for the
hosted deployment, but `PORT` is commonly exported by unrelated dev tooling, so
check your environment before running `--web` on a shared or untrusted network.
The Prometheus `/metrics` endpoint is exposed on the same listener.

Uploads made through the JSON web UI are held in an in-memory cache for the
lifetime of the process, addressed by a 64-bit random key. They are not written
to disk, and are gone when the process exits.

## Image metadata

`unum image` strips metadata. EXIF orientation is applied to the pixels so
rotated photos stay upright, and everything else — including GPS coordinates,
camera serial numbers, and timestamps — is dropped on re-encode. Optimised
output is safe to share in a way the original often is not.

The one exception is the web UI, which keeps the original bytes in memory (max 8
images) so the browser can display the source you picked. Those bytes still
carry their original metadata; they are never written to disk and never leave
the local server.

## Known gaps

Tracked, not yet fixed:

* `unum hash` writes its history unconditionally — there is no `--no-history`
  flag and no built-in command to clear it.
* The JSON web upload cache has no expiry or size limit; it grows for the
  lifetime of the process.
* The `0.0.0.0` bind is triggered by the presence of `PORT` alone, which is
  easy to hit by accident.

## Questions

Open an issue at https://github.com/danielriddell21/unum/issues.
