# sddf: Search & Delete Duplicated File
sddfコマンドは、データ内容が同じファイルを探し、重複ファイルを削除する。削除後は、重複ファイルのうち、最新ファイルのみが残る。なお、ファイルの同一性は、速度面を考慮してmd5チェックサムで確認している。

## 使い方
```Usage
$ sddf [OPTIONS] PATH
```

| short option | long option | description |
|:------|:-----|:------|
| -o    | --output    | 重複ファイル一覧の出力先ファイル名を拡張子無しで指定（デフォルト：duplicated-file.sddf）|
| -h   | --help    | ヘルプメッセージを表示する |
| -v | --version  | バージョンを表示する|

## 実行例
### 重複ファイルを探す
```
(注釈) ディレクトリ構成
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

(注釈)重複ファイル一覧の確認
$ cat duplicated-file.sddf 
[6121d44e98dc80f4f39937f13827547b]
def/ev088a.bmp
def/ghi/ev088a.bmp
ev088a.bmp
```

### 重複ファイルの削除
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
