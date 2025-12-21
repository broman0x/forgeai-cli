# ForgeAI CLI

[![Release](https://img.shields.io/github/v/release/broman0x/forgeai-cli)](https://github.com/broman0x/forgeai-cli/releases/latest)

ForgeAI CLI is a nimble, developer-friendly command-line tool that puts AI where you work: the terminal. Use local models via Ollama or cloud models like Gemini to chat, review code, and apply edits — fast.

![ForgeAI CLI screenshot](https://i.imgur.com/NOTp3b2.png)

_Screenshot: interactive dashboard and example output_

Why it exists
------------
Because sometimes you want quick, contextual help without leaving your editor. ForgeAI gives you:

- Instant conversation with an AI assistant
- Smart code reviews and patch-ready diffs
- A quick system dashboard (hardware, network, Ollama)

Key features
------------
- Interactive chat and single-shot prompts
- AI-driven code review with readable reports
- Edit suggestions with unified diff previews
- Automatic provider selection: Gemini (cloud) or Ollama (local)

Prerequisites
-------------
- Go 1.20+ (modules enabled)
- Optional: Ollama running locally (default: `localhost:11434`)
- For Gemini: `GEMINI_API_KEY` in your environment

Quick start
-----------
Clone, build, run:

```bash
git clone https://github.com/broman0x/forgeai-cli.git
cd forgeai-cli
./forgeai        # or `go run .`
```

Shortcuts
---------
- Windows: `forge.bat <args>`
- Linux/Mac: `./forge.sh <args>`

Common commands
---------------
- `forgeai ask "Explain this code"` — one-off prompt
- `forgeai review path/to/file.go` — AI review report
- `forgeai edit path/to/file.go "Refactor for clarity"` — proposed code edits + diff
- `forgeai info` — system & config dashboard

Configuration
-------------
ForgeAI uses Viper. Defaults are provided in code; override them with `$HOME/.forgeai/config.yaml`:

```yaml
provider: "gemini"  # or "ollama"
model: "gemini-2.5-flash"
first_run: false
```

Environment variables
- `GEMINI_API_KEY` — required for Gemini

How providers work
------------------
- Gemini: cloud model, needs `GEMINI_API_KEY` and internet access.
- Ollama: local model server; ForgeAI will detect a running Ollama and prefer it when available.

Troubleshooting
---------------
- Ollama unavailable? Check health endpoint:

```bash
curl http://localhost:11434/health
```

- Gemini errors? Verify `GEMINI_API_KEY` and network access.

Releases
--------
Get the latest release and downloadable artifacts here:

https://github.com/broman0x/forgeai-cli/releases/latest


License
-------
MIT — see `LICENSE`.
