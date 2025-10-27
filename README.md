## Zeit

[![SEGV 
LICENSE](https://img.shields.io/static/v1?label=SEGV%20LICENSE&message=1.0&labelColor=0060A8&color=ffffff)](https://xn--gckvb8fzb.com/segv/)

![zeit](.README.md/zeit.webp)

[<img src="https://xn--gckvb8fzb.com/images/chatroom.png" width="275">](https://xn--gckvb8fzb.com/contact/)

_Zeit, erfassen_. A command line tool for tracking time spent on tasks &
projects.

[Download the latest version for macOS, Linux, FreeBSD, NetBSD, OpenBSD & Plan9
here](https://github.com/mrusme/zeit/releases/latest).

## Build

```sh
make
```

**Info**: This will build using the version 0.0.0. You can prefix the `make`
command with `VERSION=x.y.z` and set `x`, `y` and `z` accordingly if you want
the version in `zeit --help` to be a different one.

## Understand

_Zeit_'s data structure contains of the following key entities:

- `block`: A tracked time period that has a start and an end date/time
- `project`: A top-level project
- `task`: A task underneath a project or another task

A `block` consists of a `project` and a `task`. These don't have to pre-exist
and can be created on-the-fly when starting to track a new `block`. They can be
configured using the `zeit project` and the `zeit task` commands.

## Use

![zeit usage](.README.md/zeit.gif)

### Auto-Completion

_Zeit_ can generate auto-completion scripts for your shell of choice. You can
load completions into your current session via:

```sh
source <(zeit completion $(basename "$SHELL"))
```

(supported shells are `bash`, `zsh`, and `fish`; For PowerShell see below)

To load completions for every new session, add them to your completions
directory.

For Bash:

```sh
zeit completion bash > ~/.local/share/bash-completion/completions/zeit
```

For Zsh:

```sh
zeit completion zsh > ~/.local/share/zsh/completions/zeit
```

For Fish:

```sh
zeit completion fish > ~/.config/fish/completions/zeit.fish
```

For PowerShell:

```sh
zeit completion powershell | Out-String | Invoke-Expression
```

## Integrations

This is a list of integrations and extensions that work with _Zeit_:

| Integration | Description | Author |
| ----------- | ----------- | ------ |
| TODO        | TODO        | TODO   |
