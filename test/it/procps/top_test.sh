TestTopHeader() { top -bn1 | sed -n '1p' | grep -c '^top -'; }
TestTopTasks() { top -bn1 | grep -c '^Tasks:'; }
