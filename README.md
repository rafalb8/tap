# Tap
`tap` is a lightweight, zero-configuration HTTP analyzer for developers.

It spawns a subprocess, automatically intercepts its HTTP/HTTPS traffic, and inspects the payloads without modifying global system proxy settings.

It's like `time` or `sudo`, but for HTTP analysis.

## Features
* **Zero Global Configuration:** Only intercepts the command you specifically target.
* **HTTPS Decryption:** Automatic MITM proxying using self-generated or custom CA certificates.
* **Flexible Output:** Supports human-readable text, minimal logs (`--simple`), or structured data (`--json`).
* **Transparent Pass-Through:** Use the `-o` flag to pipe `tap`'s analysis to a file, leaving your terminal free for the wrapped program's native `stdin`, `stdout`, and `stderr`.

## Installation
```sh
go install github.com/rafalb8/tap@latest
```

## Usage
Using `tap` is simple. Append `--` followed by the command you want to analyze.
```sh
tap [flags] -- <command> [args...]
```

### Examples
#### Basic Inspection
```sh
tap -- curl https://google.com
```

#### JSON
Capture analysis into dump.json while interacting with the wrapped program normally via standard I/O:
```sh
tap -j -o dump.json -- curl -L google.com
```
or pipe it through `jq`:
```sh
tap -j -- curl -L google.com | jq
```
