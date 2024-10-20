echo "Building the binary locally..."
go build .
if [ $? -ne 0 ]; then
    echo "Build failed."
    exit 1
fi

SESSION_NAME="p2p"
BINARY_NAME="p2p"

tmux new-session -d -s "$SESSION_NAME" -n "$BINARY_NAME"

mkdir -p 1
mkdir -p 2
mkdir -p 3
mkdir -p 4

N=4
T=2
token="testkeygen"


tmux split-window -h
tmux split-window -v

tmux select-pane -t 0
tmux split-window -v

tmux select-pane -t 0
tmux send-keys "echo instance 1" C-m
tmux send-keys "./$BINARY_NAME keygen --vault 0  --n $N --t $T --token $token" C-m

tmux select-pane -t 1
tmux send-keys "echo instance 2" C-m
tmux send-keys "./$BINARY_NAME keygen --vault 1 --n $N --t $T --token $token" C-m

tmux select-pane -t 2
tmux send-keys "echo instance 3" C-m
tmux send-keys "./$BINARY_NAME keygen --vault 2 --n $N --t $T --token $token" C-m


tmux select-pane -t 3
tmux send-keys "echo instance 4" C-m
tmux send-keys "./$BINARY_NAME keygen --vault 3 --n $N --t $T --token $token" C-m


tmux select-pane -t 0
tmux attach-session -t "$SESSION_NAME"

tmux set-option -g remain-on-exit off
tmux setw -t "$SESSION_NAME" synchronize-panes on


echo "Session ended."