% PATH(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% 2021年9月

# 名前

path –  ファイル名パスを操作するためのコマンド

# 書式

**path** [OPTIONS] DIRECTORY_PATH

# 説明
**path**は、パスからディレクトリ名、ファイル名、拡張子を抽出します。  
また、正規化されたパスや全体化されたパスを生成します。

# 例
**絶対パスの取得**  
    $ path -a path    
      /home/nao/.go/src/otter/path

**ベースネームの取得**  
    $ path -b /etc/systemd/pstore.conf   
      pstore.conf

**正規化されたパスの取得**  
    $ path -c cmd/../scripts/installer.sh   
      scripts/installer.sh

**ディレクトリパスの取得**  
    $ path -d /etc/ssh/ssh_config  
      /etc/ssh

**ファイルの拡張子の取得**  
    $ path -e go.mod  
      .mod

# オプション
**-a**, **--absolute**
:   絶対パスを表示します。

**-b**, **--basename**
:   ベースネーム（ファイル名）を表示します。

**-c**, **--canonical**
:   正規化されたパスを表示します（デフォルト）。

**-d**, **--dirname**
:   ファイル名を除いたパスを表示します。

**-e**, **--extension**
:   拡張子を表示します。

**-h**, **--help**
:   ヘルプメッセージを表示します。

**-v**, **--version**
:   pathコマンドのバージョンを表示します。

# 終了ステータス
**0**
:   成功

**1**
:   pathコマンドの引数指定でエラー

# バグ
GitHub Issuesを参照してください。URL：https://github.com/nao1215/mimixbox/issues

# ライセンス
MimixBoxプロジェクトは、MIT License条文およびApache License 2.0条文の下でライセンスされています。