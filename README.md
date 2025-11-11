# auto-git

Auto-git is a CLI assistant that scans your repository, summarizes pending changes, asks an Ollama model for a Conventional Commit subject, lets you edit the result, and then stages, commits, and pushes everything in one go.

## Features
- Git-aware change scanner that prints staged and unstaged summaries with per-file diff stats.
- Conventional Commit single-line subject generation powered by your Ollama server (default model `llama3.2`).
- Interactive TUI for model selection and last-second commit message edits.
- Applies `git add -A`, `git commit -m "<msg>"`, and `git push` for you after you approve the message.
- Persistent config stored in `~/.config/auto-git/config.yaml`.

## Prerequisites
- Go 1.25+ (matches `go.mod`).
- Git installed and a repository with uncommitted changes.
- Access to an Ollama server that exposes the HTTP API. By default the client targets `http://219.147.100.43:11434`; change `DefaultBaseURL` in `internal/ollama/client.go` or pass a custom value into `ollama.NewClient` if you fork the project.
- Permission to push to the current repo’s default remote.

## Installation
Clone the repo and either build locally or install to your Go bin directory:

```bash
git clone <repo-url> auto-git
cd auto-git

# Option A: build into ./bin/auto-git
make build

# Option B: install to GOPATH/bin (plus create an `ag` symlink)
make install
```

`make install` runs `go install .`, then attempts to symlink the binary to `ag` inside your `$GOBIN` for quicker access. Ensure that directory is on your `PATH`.

For a onetime run without installing, you can use `go run .` from the project root.

## Configuration
Configuration lives at `~/.config/auto-git/config.yaml`:

```yaml
model: llama3.2
```

Commands:

- `auto-git config show` – display the currently saved model.
- `auto-git config set-model <model-name>` – update the default Ollama model. The command will fetch the model list from the server and let you pick interactively if the given name is missing.

If the config file does not exist yet, auto-git falls back to `llama3.2` and will prompt you to pick a model the first time you run the tool.

## Usage
Run `auto-git` from inside any git repo with changes:

1. The tool prints a colorized summary of staged and unstaged files (counts of additions/deletions per file).
2. It fetches a unified diff (`git diff --cached` and `git diff`) and sends both the diff and a change summary to the Ollama API using the system/user prompts defined in `internal/prompt`.
3. The generated subject is trimmed to a single line (<72 chars) and must include the Conventional Commit prefix (e.g., `fix(ui): tighten validation`).
4. You get a Bubble Tea text input where you can adjust the message or replace it entirely. Press **Enter** to accept or `Esc` to cancel.
5. After confirmation, auto-git runs `git add -A`, creates the commit, and pushes to the current branch’s upstream.

If there are no pending changes, the tool exits early with an explanatory error. Any failure while committing or pushing cancels the process, so your repository state is never silently altered.

## Customizing prompts
- System prompt: `internal/prompt/builder.go` contains the guidelines used to keep subjects short and properly prefixed.
- User prompt: same file under `BuildUserPrompt`, which injects both the change summary and raw diff.

Adjusting these templates is the quickest way to change tone, structure, or additional instructions that go to your Ollama model.

## Development
- `make test` (or `go test ./...`) – run the Go unit tests.
- `make clean` – remove build artifacts.

Contributions are welcome—feel free to open issues or PRs with improvements to the workflow, prompt presets, or configuration options.
