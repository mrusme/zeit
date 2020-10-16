zeit
----

Zeit erfassen. A command line tool for tracking time spent on tasks & projects.

## Build

```sh
make
```

## Usage

Please make sure to `export ZEIT_DB=~/.config/zeit.db` (or whatever location you would like to have the zeit database at).

### Track activity

```sh
zeit track --help
```

Example:

```sh
zeit track --project project --task task --begin -0:15
```

### Show current activity

```sh
zeit tracking
```

### Finish tracking activity

```sh
zeit finish --help
```

Example:

```sh
zeit finish
```

### List tracked activity

```sh
zeit list
```

### Erase tracked activity

```sh
zeit erase --help
```

Example

```sh
zeit erase 14037730-5c2d-44ff-b70e-81f1dcd4eb5f
```

### Import tracked activities

```sh
zeit import --help
```

Example:

```sh
zeit import --tyme ./tyme.export.json
```
