# Privacy

How unum handles your data, what it stores, and what it sends over the network.

## Summary

* **Your file contents never leave your machine.** No tool uploads, proxies, or
  round-trips the data you point it at. Every parser, differ, renderer, and
  encoder runs in-process.
* **Anonymous usage telemetry is on by default** and sends counts and sizes
  only — never file contents, filenames, paths, or arguments. Turn it off with
  `unum telemetry off`, `--no-telemetry`, or `DO_NOT_TRACK`. A one-time notice
  says so on first run.
* **Up to two files are written to `~/.config/unum/`**, both `0600` inside a
  `0700` directory. One of them, the `hash` history, stores the strings you
  passed to `unum hash` in plaintext; `--no-history` skips it and
  `--clear-history` deletes it.

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
  intended for secrets.

  To keep an input out of the file, or to get rid of what is already there:

  ```bash
  unum hash --no-history staging-api   # derive without recording
  unum hash --clear-history            # delete the history file
  export UNUM_NO_HASH_HISTORY=1        # never record
  ```

  Or turn it off permanently in `config.json` with `{ "hash_history": false }`.

`config.json` is only written once there is something to store — a theme
choice, a telemetry preference, or a client id. A user who opts out before
their first run leaves nothing on disk. Delete either file at any time; both
are regenerated with defaults.

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
correlated with each other. It is not linked to any account or identity, it is
only generated while telemetry is enabled, and `unum telemetry off` deletes it.

Every span is exported — there is no sampling. For a CLI emitting one span per
invocation, sampling would only reduce collector cost, not what a given run
reveals; the opt-out is the privacy control.

### Turning telemetry off

Any one of these disables it:

```bash
unum telemetry off                   # persists the choice, deletes the client id
unum json data.json --no-telemetry   # this invocation only
export DO_NOT_TRACK=1                # any non-empty value; consoledonottrack.com
export UNUM_NO_TELEMETRY=1           # any non-empty value
```

Or set it directly in `~/.config/unum/config.json`:

```json
{ "telemetry": false }
```

`unum telemetry status` shows the current state, which setting decided it, the
endpoint in use, and your stored client id. Set `UNUM_TELEMETRY_DEBUG=1` to log
to stderr exactly what would be sent.

## Web UI mode

`--web` binds to `localhost` on a random free port and opens your browser. The
data you loaded stays in that process's memory for its lifetime.

Binding beyond loopback is opt-in, and unum prints a warning to stderr whenever
it happens:

```bash
UNUM_BIND=0.0.0.0 unum json data.json --web   # explicit, every interface
```

`PORT` on its own only chooses the port — it does not change the interface,
because unrelated dev tooling exports it and a stray value must never publish
your data to the network. The hosted deployment binds publicly by setting
`PORT` together with `UNUM_ENV`.

The Prometheus `/metrics` endpoint is registered only where something scrapes
it: a public bind, or an explicit local `UNUM_METRICS=1`. Note that any server
reachable beyond loopback is unauthenticated — anyone who can reach the port can
read the document you loaded.

Uploads made through the JSON web UI are held in an in-memory cache addressed by
a 64-bit random key, capped at 32 documents and expiring 30 minutes after
upload. They are never written to disk, and are gone when the process exits.

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

* Release binaries carry the telemetry collector's bearer token, injected at
  build time. Anyone with a binary can extract it and write to the collector.
  This affects the integrity of our metrics, not your data — nothing about you
  is readable from the token.

## Questions

Open an issue at https://github.com/danielriddell21/unum/issues.
