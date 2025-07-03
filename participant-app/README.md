# クイズ参加者画面 (Quiz Participant App)

Next.js 14で構築されたスマートフォン向けクイズ参加システムのフロントエンド。

## 機能

- **ニックネーム入力画面**: クイズ参加前のニックネーム登録
- **回答画面**: A/B/C/D の4つの選択肢から回答選択
- **リアルタイム更新**: WebSocketによる問題切り替えと投票終了通知
- **待機画面**: 次の問題までの待機
- **結果表示**: 最終順位とスコア表示
- **ユニバーサルデザイン対応**: 
  - 大きなボタン (最小64px高さ)
  - 高コントラスト配色
  - 読みやすいフォント (Inter)
  - アクセシビリティ対応

## 技術スタック

- **フレームワーク**: Next.js 14 (App Router)
- **言語**: TypeScript
- **スタイリング**: Tailwind CSS
- **リアルタイム通信**: WebSocket
- **状態管理**: React Hooks
- **テスト**: Jest + React Testing Library
- **リント**: ESLint + Next.js Rules

## 開発・実行

### 開発サーバー起動
```bash
npm run dev
```

### 本番ビルド
```bash
npm run build
npm start
```

### テスト実行
```bash
# Jestテスト実行
npm run test

# テスト（ウォッチモード）
npm run test:watch

# カバレッジ付きテスト
npm run test:coverage
```

### Lint・フォーマット
```bash
# ESLintチェック
npm run lint

# 自動修正
npm run lint:fix
```

## 環境変数

`.env.local` ファイルを作成して以下を設定:

```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 画面構成

1. **ニックネーム入力** (`/`)
   - バリデーション: 2-20文字
   - エラーハンドリング

2. **待機画面**
   - 進捗バー表示
   - ローディングアニメーション

3. **回答画面**
   - 4択ボタン (A/B/C/D)
   - 選択状態のハイライト
   - 回答変更可能 (投票終了まで)

4. **結果画面**
   - 順位とスコア表示
   - 統計情報
   - 再参加ボタン

## API連携

Go言語バックエンドとの連携:
- `/api/participants` - 参加者登録
- `/api/answers` - 回答送信
- `/api/quiz/{id}/session` - クイズセッション取得
- `/ws` - WebSocket接続

## コード品質

### リント状況
- **ESLint**: 100%成功 ✅
- **TypeScript**: 厳密な型チェック
- **Next.js**: App Router規約準拠

### テスト状況
- **コンポーネントテスト**: NicknameInputコンポーネント
- **React Testing Library**: ユーザー中心のテスト
- **モック**: API・WebSocket統合テスト対応

### ビルド確認
- **Next.js**: 静的エクスポート対応
- **TypeScript**: コンパイルエラーなし
- **最適化**: プロダクション向けビルド

## ユニバーサルデザイン

- **視覚的配慮**: 高コントラスト、大きな文字
- **操作性**: 大きなタッチターゲット
- **アクセシビリティ**: ARIAラベル、キーボードナビゲーション
- **レスポンシブ**: スマートフォン最適化
- **高品質**: ESLint 100%成功、型安全性確保