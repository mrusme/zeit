zeit
----

Zeit erfassen. A command line tool for tracking time spent on tasks & projects.

## Build

```
make
```

## Usage

Please make sure to `export ZEIT_DB=~/.config/zeit.db` (or whatever location you would like to have the zeit database at).

### Track activity

```
zeit track --help
```

Example:

```
zeit track --project project --task task --begin -0:15
```

### Show current activity

```
zeit tracking
```

### Finish tracking activity

```
zeit finish --help
```

Example:

```
zeit finish
```

### List tracked activity

```
zeit list
```
