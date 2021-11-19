% MIMIXBOX(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% 2021年11月

# 名前

MimixBox – BusyBoxの模造品。シングルバイナリ内に多数のUnixコマンド（applet）を持つ。

# 書式

**mimixbox** [applet [arguments]...] [OPTIONS]

# 説明
**mimixbox**は、BusyBoxのようにシングルバイナリ内に多数のUnixコマンドを持ちます。  
しかし、BusyBoxとは別の立ち位置を目指します。具体的には、組み込み環境ではなく、  
デスクトップ環境で使う事を想定しています。また、ビルトインするコマンド（applet）は、  
Coreutils等が提供する基本的な内容から実験的なコマンドまで、幅広く取り揃える予定です。

# コマンド（applet）一覧
**一般的なUnixコマンド（applet）**  
basename cat chroot echo fakemovie false ghrdc ischroot mbsh mkdir  
mkfifo mv path rm rmdir serial sh sleep tac touch true which

# オプション
**-i**, **--install**
:   システム上に同名コマンドが存在しない場合、appletのシンボリックリンクを作成します。

**-f**, **--full-install**
:   システム状態に関わらず、appletのシンボリックリンクを作成します。

**-h**, **--help**
:   ヘルプメッセージを表示します。

**-l**, **--list**
:   MimixBoxが提供するコマンド（applet）を表示します。

**-r**, **--remove**
:   MimixBoxが提供するコマンド（applet）のシンボリックリンクを削除します。

**-v**, **--version**
:   mimixboxコマンドのバージョンを表示します。

# 終了ステータス
**0**
:   成功

**1**
:   存在しないapplet名を指定した場合、オプション不正の場合、appletでエラーが発生した場合

**2**
:   一部のappletで発生するエラー（例：ischrootなど）

# バグ
GitHub Issuesを参照してください。URL：https://github.com/nao1215/mimixbox/issues

# ライセンス
MimixBoxプロジェクトは、MIT License条文およびApache License 2.0条文の下でライセンスされています。