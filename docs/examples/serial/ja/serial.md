![Serial logo](../../../images/serial.jpg "Serial logo")
# serial: シリアル番号をファイル名に付与
serialは、指定されたディレクト以下にあるファイルに対してシリアル番号を付与するCLIコマンドである。オリジナル版は、単独のコマンドとしてリリースしましたが、現在はMimixBoxのコマンドの一部としている。

## 使い方
```Usage
$ serial [OPTIONS] DIRECTORY_PATH
```

| short option | long option | description |
|:------|:-----|:------|
| -d    | --dry-run    | ファイルを更新せずに、更新後のファイル名を出力する　|
| -f   | --force    | 同名ファイルが存在した場合、強制的に上書きする　   |
| -h   | --help    | ヘルプメッセージを表示する　 |
| -k   | --keep    | リネーム前ファイルをバックアップとして残す |
| -n | --name   | リネーム後ファイル名のベースネーム（拡張子は含まない）   |
| -p | --prefix   | シリアル番号をファイル名の先頭に付与する  |
| -s | --suffix  |  シリアル番号をファイル名の末尾（拡張子の前）に付与する（デフォルトの動作） |
| -v | --version  | バージョンを表示する|

## 実行例
### デフォルト (プレフィックスにシリアル番号をつける場合)
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

### サフィックスにシリアル番号を付与する場合
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

### オリジナルファイルを残す場合
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

## クレジット
シリアルコマンドロゴは、[DesignEvo (online logo maker)](https://www.designevo.com/)によって作成されている。
