# sddf: Search & Delete Duplicated File
sddf command looks for files with the same data content and removes duplicate files. After deletion, only the latest file remains among the duplicate files. The identity of the files is confirmed by the md5 checksum in consideration of speed.

## How to use
```Usage
$ sddf [OPTIONS] PATH
```

| short option | long option | description |
|:------|:-----|:------|
| -o    | --output    | Change output file-name without extension (default: duplicated-file.sddf)|
| -h   | --help    | Show help message |
| -v | --version  | Show version|

## Examples
### Search duplicated file
```
(comment) directory structure
$ tree .
.
├── abc
│     └── ev095b.bmp
├── def
│     ├── ev088a.bmp
│     ├── ev092g.bmp
│     └── ghi
│            └── ev088a.bmp
└── ev088a.bmp


$ sddf .
Get all file path at .
.
Find the same file on a file content basis
5 / 5 [----------------------------------------------------------------] 100.00%
Write down duplicated file list to duplicated-file.sddf
1 / 1 [----------------------------------------------------------------] 100.00%
See duplicated file list: duplicated-file.sddf
If you delete files, execute the following command.
$ sddf duplicated-file.sddf

(comment) check duplicated file list.
$ cat duplicated-file.sddf 
[6121d44e98dc80f4f39937f13827547b]
def/ev088a.bmp
def/ghi/ev088a.bmp
ev088a.bmp
```

### Delete duplicated file
```
$ sddf duplicated-file.sddf 
Restore data from duplicated-file.sddf
5 / 5 [----------------------------------------------------------------] 100.00%
Decide delete target files
Start deleting files
Delete(Success): ev088a.bmp: 3145782Byte
Delete(Success): def/ghi/ev088a.bmp: 3145782Byte
End deleting files. Size=6291564Byte

$ tree
.
├── abc
│     └── ev095b.bmp
├── def
│     ├── ev088a.bmp
│     ├── ev092g.bmp
│     └── ghi
└── duplicated-file.sddf
```
