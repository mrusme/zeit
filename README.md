zeit
----

```
                          ███████╗███████╗██╗████████╗                             
                          ╚══███╔╝██╔════╝██║╚══██╔══╝
                            ███╔╝ █████╗  ██║   ██║   
                           ███╔╝  ██╔══╝  ██║   ██║   
                          ███████╗███████╗██║   ██║   
                          ╚══════╝╚══════╝╚═╝   ╚═╝   
```

Zeit erfassen. A command line tool for tracking time spent on tasks & projects.

## Build

```sh
make
```

## Usage

Please make sure to `export ZEIT_DB=~/.config/zeit.db` (or whatever location 
you would like to have the zeit database at).

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

#### Tyme 3 JSON

It's possible to import JSON exports from [Tyme 3](https://www.tyme-app.com). 
It is important that the JSON is exported with the following options set/unset:

![Tyme 3 JSON export](documentation/tyme3json.png)

- `Start`/`End` can be set as required
- `Format` has to be `JSON`
- `Export only unbilled entries` can be set as required
- `Mark exported entries as billed` can be set as required
- `Include non-billable tasks` can be set as required
- `Filter Projects & Tasks` can be set as required
- `Combine times by day & task` **must** be unchecked

During import, *zeit* will create SHA1 sums for every Tyme 3 entry, which 
allows it to identify every imported activity. This way *zeit* won't import the 
exact same entry twice. Keep this in mind if you change entries in Tyme and 
then import them again into *zeit*.

Example:

```sh
zeit import --tyme ./tyme.export.json
```
