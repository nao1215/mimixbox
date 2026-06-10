TestSetprivDump() { setpriv --dump | grep -cE 'uid:|no_new_privs:'; }
TestSetprivRun() { setpriv --no-new-privs -- echo ran; }
