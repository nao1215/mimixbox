TestUsersExit() {
    users >/dev/null 2>&1
    echo $?
}

TestUsersMissing() {
    out=$(users /no/such/mimixbox/utmp)
    echo "[$out] rc=$?"
}
