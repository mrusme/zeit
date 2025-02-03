## Command structure

| Root | Sub        | short-opt | long-opt                    | Time Option | New |
| ---- | ---------- | --------- | --------------------------- | ----------- | --- |
| zeit |            | -h        | --help                      |             |     |
|      |            |           | --no-colors                 |             |     |
|      |            |           | --config                    |             | X   |
|      |            | -d        | --debug                     |             |     |
|      | completion | -h        | --help                      |             |     |
|      |            |           | --no-descriptions           |             |     |
|      | entry      | -b        | --begin                     | X           |     |
|      |            |           | --decimal                   |             |     |
|      |            | -s        | --finish                    | X           |     |
|      |            | -h        | --help                      |             |     |
|      |            | -n        | --notes                     |             |     |
|      |            | -p        | --project                   |             |     |
|      |            | -t        | --task                      |             |     |
|      | erase      | -h        | --help                      |             |     |
|      | export     | -h        | --help                      |             |     |
|      |            |           | --format                    |             |     |
|      |            | -p        | --project                   |             |     |
|      |            |           | --range                     |             | X   |
|      |            |           | --since                     | X           |     |
|      |            | -t        | --task                      |             |     |
|      |            |           | --until                     | X           |     |
|      | finish     | -b        | --begin                     | X           |     |
|      |            | -s        | --finish                    | X           |     |
|      |            | -h        | --help                      |             |     |
|      |            | -n        | --notes                     |             |     |
|      |            | -p        | --project                   |             |     |
|      |            | -t        | --task                      |             |     |
|      | help       | -h        | --help                      |             |     |
|      | import     |           | --format                    |             |     |
|      |            | -h        | --help                      |             |     |
|      | list       |           | --append-project-id-to-task |             |     |
|      |            |           | --decimal                   |             |     |
|      |            | -h        | --help                      |             |     |
|      |            |           | --only-projects-and-tasks   |             |     |
|      |            |           | --only-tasks                |             |     |
|      |            | -p        | --project                   |             |     |
|      |            |           | --range                     |             | X   |
|      |            |           | --since                     | X           |     |
|      |            | -t        | --task                      |             |     |
|      |            |           | --total                     |             |     |
|      |            |           | --until                     | X           |     |
|      | project    | -c        | --color                     |             |     |
|      |            | -h        | --help                      |             |     |
|      | report     | -h        | --help                      |             | X   |
|      |            | -p        | --project                   |             | X   |
|      |            |           | --range                     |             | X   |
|      |            |           | --since                     | X           | X   |
|      |            | -t        | --task                      |             | X   |
|      |            |           | --until                     | X           | X   |
|      | resume     | -h        | --help                      |             | X   |
|      |            | -b        | --begin                     | X           | X   |
|      |            | -s        | --finish                    | X           | X   |
|      | sketchy    | -h        | --help                      |             | X   |
|      | stats      |           | --decimal                   |             |     |
|      |            | -h        | --help                      |             |     |
|      | switch     | -h        | --help                      |             | X   |
|      |            | -b        | --begin                     | X           | X   |
|      |            | -n        | --notes                     |             | X   |
|      |            | -p        | --project                   |             | X   |
|      |            | -t        | --task                      |             | X   |
|      | switchback | -h        | --help                      |             | X   |
|      |            | -b        | --begin                     | X           | X   |
|      | task       | -g        | --git                       |             |     |
|      |            | -h        | --help                      |             |     |
|      | track      | -b        | --begin                     | X           |     |
|      |            | -s        | --finish                    | X           |     |
|      |            | -f        | --force                     |             |     |
|      |            | -n        | --notes                     |             |     |
|      |            | -p        | --project                   |             |     |
|      |            | -t        | --task                      |             |     |
|      |            | -h        | --help                      |             |     |
|      | tracking   | -h        | --help                      |             |     |
|      | version    | -h        | --help                      |             |     |

## Changes

### Extension Viper and Cobra

The extension of Cobra with Viper opens up the possibility of persisting time settings in a configuration file. This configuration file is optional, all default settings are chosen so that the behaviour of time does not change if this file does not exist.

By default, this file is searched for in $XDG_CONFIG_HOME/zeit/zeit.yaml (XDG_CONFIG_HOME => $HOME/.config), can also be overwritten with --config.

```
db: /Users/schretzi/OneDrive/Zeit/zeit.db
debug: false
firstWeekDayMonday: true
```

### Linting

I also use GO professionally and as a result I have linter and style checking software running on my system, which report masses of warnings and errors wih the current code. I have cleaned up the code and adapted it according to SolarLint and GO best practices so that my IDE and build tools are clear again and new, real errors are visible.

### Time Parsing

Different places use different parsing of time - track vs. entry. This always leads to errors during entry or some entries have to be unnecessarily long.
I have now combined these processes: I have incorporated dateparse from entryCmd into the helper/parseTime function and then adapted entryCmd so that this function is used via the struct method. This means that all -b and -s parameters now process the time entered in the same way. In my opinion, this should also solve issue #29 in Github.

