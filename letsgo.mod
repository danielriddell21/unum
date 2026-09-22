// letsgo.mod

// The six targets the GoReleaser config built. letsgo's default matrix is
// five, so windows/arm64 is named to keep the published asset list the same.
build (
	linux/amd64
	linux/arm64
	darwin/amd64
	darwin/arm64
	windows/amd64
	windows/arm64
)

// letsgo's own ldflags are literal, so that a release is a function of its
// commit. The two telemetry values are not — they come from the environment —
// so they are injected by a plugin that records what it compiled in. What to
// inject is in letsgo-env.mod, beside this file.
//
// Pinned by the digest of the released linux/amd64 binary, which is what the
// release runner installs: a program that decides what gets built is a build
// input exactly as the compiler is.
plugin ldflags letsgo-env v0.3.0 sha256:9b8b159e6b660de41b94da0e622f8c36b060e6427e8b92659bf996d670f1c139

brew danielriddell21/tap

image ghcr.io/danielriddell21/unum

// The base the deleted Dockerfile used, pinned by digest so that two releases
// of one commit cannot differ — bump it deliberately for base fixes. It also
// supplies the nonroot user the Dockerfile asked for by tag.
image base gcr.io/distroless/static-debian12@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

// What the deleted Dockerfile declared. It set no CMD, so there is none here.
image expose 8080

// The shared GoReleaser workflow marked releases as pre-releases after
// publishing; letsgo does it while publishing, so promote.yaml still fires on
// manual promotion.
release prerelease=true
