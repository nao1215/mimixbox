# fakemovie: 画像に動画開始ボタンを付与
fakemovieは、画像に動画開始ボタンを付与する。オリジナル版は、[mattn氏が作成したコマンド](https://github.com/mattn/fakemovie)であり、MimixBox版はそこから一部改変を加えている。

## 使い方
``` Usage
fakemovie [OPTIONS] IMAGE_FILE_NAME
```

| short option | long option | description |
|:------|:-----|:------|
|-o| --output| 出力先ファイル名を指定する<br>（デフォルト：オリジナル名にサフィックス"_fake"を付与する） |
|-p|--phub| ボタンをp-hub風とする<br>（デフォルトはTwitter風）|
|-r|--radius| ボタンの半径を設定する<br>（デフォルトは自動計算）|
| -h   | --help    | ヘルプメッセージを表示する　 |
| -v | --version  | バージョンを表示する|


## 実行例
### Twitter風ボタン
```
$ fakemovie lena.png 
```

![オリジナル](../../../images/lena.png "オリジナル")
![Twitter風](../../../images/lena_twitter_fake.png "Twitter風")

### P-hub風ボタン
```
$ fakemovie -p lena.png -o lena_phub_fake.png
```
![オリジナル](../../../images/lena.png "オリジナル")
![Phub風](../../../images/lena_phub_fake.png "Phub風")