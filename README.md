# piomx

---

<p align="center">
  <a href="https://github.com/runtime-terror404/piomx/actions/workflows/release.yml"><img src="https://github.com/runtime-terror404/piomx/actions/workflows/release.yml/badge.svg" alt="CI" style="margin: 0 10px;"></a>
  <a href="https://github.com/runtime-terror404/piomx/releases"><img src="https://img.shields.io/github/v/release/runtime-terror404/piomx" alt="Release" style="margin: 0 10px;"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/runtime-terror404/piomx" alt="License" style="margin: 0 10px;"></a>
</p>

**The bridge between PlatformIO and the boards it leaves behind.**

PlatformIO makes embedded development feel frictionless — until you pick up a board it wasn't built for. Maybe it's an STM32 where CubeMX handles pin configuration perfectly but ST wants you locked into their IDE. Maybe it's an RP2350 that didn't exist when the platform definitions were written.

`piomx` gives you the PlatformIO experience on those boards — a single static binary, no Python, no venv, no fuss.

- **STM32** — Drop a CubeMX project directory, run `piomx stm32 --yes`, and get a working PlatformIO project. `.ioc` auto-detection, dynamic HCLK extraction, SWO trace script. No CubeIDE needed.

- **Pico2 / RP2350** — Full support for RP2350-based boards (WeAct, Pimoroni, official) plus the original RP2040. Arduino or pico-sdk, dual-core templates, USB + DAP environments.

One tool, one binary — `pio run --target upload` works the same no matter what's on your desk.

- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Reference](#cli-reference)
- [Generated Files](#generated-files)
- [Wizard Walkthrough](#wizard-walkthrough)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

## Installation

### One-liner

```bash
curl -fsSL https://raw.githubusercontent.com/runtime-terror404/piomx/main/install.sh | sh
```

Downloads the latest release binary for your OS/architecture to `~/.local/bin/piomx`. Safe to re-run for updates.

Windows (PowerShell):

```powershell
powershell -c "irm https://raw.githubusercontent.com/runtime-terror404/piomx/main/install.ps1 | iex"
```

### From source

```bash
git clone https://github.com/runtime-terror404/piomx.git
cd piomx
go build -o piomx ./cmd/cli
sudo cp piomx /usr/local/bin/
```

### Uninstall

```bash
./install.sh --uninstall      # Linux/macOS
./install.ps1 -Uninstall      # Windows
```

**Requirements:** PlatformIO CLI (`pio`) on PATH. That's it.

## Quick Start

```bash
# Interactive wizard (no arguments)
piomx

# Scaffold a Pico2 project with defaults
piomx pico2 --yes

# Preview what would be generated
piomx pico2 --dry-run

# Scaffold an STM32 project (auto-detects .ioc)
cd my-cubemx-project/
piomx stm32 --yes
```

## CLI Reference

### Global flags

| Flag            | Short | Default      | Description                                             |
| --------------- | ----- | ------------ | ------------------------------------------------------- |
| `--project-dir` | `-d`  | `.`          | Target project directory                                |
| `--dry-run`     | `-n`  | false        | Preview files without writing                           |
| `--yes`         | `-y`  | false        | Skip confirmation prompts                               |
| `--force`       | —     | false        | Overwrite even if files were edited                     |
| `--adopt`       | —     | false        | Lock-file an existing project without modifying content |
| `--name`        | —     | dir basename | Project name                                            |
| `--preset`      | `-p`  | —            | Load configuration from a saved preset                  |

### `piomx` (no subcommand)

Launches the interactive wizard. Walks you through platform, board, debug probe, and all options, then shows a summary before generating files.

### `piomx pico2`

Scaffold a Raspberry Pi Pico / RP2350 / RP2040 project.

```
piomx pico2 [FLAGS]
```

| Flag                 | Short | Default          | Description                                                               |
| -------------------- | ----- | ---------------- | ------------------------------------------------------------------------- |
| `--board`            | `-b`  | `weact`          | Board variant: `pico`, `pico2`, `weact`, `official`, `pimoroni`, `custom` |
| `--framework`        | —     | `arduino`        | Framework: `arduino`, `pico-sdk`                                          |
| `--core`             | —     | `earlephilhower` | Arduino core: `earlephilhower`, `mbed`                                    |
| `--environments`     | —     | `usb,dap`        | Environments (comma-separated): `usb`, `dap`                              |
| `--baud`             | —     | `115200`         | Serial monitor baud rate                                                  |
| `--log` / `--no-log` | —     | `--log`          | Add `monitor_filters = time, log2file`                                    |
| `--libs`             | `-l`  | —                | PlatformIO libraries (e.g. `SPI, Wire, Adafruit NeoPixel`)                |
| `--git`              | —     | false            | Initialize git repo + `.gitignore` + initial commit                       |
| `--ci`               | —     | false            | Generate `.github/workflows/pio_build.yml`                                |

**Board variants**

| `--board`  | Chip    | PlatformIO `board` | Notes                        |
| ---------- | ------- | ------------------ | ---------------------------- |
| `weact`    | RP2350A | `rpipico2`         | WeAct RP2350A (default)      |
| `pico2`    | RP2350  | `rpipico2`         | Official Raspberry Pi Pico 2 |
| `pico`     | RP2040  | `pico`             | Original Raspberry Pi Pico   |
| `official` | RP2350  | `rpipico2`         | Official Pico 2 (alias)      |
| `pimoroni` | RP2350  | `pimoroni_pico2`   | Pimoroni Pico Plus 2         |
| `custom`   | RP2350  | `rpipico2`         | Custom RP2350 board          |

### `piomx stm32`

Scaffold an STM32 project from a CubeMX `.ioc` file.

```
piomx stm32 [FLAGS]
```

| Flag                 | Short | Default     | Description                                 |
| -------------------- | ----- | ----------- | ------------------------------------------- |
| `--ioc`              | —     | auto-detect | Path to `.ioc` file                         |
| `--debug`            | —     | `stlink`    | Debug probe: `stlink`, `cmsis-dap`, `jlink` |
| `--swo` / `--no-swo` | —     | `--swo`     | Generate SWO trace script                   |
| `--baud`             | —     | `115200`    | Serial monitor baud rate                    |
| `--log` / `--no-log` | —     | `--log`     | Add `monitor_filters = time, log2file`      |
| `--libs`             | `-l`  | —           | PlatformIO libraries                        |
| `--git`              | —     | false       | Initialize git repo                         |
| `--ci`               | —     | false       | Generate `.github/workflows/pio_build.yml`  |

**Debug probes**

| `--debug`   | Upload protocol | Debug tool  | OpenOCD interface         |
| ----------- | --------------- | ----------- | ------------------------- |
| `stlink`    | `stlink`        | `stlink`    | `interface/stlink.cfg`    |
| `cmsis-dap` | `cmsis-dap`     | `cmsis-dap` | `interface/cmsis-dap.cfg` |
| `jlink`     | `jlink`         | `jlink`     | `interface/jlink.cfg`     |

**.ioc handling**

`.ioc` files are auto-detected in the project directory. The file is parsed for:

- `Mcu.UserName` → PlatformIO `genericSTM32XXXXXX` board ID
- `RCC.HCLKFreq_Value` → `-traceclk` frequency in the SWO script

Without an `.ioc` file, the STM32 wizard exits with an error — create a CubeMX project first.

### `piomx presets`

Save, load, list, and delete named presets. Stored in `~/.config/piomx/presets.json`.

```bash
piomx presets save my-pico2 --board weact --baud 115200 --framework arduino
piomx presets list
piomx pico2 --preset my-pico2 --yes
piomx presets delete my-pico2
```

### Overwrite protection

`piomx` writes a `.piomx.lock.yml` into every scaffolded project. On re-run, it:

1. Compares on-disk file hashes against the lock file
2. **Hash matches** → safe to regenerate
3. **Hash differs** (you edited the file) → refuses, shows a diff summary
4. Use `--force` to overwrite anyway, `--adopt` to lock-file an existing project

## Generated Files

### pico2

```
project-dir/
├── platformio.ini          # [env] + [env:usb] + [env:dap]
├── src/
│   └── main.cpp            # Arduino dual-core or pico-sdk main()
└── .piomx.lock.yml         # SHA-256 manifest (commit this)
```

### stm32

```
project-dir/
├── platformio.ini          # [platformio] + [env] + [env:{board_id}]
├── swo_trace.py            # OpenOCD SWO trace custom target
└── .piomx.lock.yml         # SHA-256 manifest
```

STM32 projects do not get a `src/main.cpp` — PlatformIO compiles from `Core/Src/` (CubeMX sources).

### platformio.ini examples

**pico2** (`piomx pico2 --board weact`):

```ini
[env]
platform = https://github.com/maxgerhardt/platform-raspberrypi.git
board = rpipico2
framework = arduino
board_build.core = earlephilhower
board_upload.maximum_size = 16777216
monitor_speed = 115200
monitor_filters = time, log2file

; Default environment: Flashes via USB-C (1200bps touch)
[env:usb]

; Debug environment: Flashes via SWD DAPLink
[env:dap]
; upload_protocol = cmsis-dap
```

**stm32** (`piomx stm32` in CubeMX dir):

```ini
[platformio]
src_dir = Core/Src
include_dir = Core/Inc

[env]
platform = ststm32
board = genericSTM32F103C8
framework = stm32cube
upload_protocol = stlink
debug_tool = stlink
monitor_speed = 115200
monitor_filters = time, log2file
extra_scripts = swo_trace.py

[env:genericSTM32F103C8]
```

## Wizard Walkthrough

```
=== piomx interactive wizard ===

Select target platform:
  [1] pico2
  [2] stm32
Select [1]:

Select board variant:
  [1] Custom RP2350 board (custom)
  [2] Raspberry Pi Pico 2 (official) (official)
  [3] Original Raspberry Pi Pico (RP2040) (pico)
  [4] Official Raspberry Pi Pico 2 (RP2350) (pico2)
  [5] Pimoroni Pico Plus 2 (pimoroni)
  [6] WeAct RP2350A (weact)
Select [1]:

Select framework:
  [1] arduino
  [2] pico-sdk
Select [1]:

...

--- Configuration Summary ---
  Platform:   Raspberry Pi Pico / RP2350 / RP2040
  Board:      WeAct RP2350A
  Framework:  arduino
  Envs:       usb
  ...

Proceed with these settings? [Y/n]:
```

## Examples

```bash
# Minimal Pico2 project
piomx pico2 --yes

# Original Pico with custom name
piomx pico2 --board pico --name my-blink -d ~/projects/ --yes

# Pico2 with git, CI, and libraries
piomx pico2 --git --ci --libs "Adafruit NeoPixel, Wire" --yes

# STM32 with J-Link, no SWO
piomx stm32 --debug jlink --no-swo --yes

# STM32 with specific .ioc file
piomx stm32 --ioc nucleo-f411re.ioc --yes

# Dry-run preview
piomx pico2 --git --ci --libs "Wire, SPI" -n

# Preset with override
piomx pico2 --preset my-pico2 --baud 9600 --yes

# Adopt an existing project
piomx pico2 --board weact --adopt -d my-existing-project
```

## Building from source

```bash
git clone https://github.com/runtime-terror404/piomx.git
cd piomx
go build -o piomx ./cmd/cli
go test ./... -race          # 70+ tests
```

## Contributing

Bug reports and PRs welcome. Architecture:

```
cmd/cli/          # cobra commands (flag parsing + output only)
  └─▶ internal/actions/   # all business logic — never prints
        └─▶ internal/core/       # Board, DebugProbe, Registry
        └─▶ internal/ioc/        # .ioc parser + STM32 family table
        └─▶ internal/generate/   # content generation (no I/O)
        └─▶ internal/writer/     # atomic writes + drift detection
        └─▶ internal/lockfile/   # .piomx.lock.yml
```

- `cmd/cli` may only import `internal/actions` — enforced by linter
- Test coverage ≥70-90% across all packages
- Specs in `refactor_instruction/`

## License

MIT — see [LICENSE](LICENSE) for full text.

---

<sub>The original Python version of this tool is archived on the [`pre_refactor`](https://github.com/runtime-terror404/piomx/tree/pre_refactor) branch.</sub>
