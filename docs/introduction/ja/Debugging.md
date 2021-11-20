# デバッグ環境（ドッグフーディング環境）
MimixBoxは、Version1.0.0となるまで開発環境にインストールしてはいけない。2021年11月段階で、ロクにテストされていないコマンドしか無い。そのため、システム上のCoreutilsやBusyBoxを差し置いてMimixBox（Applet）が使用された場合、システムが不安定な動作をする可能性が非常に高い。  

例えば、作者の環境（Ubuntu）では、MimixBoxのcatコマンドに不具合があったため、GUIが起動しなくなった。このようなトラブルシューティングは、やや難易度が高い。そのため、ドッグフーディング用の安全な環境が必要である。

## デバッグ環境の例
1. Docker
2. jail環境
3. Raspberry Pi環境（イメージバックアップを残すと良い）
4. Virtual Box仮想環境（オススメしない）

上記は、オススメ順である。DockerとJail環境を利用したデバッグ方法のみ以下に示す。上記の3〜4は、やや面倒な方法であるため、手順の説明は書かない予定である。将来的にディストリビューションパッケージになれば、上記の3〜4.は現実的なデバッグ手段になるが、まだ早い。

### Docker環境の作り方
```
# sudo apt install docker.io      (注釈) Ubuntu環境の場合
$ make docker
(注釈) Docker imageのビルドが完了すれば、コンテナの中に入る
$ 
```
### Jail環境の作り方
``` 
$ sudo apt install debootstrap    (注釈) Debian系の環境でdebootstrapをインストールしていない場合
$ make build                      (注釈) mimixboxバイナリの作成
$ sudo make jail                  (注釈) jail環境を"/tmp/mimixbox/jail"に作成

$ sudo chroot /tmp/mimixbox/jail /bin/bash   (注釈) jailに入る
# mimixbox --full-install /usr/local/bin     (注釈) jail内でmimixboxをインストール
```

# MimixBoxにおけるロギング（検討中）
MimixBoxは動作が不安定な状態なので、ロギング機能を設ける事にした。将来的にMimixBoxが安定動作すれば、複数のユーザがMimixBoxが提供するコマンドをほぼ同時に実行する可能性が考えられる。その場合、複数のプロセス（MimixBox）が、同時に一つのログファイルに対して書き込みを行う。この時、ファイルへログが期待通りに書き込まれるかどうかは保証されない。  
そこで、各プロセスは名前付きパイプにログを書き込む。名前付きパイプへの書き込みは、PIPE_BUFサイズ以下の書き込みであればアトミックであるため、ログメッセージが競合しない（混ざらない）。MimixBoxのログサイズは、PIPE_BUF（最低512Byte）を下回る想定であり、アトミック性が高い確率で保証される。  
ログ読み込みとログファイルへの書き込みは、ロギング専用のデーモンが行う。
![MimixBoxロギングの流れ](/docs/images/debug_logging.jpg "MimixBoxロギングの流れ")

大げさな設計である。本当はアトミック性など気にしておらず、「名前付きパイプ」と「systemdに頼らないデーモン化」を試したかっただけ。