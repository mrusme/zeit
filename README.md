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

## Understand

### Data structure

_Zeit_'s data structure contains of the following key entities:

- `config`: A fixed-key entity that contains user configuration
- `project`: A top-level project
- `task`: A task underneath a project or another task
- `block`: A tracked time period that has a start and an end date/time and
  references a project/task by _SID_ (_simplified ID_)
- `activeblock`: A fixed-key entity that contains the current, actively running
  block

A `block` references a `project` and a `task` by _SID_. A _SID_ is a _simplified
ID_ that
[matches a specific regex](https://github.com/mrusme/zeit/blob/master/helpers/val/val.go#L13)
and is auto-generated e.g. during import from a project's/task's display name,
by removing whitespaces and other special characters.

The project/task doesn't have to pre-exist and can be created on-the-fly when
starting to track a new `block`. They can be configured afterwards using the
`zeit project` and the `zeit task` commands.

### _Natural_ command line arguments vs. flags

_Zeit v1_ supports command line flags, yet its primary focus are _natural_
command-line arguments:

```sh
zeit start block \
  with note "Research: Coca-Cola Colombian death squads" \
  on personal/knowledge \
  4 hours ago \
  ended 10 minutes ago
```

As demonstrated by this otherwise complex example, which tracks a new _block_ of
time with a note on the _personal_ project and _knowledge_ task, starting four
hours ago and ending ten minutes ago, the use of a more _natural_ approach to
command-line arguments significantly enhances a user's understanding of the
command. However, because _Zeit_ still supports flags, the same command can also
be executed using those:

```sh
zeit start \
  --note "Research: Coca-Cola Colombian death squads" \
  --project "personal" \
  --task "knowledge" \
  --start "4 hours ago" \
  --end "10 minutes ago"
```

The structure is kept (almost) identical across various commands and can hence
be as well used for filters:

```sh
zeit blocks \
  on personal/knowledge \
  from last week \
  until two hours ago
```

This command lists all tracked time blocks for the _personal_ project and
_knowledge_ task, from last week (at this time) until two hours ago today. As
shown, the need for a detailed explanation is minimal, as the command's purpose
is easily understood just by looking at it. Similarly, as demonstrated in the
previous example, the same flags can also be used with the `blocks` command:

```sh
zeit blocks \
  --project "personal" \
  --task "knowledge" \
  --start "last week" \
  --end "two hours ago"
```

If you use _Zeit_ daily, you may find the _natural arguments_ interface more
intuitive and enjoyable than working with flags. However, if you're building a
tool that interacts with `zeit` to inject or extract data, you'll likely prefer
sticking to the more programmatically robust flags.

### Migrate from v0 to v1

**Warning**: The latest _Zeit_ version is incompatible with v0 releases
(`0.*.*`)! If you're upgrading from _Zeit_ v0 to v1 or later, please first
export your existing _Zeit_ v0 database via the following command:

```sh
zeit export --format zeit > ~/zeit_export.json
```

Only then upgrade _Zeit_ to the latest release and run the following command to
import the previously exported v0 database:

```sh
zeit import -f v0 ~/zeit_export.json
```

### Life hacks

#### `dmenu` compatible project/task selector

Requires `jq` to be installed and preferred `dmenu` launcher (e.g. `bemenu`,
`rofi`, etc.) to be set as `DMENU_PROGRAM`:

```sh
zeit projects -f json | jq -r '.[] | .sid as $parent_sid | .tasks? // [] | .[] | "\($parent_sid)/\(.sid)"' | $DMENU_PROGRAM
```

## Integrations

This is a list of integrations and extensions that work with _Zeit_:

| Integration | Description | Author |
| ----------- | ----------- | ------ |
| TODO        | TODO        | TODO   |

## Development

For development purposes `ZEIT_DATABASE` can be set to an empty string
(`export ZEIT_DATABASE=""`) to activate the non-persisting in-memory database.
In addition, _Zeit_ can be run with the global flag `--debug`
(`zeit --debug ...`) to enable additional command line output.

To run all available tests, `make test` can be utilized. To build _Zeit_
`make build` can be used.
