TestUptime() { uptime | grep -c 'load average'; }
TestPwdx() {
    cd /tmp
    pwdx $$ | grep -c '/'
}
