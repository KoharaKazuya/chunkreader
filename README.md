# ReadChunker

- `io.Reader` の一種
- 内部にはユーザー定義の ChunkReader を持つ (好きなサイズの `[]byte` を Read の結果にできる)
- 外部には通常の `Read` を提供する
- 外部からの `Read` のときに渡された `[]byte` のサイズが小さくても ChunkReader から受け取ったデータは破棄せず次の Read 時に読み出す

8 バイトずつ変換しながら処理したい、みたいなユースケースを想定している
