# リアルタイムクイズ大会システム

リアルタイムクイズ大会を開催するための統合システムです。Go言語バックエンドAPI、Nuxt3管理ダッシュボード、Next.js参加者アプリから構成されています。

## 機能

- 管理者認証（JWT）
- 問題作成・編集・削除
- 画像アップロード
- セッション管理（クイズ開始/終了）
- 参加者登録（ニックネーム）
- 回答送信・変更
- リアルタイム集計結果
- ランキング機能
- WebSocketによるリアルタイム更新

## システム構成

### 🔧 Go言語バックエンドAPI (ルート)
- **言語**: Go 1.24+
- **フレームワーク**: Gin  
- **データベース**: PostgreSQL
- **認証**: JWT
- **リアルタイム通信**: WebSocket
- **パスワードハッシュ**: bcrypt

### 🎛️ Nuxt3管理ダッシュボード (admin-dashboard/)
- **フレームワーク**: Nuxt3
- **スタイリング**: Tailwind CSS
- **機能**: 問題管理、セッション制御、リアルタイム統計表示

### 📱 Next.js参加者アプリ (participant-app/)
- **フレームワーク**: Next.js 14 (App Router)
- **言語**: TypeScript
- **スタイリング**: Tailwind CSS  
- **機能**: スマートフォン向けクイズ参加、ユニバーサルデザイン対応

## セットアップ

### 1. 全体の依存関係インストール

#### Go言語バックエンド
```bash
go mod download
```

#### Nuxt3管理ダッシュボード
```bash
cd admin-dashboard
npm install
```

#### Next.js参加者アプリ  
```bash
cd participant-app
npm install
```

### 2. 環境設定

`.env.example`をコピーして`.env`ファイルを作成：

```bash
cp .env.example .env
```

`.env`ファイルを編集してデータベース設定等を行います：

```env
# データベース設定
DB_HOST=localhost
DB_PORT=5432
DB_USER=quiz_user
DB_PASSWORD=quiz_password
DB_NAME=quiz_db
DB_SSLMODE=disable

# JWT設定
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRES_HOURS=24

# サーバー設定
PORT=8080
```

### 3. データベースセットアップ

PostgreSQLデータベースを作成し、`database_schema.sql`を実行：

```bash
# PostgreSQLにログイン
psql -U postgres

# データベースとユーザーを作成
CREATE DATABASE quiz_db;
CREATE USER quiz_user WITH PASSWORD 'quiz_password';
GRANT ALL PRIVILEGES ON DATABASE quiz_db TO quiz_user;

# スキーマを適用
\c quiz_db
\i database_schema.sql
```

### 4. 管理者ユーザーの作成

初回起動前に管理者ユーザーを手動で作成する必要があります：

```sql
-- パスワードハッシュを生成（例: password123）
INSERT INTO administrators (username, password_hash, email) 
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin@example.com');
```

### 5. アプリケーション起動

#### Go言語バックエンド
```bash
go run main.go
```
サーバーは `http://localhost:8080` で起動します。

#### Nuxt3管理ダッシュボード  
```bash
cd admin-dashboard
npm run dev
```
管理ダッシュボードは `http://localhost:3000` で起動します。

#### Next.js参加者アプリ
```bash
cd participant-app
npm run dev  
```
参加者アプリは `http://localhost:3001` で起動します。

## API仕様

詳細なAPI仕様は`api_design.md`を参照してください。

### 主要エンドポイント

#### 管理者認証
- `POST /api/admin/login` - ログイン
- `POST /api/admin/logout` - ログアウト
- `GET /api/admin/verify` - トークン検証

#### 問題管理
- `GET /api/admin/quizzes` - 問題一覧
- `POST /api/admin/quizzes` - 問題作成
- `PUT /api/admin/quizzes/{id}` - 問題更新
- `DELETE /api/admin/quizzes/{id}` - 問題削除

#### セッション管理
- `POST /api/admin/session/start` - セッション開始
- `POST /api/admin/session/next` - 次の問題
- `POST /api/admin/session/end` - セッション終了
- `GET /api/session/status` - セッション状態取得

#### 参加者・回答
- `POST /api/participants/register` - 参加者登録
- `POST /api/answers` - 回答送信
- `PUT /api/answers/{id}` - 回答変更

#### 集計・ランキング
- `GET /api/results/current` - 現在の集計結果
- `GET /api/ranking/overall` - 総合ランキング

#### WebSocket
- `WS /api/ws/results` - リアルタイム結果更新

## 使用例

### 1. 管理者ログイン

```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password123"}'
```

### 2. 問題作成

```bash
curl -X POST http://localhost:8080/api/admin/quizzes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "question_text": "Go言語の開発元は？",
    "option_a": "Google",
    "option_b": "Microsoft", 
    "option_c": "Apple",
    "option_d": "Meta",
    "correct_answer": "A"
  }'
```

### 3. 参加者登録

```bash
curl -X POST http://localhost:8080/api/participants/register \
  -H "Content-Type: application/json" \
  -d '{"nickname": "参加者A"}'
```

### 4. 回答送信

```bash
curl -X POST http://localhost:8080/api/answers \
  -H "Content-Type: application/json" \
  -d '{
    "participant_id": 1,
    "quiz_id": 1,
    "selected_option": "A"
  }'
```

## ファイル構成

```
.
├── main.go                        # Go バックエンドメイン
├── go.mod                         # Go モジュール定義
├── .env.example                   # 環境変数テンプレート
├── database_schema.sql            # データベーススキーマ
├── api_design.md                  # API設計書
├── CLAUDE.md                      # 開発ルール
├── README.md                      # プロジェクト説明
├── internal/                      # Go バックエンド
│   ├── database/                  # データベース接続
│   ├── handlers/                  # HTTPハンドラ
│   ├── middleware/                # ミドルウェア
│   ├── models/                    # データモデル
│   └── services/                  # ビジネスロジック
├── uploads/                       # アップロードファイル
│   └── images/
├── admin-dashboard/               # Nuxt3管理ダッシュボード
│   ├── components/                # Vueコンポーネント
│   ├── pages/                     # ページ
│   ├── layouts/                   # レイアウト
│   ├── middleware/                # ミドルウェア
│   ├── assets/                    # アセット
│   ├── nuxt.config.ts             # Nuxt設定
│   ├── package.json               # 依存関係
│   └── README.md                  # 管理ダッシュボード説明
└── participant-app/               # Next.js参加者アプリ
    ├── src/
    │   ├── app/                   # App Router
    │   ├── components/            # Reactコンポーネント
    │   ├── lib/                   # ユーティリティ
    │   └── types/                 # TypeScript型定義
    ├── next.config.js             # Next.js設定
    ├── package.json               # 依存関係
    └── README.md                  # 参加者アプリ説明
```

## セキュリティ

- JWT認証によるAPI保護
- bcryptによるパスワードハッシュ化
- CORS設定対応
- レート制限実装
- ファイルアップロード制限

## 開発・デプロイ

### 開発モード

```bash
# 開発モードで起動（詳細ログ）
GIN_MODE=debug go run main.go
```

### プロダクションビルド

```bash
# バイナリビルド
go build -o quiz-server main.go

# 実行
./quiz-server
```

### Docker対応

Dockerfileを追加する場合：

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o quiz-server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/quiz-server .
COPY --from=builder /app/.env .
CMD ["./quiz-server"]
```

## ライセンス

MIT License