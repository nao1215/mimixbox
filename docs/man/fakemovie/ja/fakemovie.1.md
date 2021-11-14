% FAKEMOVIE(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% 2021年9月

# 名前

fakemovie –  画像に動画再生ボタンを付与します。

# 書式

**fakemovie** [OPTIONS] IMAGE_FILE_NAME

# 説明
**fakemovie**は、画像に動画再生ボタンを付与します。  
対応している画像フォーマットは、pngかjpgです。それ以外は実行時にエラーとなります。

# 例
**デフォルトの出力ファイル名を使用する場合**  
    $ fakemovie image.jpg  
    $ ls  
      image.jpg image_fake.jpg  

**出力ファイル名、ボタン色、ボタンサイズを指定する場合**  
    $ fakemovie -p -o output.jpg -r 50 output.jpg  
    $ ls  
      image.jpg output.jpg  


# オプション
**-o**, **--output**
:   アウトプットファイル名を指定します。デフォルトは、オリジナル名に"_fake"を付与します。

**-p**, **--phub**
:   ボタンの色をp-hubライクにします。デフォルトはTwitterに似た色です。

**-r**, **--radius**
:   ボタンの半径（整数）を指定します。デフォルトは画像サイズを元にした自動計算値です。

**-h**, **--help**
:   ヘルプメッセージを表示します。

**-v**, **--version**
:   fakemovieコマンドのバージョンを表示します。

# 終了ステータス
**0**
:   成功

**1**
:   fakemovieコマンドの引数指定でエラー、もしくは画像処理の実行時エラー

# バグ
GitHub Issuesを参照してください。URL：https://github.com/nao1215/mimixbox/issues

# ライセンス
MimixBoxプロジェクトは、MIT License条文およびApache License 2.0条文の下でライセンスされています。