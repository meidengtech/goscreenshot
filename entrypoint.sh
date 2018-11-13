#!/bin/bash

/headless-shell/headless-shell --disable-gpu --headless --no-sandbox --remote-debugging-address=0.0.0.0 --remote-debugging-port=9222 &
sleep 3

CMD="$@"

echo $CMD
echo $@

if [[ "$CMD" == "" ]]; then
    CMD=/usr/bin/html2image
fi

exec $CMD
