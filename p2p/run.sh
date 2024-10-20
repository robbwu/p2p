echo "Building the binary locally..."
go build .
if [ $? -ne 0 ]; then
    echo "Build failed."
    exit 1
fi
echo "Building the binary for remote linux-amd64..."
GOOS=linux GOARCH=amd64 go build -o p2p-linux-amd64 .

REMOTE_DIR="/tmp"
BINARY_NAME="p2p"
SESSION_NAME="local_remote_session"



ssh ex44 killall p2p
ssh dell killall p2p

echo "Transferring the binary to the remote server..."
scp -C p2p-linux-amd64 ex44:$REMOTE_DIR/$BINARY_NAME
scp -C p2p-linux-amd64 dell:$REMOTE_DIR/$BINARY_NAME


tmux new-session -d -s "$SESSION_NAME" -n "$BINARY_NAME"

tmux split-window -h
tmux split-window -v

tmux select-pane -t 0
tmux send-keys "echo Running $BINARY_NAME locally..." C-m
tmux send-keys "./$BINARY_NAME " C-m

tmux select-pane -t 1
tmux send-keys "echo Running $BINARY_NAME on remote server..." C-m
tmux send-keys "ssh ex44 'cd $REMOTE_DIR && chmod +x $BINARY_NAME && ./$BINARY_NAME'" C-m

tmux select-pane -t 2
tmux send-keys "echo Running $BINARY_NAME on remote server..." C-m
tmux send-keys "ssh dell 'cd $REMOTE_DIR && chmod +x $BINARY_NAME && ./$BINARY_NAME'" C-m

tmux select-pane -t 0
tmux attach-session -t "$SESSION_NAME"

tmux set-option -g remain-on-exit off
tmux setw -t "$SESSION_NAME" synchronize-panes on


echo "Session ended."