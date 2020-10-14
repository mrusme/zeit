zeit
----

Zeit erfassen. A command line tool for tracking time spent on tasks & projects.

## Build

```
make
```

## Usage

Please make sure to `export ZEIT_DB=~/.config/zeit.db` (or whatever location you would like to have the zeit database at).

### Start tracking

```
zeit track --help
```

Example:

```
zeit track --project project --task task --begin -0:15
```

### Finish tracking

```
zeit finish --help
```

Example:

```
zeit finish
```

