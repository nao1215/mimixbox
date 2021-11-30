<div align="center">
<img src="https://github.com/nao1215/mimixbox/blob/main/docs/images/logo.jpg" width="100">
</div>

[![Build](https://github.com/nao1215/mimixbox/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nao1215/mimixbox/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml/badge.svg?branch=main&event=push)](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml)
[![IntegrationTest](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml/badge.svg?event=push)](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml)
![GitHub](https://img.shields.io/github/license/nao1215/mimixbox)
![GitHub all releases](https://img.shields.io/github/downloads/nao1215/mimixbox/total)
![Lines of code](https://img.shields.io/tokei/lines/github/nao1215/mimixbox?style=plastic)

[[英語](../../../README.md)]
# MimixBox - mimic BusyBox on Linux
MimixBoxは、シングルバイナリの中に多数のUnixコマンドを持ちます。しかし、MimixBoxはBusyBoxと異なる目標を持ちます。デスクトップ環境での使用を念頭に置き、組み込み環境での使用は考えていません。  
同様に、MimixBoxプロジェクトメンテナは、Coreutilsに代表されるような一般的なコマンドから実験的なコマンド（applets）まで幅広く組み込み事を計画しています。

# インストール方法
[リリースページ](https://github.com/nao1215/mimixbox/releases)で、ソースコードとバイナリをzip形式とtar.gz形式で配布しています。OSとCPUアーキテクチャにあうバイナリを選択してください。  
例えば、Linux（amd64）の場合、次のコマンドでMimixBoxとドキュメントをインストールできます。

```
$ tar xf mimixbox-0.0.1-linux-arm64.tar.gz
$ cd mimixbox-0.0.1-linux-arm64
$ sudo ./installer.sh
```

Golang開発環境がある場合、以下の方法でもインストールできます。この方法では、ドキュメントがインストールされません。
```
$ go install github.com/nao1215/mimixbox/cmd/mimixbox
$ mimixbox --install /usr/local/bin
```

# 開発方法
## ツール
下表は、MimixBoxプロジェクトで開発する上で用いるツール一覧です。
| ツール名 | 説明 |
|:-----|:------|
| go-licenses | 依存ライブラリのライセンス管理で使用|
| pandoc   | Markdownファイルをmanページに変換するために使用 |
| make   | ビルド、テスト、リリースなど使用 |
| gzip   | manページの圧縮で使用 |
| curl | ShellSpecのインストールで使用 |
| install   | MimixBoxバイナリとドキュメントをインストールするために使用 |
| docker| Docker内でMimixBoxをテストするために使用|
| debootstrap| Jail内でMimixBoxをテストするために使用|
|  shellspec   | 統合テスト時に使用|  

Debian系ディストリビューション（例：Debian、Ubuntu、Kali Linux、Raspberry Pi OS）を使用している場合、次のコマンドでツールをインストールできます。

```
$ sudo apt install build-essential curl git pandoc gzip docker.io debootstrap
$ go install github.com/google/go-licenses@latest
$ curl -fsSL https://git.io/shellspec | sh -s -- --yes
```
  
## ビルド方法
```
$ git clone https://github.com/nao1215/mimixbox.git
$ cd mimixbox
$ make build
```

# デバッグ方法
### Docker環境の作成方法
```
$ make docker

(注釈) Dockerイメージのビルドが完了次第、Dockerコンテナ内に入ります。
$ 
```
### Jail環境
``` 
$ make build                      ※ MimixBoxバイナリの作成
$ sudo make jail                  ※ /tmp/mimixbox/jailを作成

$ sudo chroot /tmp/mimixbox/jail /bin/bash   ※ Jail環境の中へ移動
# mimixbox --full-install /usr/local/bin     ※ MimixBox組み込みコマンドをJail内にインストール
```

# Roadmap
- Step1. 多くのUnixコマンドを開発(〜Version 0.x.x).
- Step2. コマンドオプションの拡充 (〜Version 1.x.x).
- Step3. コマンドに近代的な仕様を追加(〜Version 2.x.x)
  
現在、MimixBoxは、充分な数のコマンドを実装していません([サポートコマンドリストはこちら](./docs/introduction/en/CommandAppletList.md))。そのため、プロジェクトとしては、ドッグフーディングできる状態までコマンド数を増やすことが最優先の目標です。

次に、コマンドオプション数をCoreutilsや他のパッケージと同レベルまで増やします。ただし、Coreutilsと同じコマンドを開発することを目標としていません。しかし、Linuxユーザが期待するオプションはサポートしたいと考えています。

最後に、MimixBoxを独自のコマンドとする段階があります。catコマンドを改良した[bat](https://github.com/sharkdp/bat/blob/master/doc/README-ja.md#%E3%83%97%E3%83%AD%E3%82%B8%E3%82%A7%E3%82%AF%E3%83%88%E3%81%AE%E7%9B%AE%E6%A8%99%E3%81%A8%E6%97%A2%E5%AD%98%E3%81%AE%E9%A1%9E%E4%BC%BC%E3%81%97%E3%81%9FOSS)やlsコマンドを改良した[lsd](https://github.com/Peltoche/lsd)のように、コマンドの機能性を上げていきます。

# MimixBoxオリジナルコマンド
MimixBoxは、Coreutilsのようなパッケージに含まれていないオリジナルコマンドがあります。
|コマンド名 | 説明|
|:--|:--|
|[fakemovie](./docs/examples/fakemovie/en/fakemovie.md) | 画像にビデオ開始ボタンを付与|
|[ghrdc](./docs/examples/ghrdc/en/ghrdc.md) | GitHub Relaseのダウンロード数をカウント|
|[path](./docs/examples/path/en/path.md) | PATH情報を操作するコマンド|
|[serial](./docs/examples/serial/en/serial.md) | ファイル名をシリアル番号付きにリネーム|

# 連絡先
「バグ発見」や「追加機能の要望」などのコメントを送りたい場合、以下の連絡先のいずれかを使用してください。

- [GitHub Issue](https://github.com/nao1215/mimixbox/issues)
- [Twitter@mimixbox156](https://twitter.com/mimixbox156)

# ライセンス
MimixBoxプロジェクトは、MITライセンス条文およびApache License Version 2.0ライセンス条文の下でライセンスされています。詳細は[LICENSE](../../../LICENSE)および[NOTICE](../../../NOTICE)をご確認ください。
