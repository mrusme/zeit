#!/usr/bin/env bash

## Skeytchybar configuration:
# sketchybar --add item zeit e \
#            --set zeit icon=󱏁  \
#                  script="$HOME/bin/zeit-sketchybar.sh" \
#                  update_freq=15


ZEIT_BIN=$HOME/bin/zeit
SKETCHY_BIN=/opt/homebrew/bin/sketchybar

line_identifier='^ ▶ tracking'

tracking=$($ZEIT_BIN tracking --no-colors | grep "$line_identifier" | sed -e "s/$line_identifier//")

echo $tracking
$SKETCHY_BIN --set zeit label="$tracking"
