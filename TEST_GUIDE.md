# クイズ大会システム テストガイド

このドキュメントでは、クイズ大会システムの包括的なテストについて説明します。

## 📋 テスト概要

本システムでは以下のテストが実装されています：

### 1. 単体テスト (Unit Tests)
- **Go言語API**: `internal/handlers/*_test.go`, `internal/services/*_test.go`
- **Nuxt3管理ダッシュボード**: `admin-dashboard/tests/`
- **Next.js参加者アプリ**: `participant-app/src/components/__tests__/`

### 2. 統合テスト (Integration Tests)
- **APIエンドポイント**: `integration_test.go`
- **データベース連携**: テスト用データベースを使用
- **認証フロー**: JWT認証の完全なフロー

### 3. WebSocket通信テスト
- **リアルタイム通信**: `internal/handlers/websocket_test.go`
- **同時接続**: 最大70人の同時接続テスト
- **メッセージング**: ブロードキャスト機能のテスト

### 4. パフォーマンステスト
- **同時接続**: 70人同時参加者登録テスト
- **負荷テスト**: システム全体の負荷テスト
- **WebSocket負荷**: 同時WebSocket接続テスト

### 5. E2Eテスト (End-to-End Tests)
- **管理者フロー**: ログイン〜クイズ作成〜セッション管理
- **参加者フロー**: 登録〜回答〜結果表示
- **統合フロー**: 管理者・参加者の連携動作

## 🚀 テスト実行方法

### 全テスト実行
```bash
./test_runner.sh
```

### 個別テスト実行

#### Go言語バックエンド
```bash
# 単体テスト
go test ./internal/...

# 統合テスト
go test -tags=integration ./...

# WebSocketテスト
go test -run TestWebSocket ./internal/handlers/

# パフォーマンステスト
go test -run TestConcurrent -timeout 300s
go test -run TestSystemLoad -timeout 300s

# カバレッジ付き実行
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### フロントエンド

**Nuxt3管理ダッシュボード:**
```bash
cd admin-dashboard
npm install
npm run test
npm run test:watch      # ウォッチモード
npm run test:coverage   # カバレッジ付き
```

**Next.js参加者アプリ:**
```bash
cd participant-app
npm install
npm run test
npm run test:watch      # ウォッチモード
npm run test:coverage   # カバレッジ付き
```

#### E2Eテスト
```bash
cd e2e
npm install
npm run cypress:run     # ヘッドレス実行
npm run cypress:open    # GUI実行
```

### 特定テストのみ実行
```bash
./test_runner.sh -u    # 単体テストのみ
./test_runner.sh -i    # 統合テストのみ
./test_runner.sh -p    # パフォーマンステストのみ
./test_runner.sh -e    # E2Eテストのみ
./test_runner.sh -f    # フロントエンドテストのみ
```

## 📊 テスト結果とレポート

### テスト結果の場所
- **実行ログ**: `test-results/test-execution-YYYYMMDD-HHMMSS.log`
- **カバレッジレポート**: `test-results/coverage.html`
- **Cypressビデオ**: `e2e/cypress/videos/`
- **スクリーンショット**: `e2e/cypress/screenshots/`

### パフォーマンス基準

#### 70人同時接続テスト
- **参加者登録**: エラー率 < 5%, 平均レスポンス時間 < 5秒
- **WebSocket接続**: 成功率 > 95%, 接続確立時間 < 10秒
- **回答送信**: エラー率 < 5%, 平均レスポンス時間 < 3秒

#### システム負荷テスト
- **エラー率**: < 2%
- **平均レスポンス時間**: < 2秒
- **スループット**: > 50 req/sec

## 🔧 テスト環境セットアップ

### 前提条件
- Go 1.21+
- Node.js 18+
- PostgreSQL (テスト用データベース)
- Docker (オプション)

### データベースセットアップ
```bash
# PostgreSQL テストデータベース作成
createdb quiz_test

# 環境変数設定
export DATABASE_URL="postgres://postgres:password@localhost:5432/quiz_test?sslmode=disable"
export JWT_SECRET="test_secret_key_for_testing_only"
export ENV="test"
```

### 依存関係インストール
```bash
# Go依存関係
go mod tidy

# フロントエンド依存関係
cd admin-dashboard && npm install && cd ..
cd participant-app && npm install && cd ..
cd e2e && npm install && cd ..
```

## 📝 テストファイル構成

```
quiz/
├── *_test.go                           # Go単体テスト
├── integration_test.go                 # Go統合テスト
├── performance_test.go                 # パフォーマンステスト
├── internal/
│   ├── handlers/
│   │   ├── quiz_test.go
│   │   ├── participant_test.go
│   │   └── websocket_test.go
│   └── services/
│       ├── auth_service_test.go
│       ├── jwt_service_test.go
│       └── quiz_service_test.go
├── admin-dashboard/
│   ├── vitest.config.ts
│   └── tests/
│       └── components/
│           └── RealtimeChart.test.ts
├── participant-app/
│   ├── jest.config.js
│   ├── jest.setup.js
│   └── src/components/__tests__/
│       └── NicknameInput.test.tsx
└── e2e/
    ├── cypress.config.js
    └── cypress/
        ├── e2e/
        │   ├── 01-admin-flow.cy.js
        │   ├── 02-participant-flow.cy.js
        │   └── 03-integrated-flow.cy.js
        └── support/
            ├── commands.js
            └── e2e.js
```

## 🎯 テストケース詳細

### API単体テスト
- クイズCRUD操作
- 参加者登録・管理
- 回答送信・更新
- セッション管理
- 認証・認可

### WebSocketテスト
- 接続・切断処理
- メッセージ送受信
- ブロードキャスト機能
- エラーハンドリング

### フロントエンドテスト
- コンポーネントレンダリング
- ユーザーインタラクション
- バリデーション
- WebSocket連携
- レスポンシブデザイン

### E2Eテスト
- 管理者ログインフロー
- クイズ作成・編集
- セッション制御
- リアルタイム統計
- 参加者登録・回答
- 結果表示・ランキング

## 🔍 デバッグとトラブルシューティング

### よくある問題

#### テストデータベース接続エラー
```bash
# データベースの状態確認
pg_isready -h localhost -p 5432

# テスト用データベースの再作成
dropdb quiz_test --if-exists
createdb quiz_test
```

#### WebSocketテスト失敗
```bash
# ポートの確認
lsof -i :8080

# サーバープロセスの確認
ps aux | grep quiz
```

#### フロントエンドテスト失敗
```bash
# Node.jsバージョン確認
node --version

# 依存関係の再インストール
rm -rf node_modules package-lock.json
npm install
```

### ログとデバッグ
- テスト実行時の詳細ログは `test-results/` ディレクトリに保存
- Cypressテストは動画とスクリーンショットを自動生成
- WebSocketテストはコンソールに接続状況を出力

## 📈 継続的インテグレーション

### GitHub Actions設定例
```yaml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: quiz_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Run tests
        run: ./test_runner.sh
```

## 🎉 テスト成功基準

すべてのテストが以下の基準を満たす必要があります：

- **単体テスト**: 100%成功
- **統合テスト**: 100%成功
- **パフォーマンステスト**: 基準値内
- **E2Eテスト**: 主要フロー100%成功
- **コードカバレッジ**: 80%以上

テストが失敗した場合は、ログを確認して原因を特定し、修正してから再実行してください。