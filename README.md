# cevigo-voicevoxapi

by Richinosan

## 概要

[cevigo](https://github.com/gotti/cevigo)を使って[Docker版yomiage-VOICEVOX](https://github.com/richinosan/yomiage_VOICEVOX_verT-Docker)および[yomiage-VOICEVOX-VerT(pythonコードをいじる必要あり)](https://github.com/TaktstockJp/yomiage_VOICEVOX_verT)で使えるvoicevoxapiを雑に実装します。ちょっと強引な方法で実装しているので他環境だと使いにくいかも。

## 起動方法とコマンドライン引数
ReleaseからDLしたZIPファイルを解凍し、main.exeと同じディレクトリにあるstart.batを起動してください。

### batの中身
```bat
start /b main.exe -api cevio  -port 10001
start /b main.exe -api cevioai -port 10002
```
### コマンドライン引数の詳細
| 引数名 | 説明 | value | default |
|:-----------:|:------------:|:------------:|:------------:|
| -api | 使用するCevioAPIを選択します | cevio / cevioai  | cevio|
| -port | 使用するportを選択します | 0-65535 | 10001|
| -debug | 簡易的なデバッグモードに切り替えます | True / False | False |

## docker-composeのサンプル
enviromentを追加します。
カンマ区切りで複数のサービスを設定できます。
```yaml
version: '3'
services:
  yomiage_voicevox:
    container_name: yomiage_voicevox
    build: 
      context: .
      dockerfile: Dockerfile
    tty: true
    volumes:
      - ./yomiage_VOICEVOX:/yomiage_VOICEVOX
    environment:
      TZ: "Asia/Tokyo"
      TOKEN: "TOKEN"
      COMMAND_SYNTHAX : "!"
      COMMENT_SYNTHAX : ">"
      OTHER_BOTS_SYNTHAX : "$$,%,*"
      USE_VOICEVOX: "True"
      USE_COEIROINK: "True"
      USE_LMROID: "False"
      USE_SHAREVOX: "False"
      FLAG_LIST_PATH: "data/flag_list.csv"
      VOICE_LIST_PATH: "data/voice_list.csv"
      WORD_LIST_PATH: "data/word_list.csv"
      STYLE_SETTING_PATH: "data/style_setting.csv"
      OTHER_VOICEVOX_APP: "Cevio,CevioAI"
      OTHER_VOICEVOX_PORT: "10001,10002"
    entrypoint: "python3 /yomiage_VOICEVOX/discordbot.py"
```

## yomiage-VOICEVOX-VerTでのpython追記部分
下記に追加

第一引数はサービス名(任意の名称)を入力、第二引数はポート番号を入力

for_developer/discordbot_functions.py
```python
class room_information():
  async def reload(self):
    self.createVoiceVoxGenerator('Cevio', '10001')
    self.createVoiceVoxGenerator('CevioAI', '10002')
```

## 更新履歴
### v1.1
  - Audioqueryをより一般的なvoicevoxapi用に修正(jsonのMarshalじゃなくして出力を直書きすることで他のvoicevoxapiに準じた)

### v1.0
  - githubにアップ