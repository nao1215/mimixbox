TestTimeoutFinishes() {
    timeout 5 echo done
}
TestTimeoutExpires() {
    timeout 0.1 sleep 5
    echo "exit:$?"
}
