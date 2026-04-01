# Stockyard Dispatch

**Email list and newsletter.** Manage subscribers, send campaigns via your own SMTP, track opens. Like Listmonk but actually simple to install. Single binary, no external dependencies.

Part of the [Stockyard](https://stockyard.dev) suite of self-hosted developer tools.

## Quick Start

```bash
export SMTP_HOST=smtp.example.com SMTP_PORT=587 SMTP_USER=you SMTP_PASS=pass SMTP_FROM=you@example.com
curl -sfL https://stockyard.dev/install/dispatch | sh
dispatch
```

## Usage

```bash
# Create a list
curl -X POST http://localhost:8900/api/lists \
  -H "Content-Type: application/json" \
  -d '{"name":"Newsletter"}'

# Add subscribers
curl -X POST http://localhost:8900/api/lists/{id}/subscribers \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"Jane"}'

# Embed subscribe form
# <form method="POST" action="http://localhost:8900/subscribe/{list_id}">
#   <input name="email" type="email" required>
#   <button>Subscribe</button>
# </form>

# Create and send a campaign
curl -X POST http://localhost:8900/api/lists/{id}/campaigns \
  -H "Content-Type: application/json" \
  -d '{"subject":"Weekly Update","body_html":"<h1>Hello!</h1><p>News this week...</p>"}'

curl -X POST http://localhost:8900/api/campaigns/{id}/send
```

## Free vs Pro

| Feature | Free | Pro ($4.99/mo) |
|---------|------|----------------|
| Lists | 1 | Unlimited |
| Subscribers | 100 | Unlimited |
| Sends/month | 500 | Unlimited |
| Open tracking | — | ✓ |
| Click tracking | — | ✓ |
| Send log | 7 days | 1 year |

## License

Apache 2.0 — see [LICENSE](LICENSE).
