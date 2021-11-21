![Serial logo](../../../images/serial.jpg "Serial logo")
# serial: Add serial number to file name
serial is a CLI command that renames files under any directory to the format "specified file name + serial number".
The original version was released as a standalone command, but now it is one of the MimixBox commands.

## How to use
```Usage
$ serial [OPTIONS] DIRECTORY_PATH
```

| short option | long option | description |
|:------|:-----|:------|
| -d    | --dry-run    | Output the file renaming result to standard output (do not update the file)　   |
| -f   | --force    | Forcibly overwrite and save even if a file with the same name exists　   |
| -h   | --help    | Show help message　   |
| -k   | --keep    | Keep the file before renaming　   |
| -n | --name   | Base file name with/without directory path (assign a serial number to this file name)   |
| -p | --prefix   | Add a serial number to the beginning of the file name  |
| -s | --suffix  | Add a serial number to the end of the file name(default) |
| -v | --version  | Show serial command version |

## Examples
### Default (Add serial number at prefix)
```
$ ls
a.txt  b.txt  c.txt  d.txt  e.txt

$ serial .
Rename a.txt to 0_a.txt
Rename b.txt to 1_b.txt
Rename c.txt to 2_c.txt
Rename d.txt to 3_d.txt
Rename e.txt to 4_e.txt

$ ls
0_a.txt  1_b.txt  2_c.txt  3_d.txt  4_e.txt
```

### Rename filename and add serial number at suffix
```
$ ls
a.txt  b.txt  c.txt  d.txt  e.txt

$ serial --suffix --name=demo .
Rename a.txt to demo_0.txt
Rename b.txt to demo_1.txt
Rename c.txt to demo_2.txt
Rename d.txt to demo_3.txt
Rename e.txt to demo_4.txt

$ ls
demo_0.txt  demo_1.txt  demo_2.txt  demo_3.txt  demo_4.txt
```

### Keep original files
```
$ ls
a.txt  b.txt  c.txt  d.txt  e.txt

$ serial --keep .
Copy a.txt to 0_a.txt
Copy b.txt to 1_b.txt
Copy c.txt to 2_c.txt
Copy d.txt to 3_d.txt
Copy e.txt to 4_e.txt

$ ls
0_a.txt  1_b.txt  2_c.txt  3_d.txt  4_e.txt  a.txt  b.txt  c.txt  d.txt  e.txt
```


##  Credits
The serial command project logo was created by the [DesignEvo (online logo maker)](https://www.designevo.com/).