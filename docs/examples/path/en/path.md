# path: manipulating filename path
path extracts the directory name, file name, and extension from the path. It also create canonical path and absolute path.

## How to Use
``` Usage
$ path [OPTIONS] PATH
```

| short option | long option | description |
|:------|:-----|:------|
| -a    | --absolute    | Print absolute path.　|
| -b   | --basename    | Print basename (filename).  |
| -c   | --canonical    | Print canonical path (default). |
| -d   | --dirname    |  Print path without filename.   |
| -e   | --extension    |  Print file extention.  |
| -h   | --help    | Show help message.　 |
| -v | --version  | Show version.|

## Examples
### Get absolute path
```
$ path -a path  
/home/nao/.go/src/otter/path
```

### Get basename
```
$ path -b /etc/systemd/pstore.conf
pstore.conf
```

### Get canonical path
```
$ path -c cmd/../scripts/installer.sh   
scripts/installer.sh
```

### Get dirctory name
```
$ path -d /etc/ssh/ssh_config  
/etc/ssh
```

### Get file extension
```
$ path -e go.mod  
.mod
```
