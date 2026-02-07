# Copilot Instructions for np2misk

## プロジェクト概要

SpotifyのNow Playing情報を取得してMisskeyに自動投稿するGoアプリケーション。
クリーンアーキテクチャを採用し、domain/application/infrastructure/interfacesの4層構造で実装。

## アーキテクチャ

### ディレクトリ構成

- `internal/domain/entity/`: ビジネスロジックを含むエンティティ（Track, Note）
- `internal/domain/repository/`: リポジトリインターフェース定義
- `internal/application/`: アプリケーションサービス層（NowPlayingService）
- `internal/infrastructure/`: 外部サービスとの接続実装
  - `spotify/`: Spotify API クライアント
  - `misskey/`: Misskey API クライアント
- `internal/interfaces/config/`: 設定管理

### 依存関係ルール

- domain層は他のどの層にも依存しない
- application層はdomain層のみに依存する
- infrastructure層とinterfaces層はdomain層に依存する
- main.goですべての依存関係を注入する

## コーディング規約

### 基本方針

- コメントなしのコード：関数名や変数名で意図を明確にする
  - 実装コード内への説明的コメントは不要
  - 公開APIのGoDocコメントも省略可
  - パッケージレベルのdoc.goは不要
- シンプルさ優先：読みやすく保守しやすいコードを重視
- エラーハンドリング必須：すべてのエラーを適切に処理
- Go標準準拠：gofmt, golint, go vet に従う

### 命名規則

- パッケージ名：小文字のみ（`spotify`, `misskey`）
- エクスポートする識別子：PascalCase（`Track`, `NewNoteRepository`）
- プライベート識別子：camelCase（`rateLimiter`, `lastTrackTitle`）
- インターフェース名：名詞形（`SpotifyRepository`, `NoteRepository`）
- コンストラクタ：`New<TypeName>`の形式（`NewTrack`, `NewNowPlayingService`）
- レシーバー名：1〜2文字の短縮形（`s *NowPlayingService`, `r *spotifyRepository`）

### エラーハンドリング

- エラーは`fmt.Errorf`で詳細なコンテキストを付与してラップする
- `%w`を使用してエラーチェーンを保持する
- ログ出力は`log.Printf`を使用する
- 致命的なエラーは`log.Fatal`で終了する

### テストコード

- テストファイル名：`<ファイル名>_test.go`
- テスト関数名：`Test<関数名>`または`Test<型名>_<メソッド名>`
- テーブル駆動テストを使用する
- テストケースには`name`フィールドで説明を記載する

### 依存性注入

- すべての依存関係はコンストラクタで注入する
- インターフェースを使用して疎結合を実現する
- `main.go`で具象型を生成し、依存関係を組み立てる

### 並行処理

- `context.Context`を第一引数として受け取る
- `sync.Mutex`でクリティカルセクションを保護する
- ゴルーチン起動時は終了処理を明確にする

### 設定管理

- 環境変数で設定を管理する（`envconfig`パッケージ使用）
- `.env`ファイルのロードは`godotenv`を使用する
- デフォルト値を適切に設定する
- 設定値の変換用メソッドを提供する（例：`GetPollingInterval()`）

## レビュー時の観点

### [critical] 必須修正事項

- セキュリティ上の問題（認証情報のハードコード、インジェクション脆弱性）
- 致命的なバグ（nil参照、データ競合、デッドロック、リソースリーク）
- アーキテクチャ違反（依存関係の逆転、層の責務違反）
- context.Contextの不適切な使用

### [important] 重要な改善提案

- エラーハンドリングの不足や不適切な処理
- テストカバレッジの不足（特にエッジケース、エラーパス）
- パフォーマンスに影響する問題
- 保守性を著しく低下させる実装
- defer、close、cancelの適切な使用漏れ

### [nitpick] 軽微な改善提案

- 命名規則の統一（Goの慣用的な命名への準拠）
- 冗長なコードの削減（不要な変数、重複ロジック）
