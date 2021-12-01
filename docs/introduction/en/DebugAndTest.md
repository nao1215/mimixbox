# Purpose of this document
This document describes how to safely debug MimixBox and add tests. As a prerequisite, you had better not fully install MimixBox in your development environment until Version 1.0.0. As of November 2021, MimixBox has not been fully tested.  

If MimixBox is used with Coreutils or BusyBox on the system, the system does not work properly.

# Debugging environment (Dog fooding environment)
MimixBox is safe to execute commands within the Docker environment. In [Current Docker Preferences] (../../../Dockerfile), all MimixBox built-in commands are symbolically linked under /usr/local/bin. If you add any command, follow the procedure below to check the operation. However, there is a problem now and I haven't built MimixBox in Docker using locally modified code [(Issue # 4)] (https://github.com/nao1215/mimixbox/issues/4)
```
$ make docker
$ (here, in Docker) 
```
Bash is used in Docker because the MimixBox shell is incomplete. You can also check which command has been replaced by MimixBox with the following command.
```
$ mimixbox --list
```

# Unit Test
Unit tests are created according to the manners of the golang language. In other words, "target.go" and "target_test.go" exist in the same directory as shown in the directory structure below.
```
   └── lib
       ├── file.go
       ├── file_test.go
```
The MimixBox project does not aim for 100% unit test coverage. However, since the MimixBox library (mimixbox/internal/lib) has a lot of general-purpose code, I want to have 100% coverage as much as possible.

Unit tests can be easily run with the make command.
```
$ make ut
```
In the MimixBox project, if "$ git push" is detected, GitHub Action is running the unit test. Therefore, we recommend that you run unit tests before pushing your code.

# Integration Test
In the integration test, check the behavior of each command in [ShellSpec] (https://github.com/shellspec/shellspec). The test code is stored under the test directory.
```
test/
├── it
│   └── shellutils
│       └── echoTest.sh  ※ Definition of shell script function for unit test 
├── spec
│   ├── echo_spec.sh     ※ Unit test expectations
│   └── spec_helper.sh
└── ut
```
Like the unit test, the integration test can be easily executed with the make command (GitHub Action is also used).
```
$ make build
$ sudo make full-install   ※ Create symbolic link for mimixbox builtin commands for a limited time only.
$ make it                  ※ Execute test
$ sudo make remove         ※ Delete symblic link
```
# When installing MimixBox in the HOST environment
## Proper use of installation options
MimixBox has two options for creating symbolic links to built-in commands.   

The first is the --install option. MimixBox does not create a symbolic link if a command with the same name exists on the system. --install was introduced as a safe install based on the experience of breaking the system in the past.
```
$ sudo mimixbox --install /usr/local/bin
```
The second is the --full-install option. Create symbolic links for all commands, regardless of system state. This option is deprecated at this stage.
```
$ sudo mimixbox --full-install /usr/local/bin
```
## If the system breaks (e.g. GUI does not start)
You need to remove the MimixBox symlinks from your system. The specific procedure is as follows.  

1. PC power off
2. Start in rescue mode
3. Execute "$ sudo mimixbox --remove $(directory where symbolic links exist)" 
   e.g. sudo mimixbox --remove /usr/local/bin
4. Reboot
```
$ sudo ./mimixbox --remove /usr/local/bin/
Delete symbolic link: /usr/local/bin/fakemovie
Delete symbolic link: /usr/local/bin/mbsh
Delete symbolic link: /usr/local/bin/path
Delete symbolic link: /usr/local/bin/serial
Delete symbolic link: /usr/local/bin/sh
Delete symbolic link: /usr/local/bin/true
Delete symbolic link: /usr/local/bin/which
Delete symbolic link: /usr/local/bin/cat
Delete symbolic link: /usr/local/bin/echo
Delete symbolic link: /usr/local/bin/false
Delete symbolic link: /usr/local/bin/ghrdc
Delete symbolic link: /usr/local/bin/mkdir
```
If possible, please report the bug to [MimixBox Issues] (https://github.com/nao1215/mimixbox/issues).

# Logging in MimixBox (under consideration)
Since the operation of MimixBox is unstable, we are considering the logging function.  

If MimixBox operates stably in the future, it is possible that multiple users will execute MimixBox built-in commands at the same time. In that case, multiple processes (MimixBox) write to one log file at the same time. Due to the simultaneous writing, there is no guarantee that the logs will be written to the log file as expected.  

Therefore, each process is considering the specification to write a log to the named pipe. Writes to named pipes are atomic if they are less than or equal to PIPE_BUF size and log messages do not conflict (do not mix).  

The log size of MimixBox is assumed to be smaller than PIPE_BUF (minimum 512Byte), and it is guaranteed to be highly atomic. A dedicated logging daemon is responsible for reading logs and writing to log files.
![MimixBox logging flow](/docs/images/debug_logging.jpg "MimixBox logging flow")

It is a slightly exaggerated design, and there is a problem in recovery when the logging daemon suddenly dies. Therefore, we are also considering how to create a log file for each user.  
