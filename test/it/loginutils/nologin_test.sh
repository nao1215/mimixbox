TestNologinRefuses() { nologin; echo "rc=$?"; }
TestNologinIgnoresArgs() {
    count=$(nologin -c "echo pwned" 2>/dev/null | grep -c pwned)
    echo "$count"
}