#### now Library vs. DataParse

Elsewhere, the now library is used to parse time. I did not succeed in deduplicating the two libraries to one, as now does not successfully master most of the test cases. Therefore, I still use DataParse for the inputs and now for the list selection

#### Test Cases

```
go run . track -p "TESTS" -t "Zeit-Test" -b '10:00' -s '11:00'
go run . track -p "TESTS" -t "Zeit-Test" -b '01:00pm' -s '02:00pm'
go run . track -p "TESTS" -t "Zeit-Test" -b '-04:00' -s '-03:00'
go run . track -p "TESTS" -t "Zeit-Test" -b '-02:00' -s '+01:00'
go run . track -p "TESTS" -t "Zeit-Test" -b '2023-09-11 10:00 +0300' -s '2023-09-11 12:00 +0400'
go run . track -p "TESTS" -t "Zeit-Test" -b '2023-09-11 20:00' -s '2023-09-11 21:00'
go run . track -p "TESTS" -t "Zeit-Test" -b '2023-09-11T20:00' -s '2023-09-11T21:00'
go run . track -p "TESTS" -t "Zeit-Test" -b '2023-09-11T20:00:00' -s '2023-09-11T21:00:00'
go run . track -p "TESTS" -t "Zeit-Test" -b '2023-09-11T20:00:00+02:00' -s '2023-09-11T21:00:00+02:00'
```

### Relative Time

Relative time entries are always calculated by time.Now. In the context of tracking new entries this also makes sense, when editing existing entries I would expect the changes to be applied to the existing time, this has now been changed.

```
go run . track -p TESTS -t Rel-Tests -b -01:00 -s +01:00

go run . list
f4297201-c5d9-415b-a7f2-5f39f4fbf19b Rel-Tests on TESTS from 2024-05-18 20:52 +0200 to 2024-05-18 22:52 +0200 (2:00h)

go run . entry f4297201-c5d9-415b-a7f2-5f39f4fbf19b -b -01:00 -s +01:00

go run . list
f4297201-c5d9-415b-a7f2-5f39f4fbf19b Rel-Tests on TESTS from 2024-05-18 19:52 +0200 to 2024-05-18 23:52 +0200 (4:00h)
```

### Round to Minute

My applications do not require billing to the second, minutes are sufficient. At the moment, however, there may be deviations or ambiguities due to rounding. I have added an optional setting which always rounds to the full minute

```
time:
   no-seconds: true
```

```
{"begin":"2024-05-18T21:14:06.74637+02:00","finish":"2024-05-18T23:14:06.746393+02:00","project":"TESTS","task":"Rel-Tests","user":"schretzi"}
{"begin":"2024-05-18T21:14:00+02:00","finish":"2024-05-18T23:14:00+02:00","project":"TESTS","task":"Rel-Tests","user":"schretzi"}
```

### Project and Task mandatory with default Project

For my use case, entries without project and task make no sense, most of the time is booked to a project (job). I therefore have the following optional settings:

- Project and task are required, no entry can be created without them
- If no project is passed as a parameter, a default value can be used

```
project:
  mandatory: true
  default: TESTS
task:
  mandatory: true
```

### since/until/range

I always need the same relative time ranges for the list view or the report, setting --since and --until for this is time consuming, so I added the optional --range parameter.If this is set, --since and --until are set to the corresponding values via the now library:

- today
- yesterday
- thisWeek
- lastWeek
- thisMonth
- lastMonth

#### Testcases

```
MONTH=4
YEAR=2024
for i in $(seq 1 30)
do
	go run . track -p "TESTS" -t "RangeTests" -b "2024-${MONTH}-${i} 10:00" -s "2024-${MONTH}-${i} 17:00"
done

MONTH=5
YEAR=2024
for i in $(seq 1 19)
do
	go run . track -p "TESTS" -t "RangeTests" -b "2024-${MONTH}-${i} 10:00" -s "2024-${MONTH}-${i} 17:00"
done

go run . list --range today
go run . list --range thisWeek
```

### FmtDuration Bug - Open:

fmt.Println(trackDiff)

taskDuration := fmtDuration(trackDiff)

fmt.Println(taskDuration)

1h20m0s
1:19

### New Functions resume / switch / switchback

Some processes that occur in my everyday work have required several steps or entries that can be avoided, so there are three new functions:

- resume: The last task (last entry in the list sorted by start time) is resumed - only the times are provided as parameters, otherwise it would not be the last task
- switch: I always have to interrupt my work due to meetings or operational activities. The switch is used to end the current task at the specified time (-b or now()) and start a new one with the specified parameters
- switchback: After the meeting, I want to resume the previous activity using -b to set the time of the switch

