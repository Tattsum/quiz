# 開発ルール

## 作業前の必須タスク
- 作業開始前に必ずTODO.txtを確認・更新する
- 新機能や修正作業の際は、タスクをTODO.txtに追加してから実装を開始する

## 設計思想

### DDD (Domain Driven Design)
- ドメインロジックをビジネスルールの中心に置く
- ドメインモデルを明確に定義し、技術的関心事から分離する
- ubiquitous language（ユビキタス言語）を使用してドメインエキスパートと開発者の共通理解を図る

### t-wada思想
- テスト駆動開発（TDD）を基本とする
- テストは仕様書として機能させる
- 設計品質を保つためのテストを重視する
- 実装より先にテストを書く習慣を維持する

### BFF (Backend for Frontend)
- フロントエンド専用のAPIレイヤーを設ける
- クライアント特有の要求に最適化されたAPIを提供する
- 複数のバックエンドサービスを統合し、フロントエンドの複雑性を軽減する

## プロジェクト構成

### Go言語バックエンドAPI (ルート)
- internal/models: ドメインモデルの定義
- internal/services: ビジネスロジックの実装
- internal/handlers: BFFとしてのAPI層の実装
- main.go: メインアプリケーション
- テストファイルは対応するソースファイルと同じディレクトリに配置

### Nuxt3管理ダッシュボード (admin-dashboard/)
- 管理者向けリアルタイムクイズ管理システム
- 問題作成・編集、セッション制御、統計表示
- WebSocketによるリアルタイム更新対応

### Next.js参加者アプリ (participant-app/)  
- スマートフォン向けクイズ参加システム
- ユニバーサルデザイン対応
- WebSocketによるリアルタイム回答

## テスト実行コマンド

### Go言語バックエンド
```
go test ./...
```

### Nuxt3管理ダッシュボード
```
cd admin-dashboard
npm run test  # (テストが設定されている場合)
```

### Next.js参加者アプリ
```
cd participant-app
npm run test  # (テストが設定されている場合)
```

## リント・フォーマットコマンド

### Go言語バックエンド
#### 基本コマンド
```
go fmt ./...
go vet ./...
```

#### 推奨Makeタスク（commit前に必須実行）
```
make check     # フォーマット、リント、vet、テストを一括実行
make fmt       # gofumptによるコードフォーマット
make lint      # golangci-lint v2による静的解析
make test      # テスト実行
```

#### ツールのインストール
```
make install-tools  # gofumptとgolangci-lintをインストール
```

### Nuxt3管理ダッシュボード
```
cd admin-dashboard
npm run lint
npm run dev  # 開発サーバー起動
npm run build  # ビルド
```

### Next.js参加者アプリ
```
cd participant-app
npm run lint
npm run dev  # 開発サーバー起動
npm run build  # ビルド
```

## Git運用ルール

- 作業毎に必ずcommitする（まとめてcommitしない）
- 各タスク完了時に個別にcommitを実行する
- すべての作業完了後にプルリクエスト（PR）を作成する
