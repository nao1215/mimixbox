# 本ドキュメントの目的
本ドキュメントでは、MimixBoxを安全にデバッグし、どのようにテストを追加するかを説明します。前提として、Version1.0.0となるまでMimixBoxを開発環境にフルインストールしてはいけません。2021年11月現在では、MimixBoxはテストが不十分です。

MimixBoxがシステム上のCoreutilsやBusyBoxを差し置いて使用された場合、システムが不安定な動作をする可能性があります。

# デバッグ環境（ドッグフーディング環境）
MimixBoxは、Docker環境内でコマンドを実行するのが安全です。[現在のDocker環境設定](../../../Dockerfile)では、MimixBox組み込みコマンドの全てが/usr/local/bin以下にシンボリックリンクされます。何らかのコマンドを追加した場合、以下の手順で動作確認してください。

```
$ make docker
$ (ここからDocker環境) 
```
MimixBoxシェルが未完成のため、Docker内ではBashを使用しています。また、どのコマンドがMimixBoxによって置き換えられているかは、以下のコマンドで確認できます。
```
$ mimixbox --list
```

# 単体テスト
単体テストは、golang言語の作法に則って作成しています。つまり、以下に示すディレクトリ構成のように、「テスト対象.go」と「テスト対象_test.go」が同じディレクトリ階層に存在します。
```
   └── lib
           ├── file.go
           ├── file_test.go
```
MimixBoxプロジェクトでは、単体テストカバレッジ100%を目指してはいません。ただし、MimixBoxライブラリ（mimixbox/internal/lib）は汎用的なコードが多いため、可能な限りカバレッジ100%にしたい気持ちがあります。  

単体テストは、makeコマンドで簡単に実行できます。
```
$ make ut
```
MimixBoxプロジェクトでは、"$ git push"を検知した場合、GitHub Actionで単体テストを実行しています。そのため、コードをpushする前に、単体テストを実行する事をオススメします。

# 結合テスト
結合テストでは、[ShellSpec](https://github.com/shellspec/shellspec)で各コマンドの振る舞いを確認します。テストコードは、testディレクトリ以下に格納しています。
```
test/
├── it
│      └── shellutils
│               └── echoTest.sh  ※ テスト用のシェルスクリプト関数定義
├── spec
│      ├── echo_spec.sh     ※ テスト期待値を定義
│      └── spec_helper.sh
└── ut
```
結合テストも単体テスト同様、makeコマンドで簡単に実行できます（GitHub Actionを使っている所も同様です）
```
$ make it
```
ただし、MimixBoxがインストールされていない環境でmakeコマンドを実行しても、Coreutilsのテストにしかなりません。統合テストがHOST環境で動作することが確認できたら、最後にDocker環境で実行してください。
```
$ make docker

$ ls           ※ ここからDocker環境
do_integration_test.sh  integration_tests   ※ HOST環境からコピーされたテストファイル

$ ./do_integration_test.sh   ※ 統合テスト実行
Running: /bin/sh [sh]
......

Finished in 0.07 seconds (user 0.05 seconds, sys 0.02 seconds)
6 examples, 0 failures
```

# HOST環境にMimixBoxをインストールする場合
## インストールオプションの使い分け
MimixBoxは、組み込みコマンドに対するシンボリックリンクを作成するオプションを2つ提供します。  

1つ目は、--installオプションです。同名のコマンドがシステムに存在する場合、MimixBoxはシンボリックリンクを作成しません。--installは、過去にシステムを壊した経験から、安全インストールとして導入しました。
```
$ sudo mimixbox --install /usr/local/bin
```
2つ目は、--full-installオプションです。システムの状態に関わらず、全コマンドのシンボリックリンクを作成します。現段階ではこちらのオプションは、非推奨です。
```
$ sudo mimixbox --full-install /usr/local/bin
```
## システムが壊れた場合（例：GUIが起動しない等）
MimixBoxのシンボリックリンクをシステムから取り除く必要があります。具体的な手順は、以下の通りです。  

1. PCの電源OFF
2. レスキューモードで起動
3. $ sudo mimixbox --remove $(シンボリックリンクが存在するディレクトリ)を実行。  
   例：sudo mimixbox --remove /usr/local/bin
4. 再起動
```
$ sudo ./mimixbox --remove /usr/local/bin/
Delete symbolic link: /usr/local/bin/fakemovie
Delete symbolic link: /usr/local/bin/mbsh
Delete symbolic link: /usr/local/bin/path
Delete symbolic link: /usr/local/bin/serial
Delete symbolic link: /usr/local/bin/sh
Delete symbolic link: /usr/local/bin/true
Delete symbolic link: /usr/local/bin/which
Delete symbolic link: /usr/local/bin/cat
Delete symbolic link: /usr/local/bin/echo
Delete symbolic link: /usr/local/bin/false
Delete symbolic link: /usr/local/bin/ghrdc
Delete symbolic link: /usr/local/bin/mkdir
```
余力があれば、[MimixBoxのIssues](https://github.com/nao1215/mimixbox/issues)にバグ報告をお願いします。

# MimixBoxにおけるロギング（検討中）
MimixBoxは動作が不安定な状態なので、ロギング機能を検討しています。  

将来的にMimixBoxが安定動作すれば、複数ユーザがMimixBox組み込みコマンドをほぼ同時に実行する可能性が考えられます。その場合、複数プロセス（MimixBox）が、同時に一つのログファイルに対して書き込みを行います。同時書き込みであるため、ログファイルへログが期待通りに書き込まれるかどうかは保証されません。  

そこで、各プロセスは名前付きパイプにログを書き込む仕様を検討しています。名前付きパイプへの書き込みは、PIPE_BUFサイズ以下の書き込みであればアトミックであり、ログメッセージが競合しません（混ざりません）。  

MimixBoxのログサイズは、PIPE_BUF（最低512Byte）を下回る想定であり、アトミック性が高い確率で保証されます。ログ読み込みとログファイルへの書き込みは、ロギング専用のデーモンが行います。
![MimixBoxロギングの流れ](/docs/images/debug_logging.jpg "MimixBoxロギングの流れ")

やや大げさな設計であり、ロギングデーモンが突然死した場合などの復旧に課題があります。そのため、ログファイルをユーザ単位で作成する方法も検討しています。  
