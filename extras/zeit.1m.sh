#!/bin/bash

# <xbar.title>zeit</xbar.title>
# <xbar.author>Marius</xbar.author>
# <xbar.author.github>mrusme</xbar.author.github>
# <xbar.desc>Control `zeit` (https://github.com/mrusme/zeit) from the macOS menu bar.</xbar.desc>
# <xbar.image>https://github.com/mrusme/zeit/raw/main/documentation/zeit.png</xbar.image>
# <xbar.abouturl>https://マリウス.com/zeit-erfassen-a-cli-activity-time-tracker/</xbar.abouturl>
# <xbar.version>1.0</xbar.version>
#
# <xbar.var>string(ZEIT_BIN="/usr/local/bin/zeit"): Your zeit binary location</xbar.var>
# <xbar.var>string(ZEIT_DB="$HOME/.zeit.db"): Your zeit database location</xbar.var>
#
# Control `zeit` (https://github.com/mrusme/zeit) from the macOS menu bar.
#
# by Marius (marius@xn--gckvb8fzb.com)
#

echo $($ZEIT_BIN tracking)
exit
