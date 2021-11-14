# MimixBoxとは
MimixBoxは、シングルバイナリに多数のUnix（Linux）ユーティリティコマンドを詰め込んだコマンドである。BusyBoxを模倣（mimic）し、かつLinux上でのみの動作を期待したツールである。  

# MimixBoxのあり方（目標）
- BusyBox（組み込み用ユーティリティ）の代替品を目指さない
- Linuxデスクトップ環境（リッチな環境）でのシェル操作を簡便にさせる
- ジョークコマンドや多数のオプションも受け入れる
- 独自ユーティリティコマンドの実験環境とする
- manページは、独自ユーティリティコマンドのみ。
  
MimixBoxは、作者がGolangを学習する題材として、開発が始まった。作者は、シェルのユーティリティ開発と独自シェル開発を考えていた。しかし、それらを別プロジェクトとして開発するよりも、一つにまとめた方が配布時に便利だと判断した。この判断から、「BusyBoxの類似コマンドを開発しよう」と思い立つまでに、時間はかからなかった。  
  
出自が学習題材であるため、Unix哲学（例：一つのことを行い、またそれをうまくやるプログラムを書け）から外れた外部仕様を許容している。MimixBoxは、作者のオモチャ箱にすぎない。オモチャ箱には好きな物が全て詰め込まれているべきだ。だから、何でもMimixBoxに組み込む。

# インストールは最小限
MimixBoxが提供するコマンドは、まだ低機能である。言い換えれば、システムに異常をきたす可能性がある。例えば、作者の環境（Ubuntu）ではGUIが起動しなくなった（原因はcatコマンドのバグ）。  
  
そのため、インストールに関しては謙虚な動作仕様とした。具体的には、--installオプションでシンボリックリンクを作成する際、同名のコマンドがシステムに存在すればシンボリックリンクを作成しない。  
  
システムの状態に関わらず、全コマンドのシンボリックリンクを作成する方法も提供している。2021年11月段階ではオススメしない。単体テストもしていないコマンドだらけだからだ。ドッグフーディングしたい場合は、DockerやRaspberry Piを使用して遊ぶ事を推奨する。

# システムが壊れた場合
MimixBoxのインストール後に、システムが明らかにおかしい。そんな場合は、怒りを鎮めてから[MimixBoxのIssues](https://github.com/nao1215/mimixbox/issues)に報告していただきたい。  
  
冗談はさておき、MimixBoxのシンボリックリンクをシステムから取り除こう。具体的な手順は、以下の通りである。  

1. PCの電源を落とす。
2. レスキューモードで起動する。
3. $ sudo mimixbox --remove $(シンボリックリンクが存在するディレクトリ)を実行する。  
   例：sudo mimixbox --remove /usr/local/bin
4. 再起動する。

--removeオプションは、mimixboxが作成したシンボリックリンクを削除する。
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
