TestIpcsAll() { ipcs | grep -c 'Message Queues'; }
TestIpcsShm() { ipcs -m; }
