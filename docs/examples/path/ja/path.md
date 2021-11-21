# path: ファイルパス操作
pathコマンドは、ディレクトリ名、ファイル名、拡張子をPATHから抽出する。Canonical PATHや絶対PATHも生成できる。

## 使い方
``` Usage
$ path [OPTIONS] PATH
```

| short option | long option | description |
|:------|:-----|:------|
| -a    | --absolute    | 絶対PATHを表示する|
| -b   | --basename    | ベースネームを表示する（デフォルト）  |
| -c   | --canonical    | Canonical PATHを表示する（デフォルト）|
| -d   | --dirname    |  ファイル名を取り除いたPATHを表示する |
| -e   | --extension    |  ファイル拡張子を表示する |
| -h   | --help    | ヘルプメッセージを表示する　 |
| -v | --version  | バージョンを表示する |

## 実行例
### 絶対PATHの取得
```
$ path -a path  
/home/nao/.go/src/otter/path
```

### ベースネームの取得
```
$ path -b /etc/systemd/pstore.conf
pstore.conf
```

### Canonical PATHの取得
```
$ path -c cmd/../scripts/installer.sh   
scripts/installer.sh
```

### ディレクトリ名の取得
```
$ path -d /etc/ssh/ssh_config  
/etc/ssh
```

### ファイル拡張子の取得
```
$ path -e go.mod  
.mod
```
