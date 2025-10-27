#/bin/sh

export tmpdb="$(mktemp -d)"
export ZEIT_DATABASE="$tmpdb"

/bin/ls -1 ./*.tape | while read tape; do vhs "$tape"; done
