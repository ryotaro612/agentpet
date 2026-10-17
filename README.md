# agentpet

agentpet is an MCP server and client that displays a desktop agent pet animated with spritesheets. Pets and animations are fully configurable: each spritesheet can have its own dimensions, display resolution, and frame count.

The server exposes MCP tools for controlling the pet. The companion client, `petowner`, calls those tools from the command line — suitable for wiring up to Claude Code hooks or other MCP-aware agents.

## Configuration

The default config file path is `~/.local/agentpet/config.toml`.
Override it with the `AGENT_PET_CONFIG` environment variable, or the `-c` flag (the flag takes precedence).
agentpet reloads automatically whenever the file is saved.

```toml
# MCP server port. 0 = auto-assign.
port = 60012

# Name of the pet that is active at startup.
pet = "cat"

# Global defaults — inherited by every pet and animation unless overridden.
[window]
width  = 320
height = 320

[frame]
width  = 32
height = 32

fps = 12

[[pets]]
name        = "cat"
description = "A pixel-art cat"

# Per-pet overrides (optional).
[pets.window]
width  = 320
height = 320

[pets.frame]
width  = 32
height = 32

fps = 8

[[pets.animations]]
name        = "idle"
filepath    = "/path/to/idle.png"
description = "Standing still"

# Per-animation overrides (all optional).
[pets.animations.frame]
width  = 32
height = 32

[pets.animations.window]
width  = 320
height = 320

# fps and count can also be set per animation.
fps   = 8
count = 4   # omit to compute from image size ÷ frame size
```

### Field reference

| Field | Level | Description |
|---|---|---|
| `port` | global | TCP port for the MCP server (`0` = auto) |
| `pet` | global | Active pet at startup |
| `window.width` / `window.height` | global / pet / animation | Display window size in pixels |
| `frame.width` / `frame.height` | global / pet / animation | Sprite frame size in pixels |
| `fps` | global / pet / animation | Playback speed in frames per second |
| `pets[].name` | pet | Unique pet name (required) |
| `pets[].description` | pet | Description exposed to MCP clients |
| `pets[].animations[].name` | animation | Unique animation name within a pet; derived from the filename without extension when omitted |
| `pets[].animations[].filepath` | animation | Path to the spritesheet image (required) |
| `pets[].animations[].count` | animation | Explicit frame count; omit to derive from image size |

`window`, `frame`, and `fps` cascade from global → pet → animation; the most specific value wins. `count` defaults to `(image_width / frame_width) × (image_height / frame_height)`.

## Client

`petowner` is the command-line client. It connects to a running agentpet server and calls its MCP tools.

```
petowner -p <port> [-v] <subcommand> [args]

Flags:
  -p <port>   MCP server port (required for most subcommands)
  -v          Verbose output

Subcommands:
  animation [name]        Play the named animation.
                          Omit name to shuffle to a random animation.
  pet [name] [animation]  Switch to the named pet and optionally set its
                          animation. Omit name to pick a random other pet.
  pets                    List all pets and their animations, marking the
                          active pet and animation.
  show                    Show the pet window.
  hide                    Hide the pet window.
  vet                     Validate the configuration file and report errors.
  completion zsh          Print the zsh completion script.
```

### Examples

```sh
# Shuffle to a random animation on the active pet
petowner -p 60012 animation

# Play a specific animation
petowner -p 60012 animation idle

# Switch to a specific pet
petowner -p 60012 pet cat

# Switch pet and set its animation in one call
petowner -p 60012 pet cat walk

# Check whether the config file is valid
petowner -p 60012 vet
```

## Zsh completion

Generate and load the completion script:

```zsh
source <(petowner completion zsh)
```

Tab completion fills in subcommand names, pet names, and animation names. Pet and animation names are fetched live from the running server when `-p` is specified.

```zsh
petowner -p 60012 animation <Tab>   # shows animations of the active pet
petowner -p 60012 pet <Tab>         # shows all pet names
petowner -p 60012 pet cat <Tab>     # shows animations for "cat"
```

## Running the server

**As a macOS app bundle** — launch `agentpet.app` from Finder or with `open`:

```sh
open dist/agentpet.app
```

The app reads the config file, opens the pet window, and starts the MCP server in the background.

**From the command line** — run the binary directly:

```sh
dist/agentpet [-c <config-file>] [-v]
```

| Flag | Default | Description |
|---|---|---|
| `-c <path>` | `~/.local/agentpet/config.toml` | Config file path (overrides `AGENT_PET_CONFIG`) |
| `-v` | off | Enable debug logging |

Both forms are equivalent; the CLI form is convenient for testing or running under a process supervisor.

## Building

macOS is the only supported platform (the pet window uses a native WebView).

**Requirements**

Go 1.27 or later

```sh
# Build the app bundle (dist/agentpet.app) and the petowner client (dist/petowner)
make all

# Build only the petowner client
make petowner

# Install petowner to ~/.local/bin
make install

# Run tests
make test
```

The `BINDIR` variable overrides the install destination:

```sh
make install BINDIR=/usr/local/bin
```
