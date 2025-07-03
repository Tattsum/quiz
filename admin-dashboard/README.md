# クイズ管理システム - 管理者ダッシュボード

Nuxt3で構築されたリアルタイムクイズ大会の管理者向けダッシュボードです。

## 主な機能

### 🔐 認証
- 管理者ログイン画面
- JWT認証によるセッション管理

### 📝 問題管理
- 問題作成フォーム（テキスト、4択、画像アップロード、正解選択）
- 問題一覧表示・編集・削除
- カテゴリ・難易度による検索・フィルタリング

### 🎮 クイズ制御
- リアルタイムクイズセッション開始・制御
- 問題切り替え・投票終了・結果表示ボタン
- WebSocket接続による双方向通信

### 📊 リアルタイム統計
- 円グラフによる回答状況表示
- 参加者数・回答率のリアルタイム更新
- 選択肢別回答分布の可視化

### 🏆 ランキング表示
- 最終ランキング表示（表彰台形式）
- 詳細統計情報（正解率、平均回答時間など）
- 過去セッションの結果参照

### 📱 レスポンシブデザイン
- モバイル・タブレット・デスクトップ対応
- ハンバーガーメニューによるナビゲーション
- 全画面表示モード対応

## 技術スタック

- **フレームワーク**: Nuxt3
- **スタイリング**: Tailwind CSS
- **アイコン**: Heroicons
- **チャート**: Chart.js + Vue-ChartJS
- **UI コンポーネント**: Headless UI
- **通信**: WebSocket (リアルタイム) + REST API
- **テスト**: Vitest + Vue Test Utils
- **リント**: ESLint + Prettier

## セットアップ

### 依存関係のインストール

```bash
npm install
```

### 開発サーバーの起動

```bash
npm run dev
```

開発サーバーは `http://localhost:3000` で起動します。

### 本番ビルド

```bash
npm run build
npm run preview
```

## 環境設定

`.env` ファイルで以下の環境変数を設定してください：

```env
API_BASE_URL=http://localhost:8080
WS_BASE_URL=ws://localhost:8080
```

## API エンドポイント

### 認証
- `POST /api/admin/login` - 管理者ログイン

### 問題管理
- `GET /api/questions` - 問題一覧取得
- `POST /api/questions` - 問題作成
- `GET /api/questions/:id` - 問題詳細取得
- `PUT /api/questions/:id` - 問題更新
- `DELETE /api/questions/:id` - 問題削除

### クイズセッション
- `POST /api/quiz/sessions` - セッション作成
- `GET /api/quiz/sessions/active` - アクティブセッション取得
- `POST /api/quiz/sessions/:id/start-question` - 問題開始
- `POST /api/quiz/sessions/:id/end-voting` - 投票終了
- `POST /api/quiz/sessions/:id/end` - セッション終了

### WebSocket
- `ws://localhost:8080/ws` - リアルタイム通信

## ディレクトリ構造

```
admin-dashboard/
├── assets/
│   └── css/
│       └── main.css          # Tailwind CSS設定
├── components/
│   └── RealtimeChart.vue     # リアルタイムチャートコンポーネント
├── layouts/
│   └── default.vue           # デフォルトレイアウト
├── middleware/
│   └── auth.js               # 認証ミドルウェア
├── pages/
│   ├── dashboard.vue         # ダッシュボード
│   ├── login.vue             # ログイン画面
│   ├── ranking.vue           # ランキング表示
│   ├── questions/
│   │   ├── index.vue         # 問題一覧
│   │   └── create.vue        # 問題作成
│   ├── quiz-control.vue      # クイズ制御
│   └── quiz-control/
│       └── realtime.vue      # リアルタイム表示
└── nuxt.config.ts            # Nuxt設定
```

## 使用方法

1. **ログイン**: `/login` で管理者認証
2. **問題作成**: `/questions/create` で新しい問題を作成
3. **クイズ開始**: `/quiz-control` でセッション作成・制御
4. **リアルタイム表示**: `/quiz-control/realtime` で参加者向け表示
5. **ランキング確認**: `/ranking` で結果確認

## 開発・テスト

### テスト実行

```bash
# Vitestテスト実行
npm run test

# テスト（ウォッチモード）
npm run test:watch

# カバレッジ付きテスト
npm run test:coverage
```

### リント・フォーマット

```bash
# ESLintチェック
npm run lint

# 自動修正
npm run lint:fix
```

### ビルド確認

```bash
# プロダクションビルド
npm run build

# ビルド結果プレビュー
npm run preview
```

## テスト状況

- **総テスト数**: 14テスト
- **成功率**: 100% ✅
- **カバレッジ対象**: RealtimeChartコンポーネント
- **Chart.js統合**: 適切なモックによるテスト

### テスト内容
- コンポーネントレンダリング
- プロパティ受け渡し
- WebSocket接続
- データ表示ロジック
- ユーザーインタラクション

## 特徴

- **リアルタイム性**: WebSocketによる即座の状態同期
- **直感的UI**: 分かりやすいアイコンとレイアウト
- **モバイル対応**: どのデバイスからでも操作可能
- **データ可視化**: 円グラフやプログレスバーによる視覚的表示
- **全画面モード**: プレゼンテーション向けの表示機能
- **高品質**: 100%テスト成功、ESLint準拠
