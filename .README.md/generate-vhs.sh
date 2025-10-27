#/bin/sh

export ZEIT_DATABASE="/tmp"

/bin/ls -1 ./*.tape | while read tape; do vhs "$tape"; done
