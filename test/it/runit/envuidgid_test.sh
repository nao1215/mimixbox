# envuidgid sets $UID/$GID from /etc/passwd; root is always present (0:0).
TestEnvuidgidRoot() { envuidgid root sh -c 'echo "$UID:$GID"'; }
TestEnvuidgidUnknown() { envuidgid nonexistent_xyz true 2>/dev/null; echo "rc=$?"; }
