% GHRDC(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% 2021年9月

# 名前

ghrdc –  GitHub APIを利用して、レポジトリのReleaseファイルダウンロード数を表示します。

# 書式

**ghrdc** [OPTIONS] USER_NAME/RPOSITORY_NAME

# 説明
**ghrdc**は、レポジトリのReleaseファイルダウンロード数を表示します。  
デフォルトでは、最新リリースに対するダウンロード数を表示します。  
GitHub API認証を行わないため、以下の制約があります。  
- 1時間あたりに60回だけ実行できます。  
- Organizationレポジトリの情報を取得できません。  

# 例
**最新リリースのダウンロード数を取得**  
    $ ghrdc nao1215/serial  
      [Name(Version)]             :Version1.0.2: Release files with installer scripts.  
      [Release Date]              :2020-11-23 05:28:11 +0000 UTC  
      [Binary Download Count]     :177  
      [Source Code Download Count]:0  

# オプション
**-a**, **--all**
:   リリース毎のダウンロード数を表示します。

**-t**, **--total**
:   全リリースのダウンロード数（合計値）を表示します。

**-h**, **--help**
:   ヘルプメッセージを表示します。

**-v**, **--version**
:   ghrdcコマンドのバージョンを表示します。

# 終了ステータス
**0**
:   成功

**1**
:   ghrdcコマンドの引数指定でエラー、もしくはGitHub APIの実行時エラー

# バグ
GitHub Issuesを参照してください。URL：https://github.com/nao1215/mimixbox/issues

# ライセンス
MimixBoxプロジェクトは、MIT License条文およびApache License 2.0条文の下でライセンスされています。