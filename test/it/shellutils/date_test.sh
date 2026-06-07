TestDateEpochDate() {
    date -u -d @0 +%F
}

TestDateEpochTime() {
    date -u -d @0 +%T
}

TestDatePercent() {
    date +%%
}

TestDateYearDigits() {
    date +%Y | grep -E '^[0-9]{4}$' > /dev/null && echo "ok"
}
