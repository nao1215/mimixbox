TestPidofInit() {
    pidof mimixbox >/dev/null 2>&1
    # pidof of a definitely-running process: the test shell's own 'sleep'
    sleep 5 &
    SLEEP_PID=$!
    RESULT=$(pidof sleep)
    kill ${SLEEP_PID} 2>/dev/null
    echo "${RESULT}" | grep -q "${SLEEP_PID}" && echo found
}
