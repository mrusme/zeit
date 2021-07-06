#!/bin/sh
#
# Example waybar configuration:
#
# "custom/zeit": {
#   "format": "{}",
#   "exec": "zeit-waybar-wofi.sh",
#   "on-click": "zeit-waybar-wofi.sh click",
#   "interval": 10
# },
#

ZEIT_BIN=zeit

tracking=$($ZEIT_BIN tracking --no-colors)

if [[ "$1" == "click" ]]
then
  if echo "$tracking" | grep -q '^ ▶ tracking'
  then
    $ZEIT_BIN finish
    exit 0
  fi

  selection=$($ZEIT_BIN list \
    --only-projects-and-tasks \
    --append-project-id-to-task \
    | wofi \
      --dmenu \
      --sort-order default \
      --cache-file /dev/null\
  )

  task=$(echo $selection | pcregrep -io1 '└── (.+) \[.+')
  project=$(echo $selection | pcregrep -io1 '.+\[(.+)\]')

  if [[ "$task" == "" ]] || [[ "$project" == "" ]]
  then
    exit 1
  fi

  $ZEIT_BIN track -p "$project" -t "$task"
  exit 0
fi

echo -n $tracking

