## Zeit

[![SEGV 
LICENSE](https://img.shields.io/static/v1?label=SEGV%20LICENSE&message=1.0&labelColor=0060A8&color=ffffff)](https://xn--gckvb8fzb.com/segv/)

![zeit](.README.md/zeit.png)

[<img src="https://xn--gckvb8fzb.com/images/chatroom.png" width="275">](https://xn--gckvb8fzb.com/contact/)

Zeit, erfassen. A command line tool for tracking time spent on tasks & projects.

[Get some more info on why I build this
here](https://マリウス.com/zeit-erfassen-a-cli-activity-time-tracker/).

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

_zeit_'s data structure contains of the following key entities:

- `block`: A tracked time period that has a start and an end date/time
- `project`: A top-level project
- `task`: A task underneath a project or another task

A `block` consists of a `project` and a `task`. These don't have to pre-exist
and can be created on-the-fly when starting to track a new `block`. They can be
configured using the `zeit project` and the `zeit task` commands.

## Use

_zeit_ will store all its data inside its own database, which is located at
`$XDG_DATA_HOME/zeit/db/`. You can adjust this location by exporting
`ZEIT_DATABASE=~/your/preferred/path`.

### `start`/`switch`/`resume`

```sh
zeit start -p MyProject -t MyTask 5 minutes ago
```

```sh
zeit start work on MyProject/MyTask in 5 minutes
```

```sh
zeit switch to MyOtherProject/MyTask 5 minutes ago
```

```sh
zeit switch -p MyOtherProject -t MyTask 5 minutes ago
```

When you don't specify the project/task at start, as soon as you stop _zeit_
will ask you for the project and task. If you don't specify it at that point,
_zeit_ will not assign any project/task to the block.

### `end`/`pause`

```sh
zeit end 20 minutes ago
```

### Auto-Completion

_zeit_ can generate auto-completion scripts for your shell of choice. You can
load completions into your current session via:

```sh
source <(zeit completion bash)
```

(replace `bash` with your shell, e.g. `zsh`, `fish`, `powershell`)

To load completions for every new session, add them to your completions
directory, e.g.:

```
sudo zeit completion bash > /etc/bash_completion.d/zeit
```

## Integrations

This is a list of integrations and extensions that work with _zeit_:

| Integration | Description | Author |
| ----------- | ----------- | ------ |
| TODO        | TODO        | TODO   |
