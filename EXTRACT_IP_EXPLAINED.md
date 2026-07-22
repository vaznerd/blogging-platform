# How extractIP Works — Full Explanation

## The Problem It Solves

When a user makes a request, it often goes through multiple servers before reaching your Go app:

```
User (203.0.113.50) → Cloudflare/nginx (104.16.0.1) → Your Go app
```

Go's `r.RemoteAddr` gives you **the IP of the last server that connected to you** — that's `104.16.0.1` (Cloudflare), not the real user IP. Proxies fix this by adding headers:

```
X-Forwarded-For: 203.0.113.50
```

**But anyone can set that header.** An attacker could send `X-Forwarded-For: 127.0.0.1` and pretend to be localhost. So you can't blindly trust it — you need to verify the request came from a known proxy first.

---

## The Code

```go
func (h *Handler) extractIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return host
	}

	for _, prefix := range h.trustedProxies {
		if prefix.Contains(addr) {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				parts := strings.Split(xff, ",")
				if clientIP := strings.TrimSpace(parts[0]); clientIP != "" {
					return clientIP
				}
			}
			if xri := r.Header.Get("X-Real-IP"); xri != "" {
				return strings.TrimSpace(xri)
			}
			break
		}
	}

	return host
}
```

---

## Line-by-Line Walkthrough

### Line 1 — Parse the connecting IP

```go
host, _, err := net.SplitHostPort(r.RemoteAddr)
```

`r.RemoteAddr` looks like `"104.16.0.1:52314"` (IP + port). `SplitHostPort` separates them. We only want the IP, so we discard the port with `_`.

If parsing fails (no port in the string), line 2 falls back to using the raw `r.RemoteAddr`.

---

### Line 6 — Convert to a typed IP

```go
addr, err := netip.ParseAddr(host)
```

Converts the string `"104.16.0.1"` into a `netip.Addr` value. This is needed because line 11 calls `prefix.Contains(addr)`, which requires a typed IP, not a string.

If parsing fails (malformed IP), line 7 returns the raw host string — safe fallback.

---

### Lines 9-22 — Check if the connecting IP is a trusted proxy

```go
for _, prefix := range h.trustedProxies {
    if prefix.Contains(addr) {
```

This loops through your configured trusted proxies (from `config.yaml`):

```yaml
trusted_proxies:
  - "127.0.0.1"       # localhost
  - "::1"              # localhost IPv6
  - "10.0.0.0/8"       # Docker/K8s range
  - "172.16.0.0/12"    # Docker bridge range
```

Each entry is a `netip.Prefix` — either a single IP (`127.0.0.1`) or a CIDR range (`10.0.0.0/8` means any IP from `10.0.0.0` to `10.255.255.255`). The `Contains` check asks: "Is the connecting IP inside this range?"

- **If yes** — the request came from a trusted proxy. We can safely read the headers.
- **If no** — the request came directly from a client (not through a proxy). We skip to line 25 and return the raw `r.RemoteAddr`.

---

### Lines 10-14 — Read X-Forwarded-For (first preference)

```go
if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
    parts := strings.Split(xff, ",")
    if clientIP := strings.TrimSpace(parts[0]); clientIP != "" {
        return clientIP
    }
}
```

`X-Forwarded-For` is a comma-separated list. The **first entry** is always the original client IP. Subsequent entries are proxies the request passed through.

Example:
```
X-Forwarded-For: 203.0.113.50, 70.41.3.18, 150.172.238.178
```

- `203.0.113.50` = original client
- `70.41.3.18` = first proxy (Cloudflare)
- `150.172.238.178` = second proxy (your nginx)

`strings.Split(xff, ",")` splits by comma, `parts[0]` grabs the first one (the real client IP), `strings.TrimSpace` removes any leading spaces.

---

### Lines 15-17 — Fallback to X-Real-IP

```go
if xri := r.Header.Get("X-Real-IP"); xri != "" {
    return strings.TrimSpace(xri)
}
```

If `X-Forwarded-For` is empty or missing, try `X-Real-IP`. This is a simpler header — just one IP, no comma-separated list. Some proxies (like nginx) use this instead.

---

### Line 18 — Break if no headers found

```go
break
```

If we confirmed the request is from a trusted proxy but neither header is present, we stop checking. We don't fall through to the next proxy in the list.

---

### Line 22 — Default return

```go
return host
```

If the connecting IP is NOT a trusted proxy (direct client connection), or if parsing failed, return `r.RemoteAddr` as-is. This is the safe default.

---

## Example Scenarios

| Scenario | `r.RemoteAddr` | Headers | Trusted? | Result |
|----------|----------------|---------|----------|--------|
| Direct user | `203.0.113.50:12345` | none | N/A | `203.0.113.50` |
| Through Cloudflare | `104.16.0.1:52314` | `XFF: 203.0.113.50` | Yes | `203.0.113.50` |
| Through Docker nginx | `172.17.0.1:8080` | `XFF: 203.0.113.50` | Yes | `203.0.113.50` |
| Spoofed header | `203.0.113.50:12345` | `XFF: 127.0.0.1` | No (not proxy) | `203.0.113.50` |
| Spoofed header from Docker | `172.17.0.1:8080` | `XFF: 127.0.0.1` | Yes | `127.0.0.1` (accepted because it came from trusted proxy) |

The key insight: **we only trust the header if the request actually came from a known proxy.** If someone outside the proxy tries to spoof it, we ignore the header and use their real IP.

---

## Trusted Proxies in config.yaml

```yaml
trusted_proxies:
  - "127.0.0.1"       # localhost (for dev when proxy runs on same machine)
  - "::1"              # localhost IPv6
  - "10.0.0.0/8"       # Docker/Kubernetes internal network
  - "172.16.0.0/12"    # Docker bridge network
```

### How to get your production proxy IPs

- **Cloudflare:** Published at https://www.cloudflare.com/ips/ (e.g., `173.245.48.0/20`)
- **nginx on same machine:** `127.0.0.1`
- **AWS ALB:** AWS publishes their IP ranges
- **Docker:** `172.16.0.0/12` or check `docker network inspect bridge`
- **Kubernetes:** Check your ingress controller's service IP or CIDR

### CIDR notation explained

`10.0.0.0/8` means "any IP where the first 8 bits match `10`". So:
- `10.0.0.1` ✓ matches
- `10.255.255.255` ✓ matches
- `11.0.0.1` ✗ doesn't match
- `192.168.1.1` ✗ doesn't match

A single IP like `127.0.0.1` is equivalent to `127.0.0.1/32` (all 32 bits must match).
