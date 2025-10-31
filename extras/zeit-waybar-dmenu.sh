#!/bin/sh
#
# Example waybar configuration:
#
# "custom/zeit": {
#   "format": "{}",
#   "exec": "zeit-waybar-bemenu.sh",
#   "on-click": "zeit-waybar-bemenu.sh click",
#   "interval": 10
# },
#

ZEIT_BIN=zeit

as_hms() {
  local total_seconds=$1

  hours=$((total_seconds / 3600))

  minutes=$(((total_seconds % 3600) / 60))

  seconds=$((total_seconds % 60))

  printf "%02d:%02d:%02d" "$hours" "$minutes" "$seconds"
}

statusOut=$($ZEIT_BIN --format json)
for key in $(echo "$statusOut" | jq -r 'keys[]'); do
  value=$(echo "$statusOut" | jq -r ".${key}")
  export "$key"="$value"
done

if [[ "$1" == "click" ]]; then
  if [[ "$is_running" == "true" ]]; then
    $ZEIT_BIN end
    exit 0
  fi

  selection=$(zeit projects -f json |
    jq -r '.[] | .sid as $parent_sid | .tasks? // [] | .[] | "\($parent_sid)/\(.sid)"' |
    $DMENU_PROGRAM)

  task=$(printf "%s" "$selection" | cut -d '/' -f1)
  project=$(printf "%s" "$selection" | cut -d '/' -f2)

  if [[ "$task" == "" ]] || [[ "$project" == "" ]]; then
    exit 1
  fi

  $ZEIT_BIN start -p "$project" -t "$task"
  exit 0
fi

if [[ "$is_running" == "true" ]]; then
  timer_fmt=$(as_hms "$timer")
  printf "{\"text\": \"%s<span color='#ffffff'>/</span>%s <span color='#ffffff'>%s</span>\", \"class\": \"custom-zeit\", \"alt\": \"%s\" }\n" "$project_sid" "$task_sid" "$timer_fmt" "$status"
else
  printf "{\"text\": \"zeit %s\", \"class\": \"custom-zeit\", \"alt\": \"%s\" }\n" "$status" "$status"
fi
