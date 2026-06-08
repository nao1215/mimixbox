TestKillall() {
    sleep 30 &
    sleep 0.2
    killall sleep
    echo "killed:$?"
}