Example procedure: I work on the development of the new functions until the end of the previous day, in the morning I resume work, at 09.00 there is the daily with my team colleagues, then the development is continued:

```
zeit track -p "TESTS" -t "Develop new features" -b "2024-05-18 15:00" -s "2024-05-18 19:00"
 ▶ tracked Develop new features on TESTS

zeit list
38cdcc16-eb58-4b5b-b564-dbfb979f537c Develop new features on TESTS from 2024-05-18 15:00 +0200 to 2024-05-18 19:00 +0200 (4:00h)

zeit resume -b "2024-05-19 07:40"
 ▶ began tracking Develop new features on TESTS

zeit list
38cdcc16-eb58-4b5b-b564-dbfb979f537c Develop new features on TESTS from 2024-05-18 15:00 +0200 to 2024-05-18 19:00 +0200 (4:00h)
7461cdf8-efc9-4468-9f7c-8990ab1a62df Develop new features on TESTS from 2024-05-19 07:40 +0200 to 2024-05-19 10:32 +0200 (2:52h) [running]


zeit switch -b 09:00 -t "Daily"
 ■ finished tracking Develop new features on TESTS for 1:20h
 ▶ began tracking Daily on TESTS

zeit switchback -b 09:20
 ■ finished tracking Daily on TESTS for 0:20h
 ▶ began tracking Develop new features on TESTS

zeit list
38cdcc16-eb58-4b5b-b564-dbfb979f537c Develop new features on TESTS from 2024-05-18 15:00 +0200 to 2024-05-18 19:00 +0200 (4:00h)
7461cdf8-efc9-4468-9f7c-8990ab1a62df Develop new features on TESTS from 2024-05-19 07:40 +0200 to 2024-05-19 09:00 +0200 (1:20h)
421f65b2-291a-47d0-a1b4-4e344ea52731 Daily on TESTS from 2024-05-19 09:00 +0200 to 2024-05-19 09:20 +0200 (0:20h)
81e2ca0f-6de7-4550-ab0b-fc9e3e576544 Develop new features on TESTS from 2024-05-19 09:20 +0200 to 2024-05-19 10:41 +0200 (1:21h) [running]

```

### New Function report

To be able to easily transfer my accumulated times to the time reports of my employer and my private projects, I need a clearer overview than list, but more detailed than stats. Therefore I have created a new function report that totals per day / project / task.

```
 zeit report -h
Reporting summaries on daily, project, task level for a given range

Usage:
  zeit report [flags]

Flags:
  -h, --help             help for report
  -p, --project string   Project to be listed
      --range string     shortcut to set since/until for a given range (today, yesterday, thisWeek, lastWeek, thisMonth, lastMonth)
      --since string     Date/time to start the list from
  -t, --task string      Task to be listed
      --until string     Date/time to list until

Global Flags:
      --config string   config file (default is $XDG_CONFIG_HOME/zeit/zeit.yaml)
  -d, --debug           Display debugging output in the console. (default: false)
      --no-colors       Do not use colors in output
```

For a quick overview, there is the option in the configuration file to define a period as the default, in my case the current week

```
report:
  default: thisWeek
```

```
 zeit report
Reporting for Timerange: thisWeek / 2024-05-27 - 2024-06-02

 2024-05-27 :  3h0m0s
     TESTS :  2h0m0s
         Testing actual status :  2h0m0s
     ZEIT :  1h0m0s
         Creating Switch function :  1h0m0s

 2024-05-28 :  1h23m0s
     ZEIT :  1h23m0s
         Creating Report function :  40m0s
         Daily :  30m0s
         Documentation :  13m0s

```

### New Function sketchy

I work on a Macbook with the Sketchybar as an additional menu bar. In this I would like to display the current status of the time recording. The output of tracking cannot be used 1:1 and I am still considering adding the total hours per day or something similar in the future. Therefore a new function ‘sketchy’ which has no parameters and calls the sketchbar function to set the label. The path to sketchybar must be set in the configuration file:

```
sketchybar:
  path: /opt/homebrew/bin/sketchybar
```

The following paragraph must be added to the sketchbarrc:

```
##### Show actual Zeit tracking

sketchybar --add item zeit e \
           --set zeit icon=󱏁  \
                 script="<PATH TO Zeit>/zeit sketchy" \
                 update_freq=15
```

This means that the current tracking in the form \<Project>|\<Task> is displayed in the sketch bar to the right of the notch (e): \<Duration>

### Custom Completions for Tasks

After entering ‘-task’, \<TAB>\<TAB> displays the list of tasks in the database.
Future topics on this:

- Filter by tasks from projects
- Performance when there are large numbers of entries in the database (it is unclear to me how the performance develops anyway)

## Future Ideas if I find time

- Listing Project (with tasks)
- Listing Tasks
- Archiving Tasks (in sense that autocompletion only shows active tasks)
- UI (report, edit existing tasks)
