# ghrdc - GitHub Release Downloads Counter
ghrdcコマンドは、GitHub APIを用いてリリースファイルのダウンロード数を表示する。デフォルトでは、最新リリースのダウンロード数を表示する。ghrdcコマンドはGitHub API認証しないため、制約がある。
- 一時間あたりに60回しか使用できない
- Organization repositoryの情報を取得できない
- Private repositoryの情報を取得できない  

## 使い方
``` Usage
$ ghrdc [OPTIONS] USER_NAME/RPOSITORY_NAME
```

| short option | long option | description |
|:------|:-----|:------|
| -a    | --all    | リリース毎のダウンロード数を表示する　|
| -t    | --total    |  全リリースのダウンロード数を表示する|
| -h   | --help    | ヘルプメッセージを表示する |
| -v | --version  | バージョンを表示する|

## 実行例
### 最新リリースのダウンロード数を表示
```
$ ghrdc  nao1215/mimixbox
[Name(Version)]             :Version 0.12.1
[Release Date]              :2021-11-20 04:00:23 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0
```

### 全リリースのダウンロード数を表示
```
$ ghrdc -t nao1215/mimixbox
[Name(Version)]                    :All release
[Release Date]                     :-
[Binary Download Count(total)]     :0
[Source Code Download Count(total)]:0
```

### リリース毎のダウンロード数を表示
```
$ ghrdc -a nao1215/mimixbox
[Name(Version)]             :Version 0.12.1
[Release Date]              :2021-11-20 04:00:23 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.9.1
[Release Date]              :2021-11-19 07:27:19 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.6.0
[Release Date]              :2021-11-18 13:54:57 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.3.0
[Release Date]              :2021-11-16 17:17:09 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :Version 0.1.1
[Release Date]              :2021-11-15 16:06:40 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0

[Name(Version)]             :MimixBox Version 0.0.1 (Buggy)
[Release Date]              :2021-11-14 13:02:27 +0000 UTC
[Binary Download Count]     :0
[Source Code Download Count]:0
```