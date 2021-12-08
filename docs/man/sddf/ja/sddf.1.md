% SDDF(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% 2021年12月

# 名前

serial –  重複するファイルを探し、削除する

# 書式

**sddf** [OPTIONS] PATH

# 説明
**sddf**は、任意のディレクトリ以下にある重複ファイルを探し出し、  
そのリスト（デフォルト：duplicated-file.sddf）を作成します。  
リストを指定して実行した場合は、リスト内容に基づいてファイルを削除します。

# 例
**カレントディレクトリ以下から重複ファイルを探索**  

    $ sddf .  

**重複ファイルを削除**  

    $  sddf duplicated-file.sddf  

# オプション
**-o, **--output**
:   重複ファイルリストのファイル名を指定します。

**-h**, **--help**
:   ヘルプメッセージを表示します。

**-v**, **--version**
:   sddfコマンドのバージョンを表示します。

# 終了ステータス
**0**
:   成功

**1**
:   sddfコマンドの引数指定でエラー、もしくはファイル操作中のエラー

# バグ
GitHub Issuesを参照してください。URL：https://github.com/nao1215/mimixbox/issues

# ライセンス
MimixBoxプロジェクトは、MIT License条文およびApache License 2.0条文の下でライセンスされています。