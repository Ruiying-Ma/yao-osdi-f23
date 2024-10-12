tmux new -s "$1" -d
tmux send-keys -t 0 "${*:2}" Enter
sleep 3