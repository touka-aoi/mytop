## プロセスごとのCPU, メモリ使用率を出力する
otelを用いてコレクターに送信する
- 送信先: localhost:4317
- 送信間隔: 5s
- 送信プロセス: top10

### 使用ライブラリ
https://pkg.go.dev/github.com/prometheus/procfs#section-readme

