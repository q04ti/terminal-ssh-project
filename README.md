# AFTERDARK — An Internet Hidden Inside SSH

```bash
ssh -p 2222 afterdark.example.com
```

> **Connect to a place that technically does not exist.**  
> AFTERDARK is an interactive, multiplayer terminal world accessed directly through SSH. No web browser, no account registration, and no client installation required—just an SSH client and a terminal.

---

## What is AFTERDARK?

AFTERDARK evokes the forgotten corner of the early-2000s internet: green phosphor glow, anonymous BBS message boards, secret numbers stations, quiet rooms, and other wanderers connected to the same shared space in real time.

When you connect, you enter a central terminal hub with real destinations:

- **THE LOBBY** — A live multiplayer gathering room. See active explorers, exchange real-time transmissions, and chat with anyone currently connected.
- **THE ARCHIVE** — A persistent repository of anonymous transmissions and lost thoughts left across time. Paginated, categorized, and preserved permanently.
- **THE WALL** — A public digital graffiti board. Carve short messages and tags into the concrete for future visitors.
- **THE VOID** — A silent chamber. Release a thought into absolute darkness without attaching any identity, IP address, or handle.
- **THE NETWORK** — Live telemetry and node topology. View active visitor counts, room distributions, archive records, and server uptime.
- **THE PROFILE** — Customize your handle, switch between 5 terminal color themes (Cyan Void, Neon Dusk, Cyber Phosphor, Amber CRT, Ice Monolith), and track your unlocked achievements.
- **THE SECRETS** — A restricted command console for discovering hidden frequencies, radio beacons, matrix memory dumps, and secret chambers.

---

## Terminal Experience & Interface

```
AFTERDARK NODE-0                                         ● ONLINE: 04
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   ▄████████    ▄████████     ███        ▄████████    ▄████████ 
  ███    ███   ███    ███ ▀█████████▄   ███    ███   ███    ███ 
  ███    ███   ███    █▀     ▀███▀▀██   ███    █▀    ███    ███ 
  ███    ███  ▄███▄▄▄         ███   ▀  ▄███▄▄▄      ▄███▄▄▄▄██▀ 
▀███████████ ▀▀███▀▀▀         ███     ▀▀███▀▀▀     ▀▀███▀▀▀▀▀   
  ███    ███   ███            ███       ███    █▄  ▀███████████ 
  ███    ███   ███            ███       ███    ███   ███    ███ 
  ███    █▀    ███           ▄████▀     ██████████   ███    ███ 

   An internet hidden inside SSH. Forgotten wires and quiet corners.

  ► [1] THE LOBBY      — Join real-time conversation with active explorers
    [2] THE ARCHIVE    — Read and preserve anonymous transmissions across time
    [3] THE WALL       — Public graffiti board for persistent visitor tags
    [4] THE VOID       — Quiet room to release thoughts into absolute silence
    [5] THE NETWORK    — Live telemetry, node topology, and activity metrics
    [6] THE PROFILE    — Personalize handle, terminal theme, view badges
    [7] THE SECRETS    — Terminal command interface for hidden protocols
    [Q] DISCONNECT     — Sever SSH connection gracefully

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Use [↑/↓] or [1-7] to select destination  •  [Enter] Confirm  •  [Q] Disconnect
```

---

## Quick Start / Connecting

You don't need to install any new apps. Use your operating system's built-in terminal:

### Linux / macOS
```bash
ssh -p 2222 afterdark.example.com
```

### Windows (PowerShell / Windows Terminal / Command Prompt)
```bash
ssh -p 2222 afterdark.example.com
```

### Optional: Connecting with your SSH Key
If you connect using your SSH key (`ssh -i ~/.ssh/id_ed25519 -p 2222 ...`), AFTERDARK hashes your public key into a secure, non-reversible fingerprint (`key:SHA256:...`) to preserve your nickname, themes, and badges across visits. If you connect without a key, you receive a temporary session identity.

---

## Navigation & Controls

| Screen | Keybindings | Description |
|---|---|---|
| **Global** | `Esc` / `q` | Return to Central Hub or previous menu |
| **Boot** | `Enter` | Accept generated handle or input custom nickname |
| **Hub** | `↑` / `↓` or `1-7` | Navigate destinations; press `Enter` to enter |
| **Lobby** | `Enter` / `PgUp` / `PgDn` | Send chat message; scroll history; `Ctrl+M` mutes chat |
| **Archive** | `N` / `P` / `W` | Next page / Previous page / Write new note |
| **Wall** | `N` / `P` / `W` | Browse graffiti tags / Carve a new tag |
| **The Void** | `Enter` / `R` | Cast thought into void / Reveal random whispers |
| **Network** | `R` | Refresh telemetry metrics & room population |
| **Profile** | `N` / `S` / `T` | Change Nickname / Change Status / Cycle Themes |
| **Secrets** | Commands | Type `help`, `beacon`, `matrix`, `chamber`, `glitch`, etc. |

---

## Architecture & Technology Stack

- **Go 1.24+**: High-concurrency single binary.
- **Charmbracelet Wish**: Modern SSH server engine managing encrypted connections, host keys, and PTY allocation.
- **Charmbracelet Bubble Tea**: Elm-architecture terminal UI framework handling view transitions, key events, and resize messages.
- **Charmbracelet Lip Gloss**: Dynamic color palettes, themes, and responsive terminal layout formatting.
- **Embedded SQLite (Pure Go)**: Lightweight persistent storage running in WAL mode with connection pooling. Zero CGO dependencies.
- **Multiplayer Hub**: Dedicated background coordinator distributing messages and presence events via non-blocking Go channels.

---

## Security & Hardening

1. **Zero Shell Access**: Incoming SSH connections launch the Bubble Tea application directly. Visitors never receive an interactive shell, bash/sh prompt, or filesystem access.
2. **Terminal Injection Protection**: All visitor text (nicknames, chat messages, notes, and tags) is filtered to strip ANSI escape sequences, OSC/CSI payloads, and control characters before storage or rendering.
3. **Rate Limiting**: Built-in token bucket rate limiting prevents message flooding in chat and abuse on the wall and archive.
4. **Persistent Host Key**: Automatically creates and maintains a persistent ED25519 host key (`afterdark_ed25519`) with `0600` permissions.
5. **Privacy First**: The Void does not store IP addresses, handles, or SSH fingerprints.

---

## Local Development

### Prerequisites
- Go 1.24+
- OpenSSH client

### Build & Run
```bash
# Clone the repository
git clone https://github.com/your-username/afterdark.git
cd afterdark

# Run tests
go test -v ./internal/...

# Run locally on port 2222
PORT=2222 go run ./cmd/afterdark
```

Connect in another terminal window:
```bash
ssh -p 2222 127.0.0.1
```

---

## Production Deployment

### Option 1: Docker / Docker Compose

```bash
docker compose up -d --build
```

### Option 2: Linux Systemd Service

1. Run the deployment script:
   ```bash
   sudo ./scripts/deploy.sh
   ```
2. Open port 2222 in your firewall:
   ```bash
   sudo ufw allow 2222/tcp
   ```
3. Check daemon status and logs:
   ```bash
   systemctl status afterdark
   journalctl -u afterdark -f
   ```

---

## License

MIT License — see [LICENSE](file:///LICENSE) for details.
