TestSttyNotTty() {
    echo x | stty 2>&1
    echo "exit=$?"
}
