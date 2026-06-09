TestTasksetPrint() {
    taskset -p $$ | grep -c 'affinity mask'
}

TestTasksetRun() {
    taskset -c 0 echo affined
}

TestTasksetInvalid() {
    taskset zzz echo x 2>/dev/null
    echo "rc=$?"
}
