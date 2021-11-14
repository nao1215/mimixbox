% SERIAL(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% 2021年9月

# 名前

serial –  シリアル番号付きのファイル名にリネームする。

# 書式

**serial** [OPTIONS] DIRECTORY_PATH

# 説明
**serial**は、任意のディレクトリ以下にあるファイルの名前を  
ユーザ指定の名前に連番を付与してリネームするCLIコマンドです。  
serialは、リネームしたファイルの格納先ディレクトリを指定でき  
ます。また、オリジナルファイルを保持したい場合、リネームでは  
なくファイルコピーができます。

# 例
**カレントディレクトリにあるファイルの名前をシリアル番号付きのファイル名にリネームする。**  

    $ ls  
      a.txt  b.txt  c.txt  
    $ serial --name demo  .  
      Rename a.txt to 1_demo.txt  
      Rename b.txt to 2_demo.txt  
      Rename c.txt to 3_demo.txt

**指定のディレクトリへファイルをコピー&リネーム**  

    $ serial -s -k -n ../../dir/demo .  
      Copy a.txt to ../../dir/demo_0.txt  
      Copy b.txt to ../../dir/demo_1.txt  
      Copy c.txt to ../../dir/demo_2.txt

# オプション
**-d**, **--dry-run**
:   標準出力にファイル名のリネーム結果を表示します（ファイル更新はしません）。

**-f**, **--force**
:   同名のファイルが存在する場合であっても、強制的に上書き保存します。

**-h**, **--help**
:   ヘルプメッセージを表示します。

**-k**, **--keep**
:   リネーム前のファイルを保持します（リネームはせず、コピーします）。

**-n new_filename**, **--name=new_filename**
:   格納先のディレクトリ名を含んだ／含まないベースファイル名（このファイル名に連番を付与します）。

**-p**, **--prefix**
:   連番をファイル名の先頭に付与します（デフォルト）。

**-s**, **--suffix**
:   連番をファイル名の末尾に付与します。

**-v**, **--version**
:   serialコマンドのバージョンを表示します。

# 終了ステータス
**0**
:   成功

**1**
:   serialコマンドの引数指定でエラー

# バグ
GitHub Issuesを参照してください。URL：https://github.com/nao1215/mimixbox/issues

# ライセンス
MimixBoxプロジェクトは、MIT License条文およびApache License 2.0条文の下でライセンスされています。