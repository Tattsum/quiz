# Go言語クイズ大会システム REST API設計書

## 1. 管理者認証エンドポイント

### 1.1 管理者ログイン
- **エンドポイント**: `POST /api/admin/login`
- **説明**: 管理者の認証を行い、JWTトークンを発行
- **リクエスト**:
```json
{
  "username": "admin_user",
  "password": "password123"
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "ログインに成功しました",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-12-31T23:59:59Z",
    "admin": {
      "id": 1,
      "username": "admin_user",
      "email": "admin@example.com"
    }
  }
}
```

### 1.2 管理者ログアウト
- **エンドポイント**: `POST /api/admin/logout`
- **説明**: JWTトークンを無効化（ブラックリスト登録）
- **ヘッダー**: `Authorization: Bearer <token>`
- **リクエスト**: なし
- **レスポンス**:
```json
{
  "success": true,
  "message": "ログアウトしました"
}
```

### 1.3 トークン検証
- **エンドポイント**: `GET /api/admin/verify`
- **説明**: 現在のトークンの有効性を確認
- **ヘッダー**: `Authorization: Bearer <token>`
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "valid": true,
    "admin": {
      "id": 1,
      "username": "admin_user",
      "email": "admin@example.com"
    }
  }
}
```

## 2. 問題管理エンドポイント

### 2.1 問題一覧取得
- **エンドポイント**: `GET /api/admin/quizzes`
- **説明**: 全ての問題を取得
- **ヘッダー**: `Authorization: Bearer <token>`
- **クエリパラメータ**:
  - `page`: ページ番号（デフォルト: 1）
  - `limit`: 取得件数（デフォルト: 20）
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "quizzes": [
      {
        "id": 1,
        "question_text": "Go言語の開発元は？",
        "option_a": "Google",
        "option_b": "Microsoft",
        "option_c": "Apple",
        "option_d": "Meta",
        "correct_answer": "A",
        "image_url": "https://example.com/image1.jpg",
        "video_url": null,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 20
  }
}
```

### 2.2 問題詳細取得
- **エンドポイント**: `GET /api/admin/quizzes/{id}`
- **説明**: 指定されたIDの問題を取得
- **ヘッダー**: `Authorization: Bearer <token>`
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "question_text": "Go言語の開発元は？",
    "option_a": "Google",
    "option_b": "Microsoft",
    "option_c": "Apple",
    "option_d": "Meta",
    "correct_answer": "A",
    "image_url": "https://example.com/image1.jpg",
    "video_url": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 2.3 問題作成
- **エンドポイント**: `POST /api/admin/quizzes`
- **説明**: 新しい問題を作成
- **ヘッダー**: `Authorization: Bearer <token>`
- **リクエスト**:
```json
{
  "question_text": "Go言語の開発元は？",
  "option_a": "Google",
  "option_b": "Microsoft",
  "option_c": "Apple",
  "option_d": "Meta",
  "correct_answer": "A",
  "image_url": "https://example.com/image1.jpg",
  "video_url": null
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "問題が作成されました",
  "data": {
    "id": 1,
    "question_text": "Go言語の開発元は？",
    "option_a": "Google",
    "option_b": "Microsoft",
    "option_c": "Apple",
    "option_d": "Meta",
    "correct_answer": "A",
    "image_url": "https://example.com/image1.jpg",
    "video_url": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 2.4 問題更新
- **エンドポイント**: `PUT /api/admin/quizzes/{id}`
- **説明**: 指定されたIDの問題を更新
- **ヘッダー**: `Authorization: Bearer <token>`
- **リクエスト**:
```json
{
  "question_text": "Go言語の開発元は？（更新版）",
  "option_a": "Google",
  "option_b": "Microsoft",
  "option_c": "Apple",
  "option_d": "Meta",
  "correct_answer": "A",
  "image_url": "https://example.com/image1_updated.jpg",
  "video_url": null
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "問題が更新されました",
  "data": {
    "id": 1,
    "question_text": "Go言語の開発元は？（更新版）",
    "option_a": "Google",
    "option_b": "Microsoft",
    "option_c": "Apple",
    "option_d": "Meta",
    "correct_answer": "A",
    "image_url": "https://example.com/image1_updated.jpg",
    "video_url": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

### 2.5 問題削除
- **エンドポイント**: `DELETE /api/admin/quizzes/{id}`
- **説明**: 指定されたIDの問題を削除
- **ヘッダー**: `Authorization: Bearer <token>`
- **レスポンス**:
```json
{
  "success": true,
  "message": "問題が削除されました"
}
```

### 2.6 画像アップロード
- **エンドポイント**: `POST /api/admin/upload/image`
- **説明**: 問題用の画像をアップロード
- **ヘッダー**: `Authorization: Bearer <token>`
- **Content-Type**: `multipart/form-data`
- **リクエスト**: 
  - `file`: 画像ファイル（JPEG, PNG, GIF対応）
- **レスポンス**:
```json
{
  "success": true,
  "message": "画像がアップロードされました",
  "data": {
    "url": "https://example.com/uploads/images/quiz_image_123.jpg",
    "filename": "quiz_image_123.jpg",
    "size": 1024000
  }
}
```

## 3. セッション管理エンドポイント

### 3.1 現在のセッション状態取得
- **エンドポイント**: `GET /api/session/status`
- **説明**: 現在のクイズセッションの状態を取得
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "session_id": 1,
    "current_quiz": {
      "id": 5,
      "question_text": "Go言語でgoroutineを開始するキーワードは？",
      "option_a": "go",
      "option_b": "run",
      "option_c": "start",
      "option_d": "async",
      "image_url": null,
      "video_url": null
    },
    "is_accepting_answers": true,
    "total_participants": 150,
    "answers_count": 120
  }
}
```

### 3.2 クイズセッション開始
- **エンドポイント**: `POST /api/admin/session/start`
- **説明**: 新しいクイズセッションを開始
- **ヘッダー**: `Authorization: Bearer <token>`
- **リクエスト**:
```json
{
  "quiz_id": 1
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "クイズセッションが開始されました",
  "data": {
    "session_id": 1,
    "quiz": {
      "id": 1,
      "question_text": "Go言語の開発元は？",
      "option_a": "Google",
      "option_b": "Microsoft",
      "option_c": "Apple",
      "option_d": "Meta",
      "image_url": "https://example.com/image1.jpg",
      "video_url": null
    },
    "is_accepting_answers": true
  }
}
```

### 3.3 次の問題に進む
- **エンドポイント**: `POST /api/admin/session/next`
- **説明**: 次の問題に進む
- **ヘッダー**: `Authorization: Bearer <token>`
- **リクエスト**:
```json
{
  "quiz_id": 2
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "次の問題に進みました",
  "data": {
    "session_id": 1,
    "quiz": {
      "id": 2,
      "question_text": "Goのパッケージ管理ツールは？",
      "option_a": "npm",
      "option_b": "go mod",
      "option_c": "pip",
      "option_d": "composer",
      "image_url": null,
      "video_url": null
    },
    "is_accepting_answers": true
  }
}
```

### 3.4 回答受付開始/停止
- **エンドポイント**: `POST /api/admin/session/toggle-answers`
- **説明**: 現在の問題の回答受付を開始/停止
- **ヘッダー**: `Authorization: Bearer <token>`
- **リクエスト**:
```json
{
  "is_accepting_answers": false
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "回答受付を停止しました",
  "data": {
    "is_accepting_answers": false
  }
}
```

### 3.5 セッション終了
- **エンドポイント**: `POST /api/admin/session/end`
- **説明**: 現在のクイズセッションを終了
- **ヘッダー**: `Authorization: Bearer <token>`
- **レスポンス**:
```json
{
  "success": true,
  "message": "クイズセッションが終了されました"
}
```

## 4. 参加者登録エンドポイント

### 4.1 参加者登録
- **エンドポイント**: `POST /api/participants/register`
- **説明**: 参加者をニックネームで登録
- **リクエスト**:
```json
{
  "nickname": "GoファンA"
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "参加者として登録されました",
  "data": {
    "participant_id": 123,
    "nickname": "GoファンA",
    "created_at": "2024-01-01T10:00:00Z"
  }
}
```

### 4.2 参加者情報取得
- **エンドポイント**: `GET /api/participants/{id}`
- **説明**: 指定されたIDの参加者情報を取得
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "id": 123,
    "nickname": "GoファンA",
    "created_at": "2024-01-01T10:00:00Z",
    "total_answers": 5,
    "correct_answers": 3
  }
}
```

## 5. 回答送信・変更エンドポイント

### 5.1 回答送信
- **エンドポイント**: `POST /api/answers`
- **説明**: 現在の問題に対する回答を送信
- **リクエスト**:
```json
{
  "participant_id": 123,
  "quiz_id": 1,
  "selected_option": "A"
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "回答が送信されました",
  "data": {
    "answer_id": 456,
    "participant_id": 123,
    "quiz_id": 1,
    "selected_option": "A",
    "is_correct": true,
    "answered_at": "2024-01-01T10:05:00Z"
  }
}
```

### 5.2 回答変更
- **エンドポイント**: `PUT /api/answers/{id}`
- **説明**: 既存の回答を変更（回答受付中のみ可能）
- **リクエスト**:
```json
{
  "selected_option": "B"
}
```
- **レスポンス**:
```json
{
  "success": true,
  "message": "回答が変更されました",
  "data": {
    "answer_id": 456,
    "participant_id": 123,
    "quiz_id": 1,
    "selected_option": "B",
    "is_correct": false,
    "answered_at": "2024-01-01T10:07:00Z"
  }
}
```

### 5.3 参加者の回答履歴取得
- **エンドポイント**: `GET /api/participants/{id}/answers`
- **説明**: 指定された参加者の全回答履歴を取得
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "participant_id": 123,
    "answers": [
      {
        "answer_id": 456,
        "quiz_id": 1,
        "question_text": "Go言語の開発元は？",
        "selected_option": "B",
        "correct_answer": "A",
        "is_correct": false,
        "answered_at": "2024-01-01T10:07:00Z"
      }
    ],
    "total_answers": 5,
    "correct_answers": 3,
    "accuracy_rate": 0.6
  }
}
```

## 6. リアルタイム集計結果取得エンドポイント

### 6.1 現在の問題の集計結果
- **エンドポイント**: `GET /api/results/current`
- **説明**: 現在の問題の回答集計結果をリアルタイムで取得
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "quiz_id": 1,
    "question_text": "Go言語の開発元は？",
    "total_answers": 150,
    "results": {
      "A": {
        "count": 120,
        "percentage": 80.0
      },
      "B": {
        "count": 20,
        "percentage": 13.3
      },
      "C": {
        "count": 5,
        "percentage": 3.3
      },
      "D": {
        "count": 5,
        "percentage": 3.3
      }
    },
    "correct_answer": "A",
    "correct_count": 120,
    "correct_percentage": 80.0,
    "is_accepting_answers": false,
    "updated_at": "2024-01-01T10:10:00Z"
  }
}
```

### 6.2 指定問題の集計結果
- **エンドポイント**: `GET /api/results/quiz/{id}`
- **説明**: 指定された問題の集計結果を取得
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "quiz_id": 1,
    "question_text": "Go言語の開発元は？",
    "total_answers": 150,
    "results": {
      "A": {
        "count": 120,
        "percentage": 80.0
      },
      "B": {
        "count": 20,
        "percentage": 13.3
      },
      "C": {
        "count": 5,
        "percentage": 3.3
      },
      "D": {
        "count": 5,
        "percentage": 3.3
      }
    },
    "correct_answer": "A",
    "correct_count": 120,
    "correct_percentage": 80.0
  }
}
```

### 6.3 WebSocket接続（リアルタイム更新）
- **エンドポイント**: `WSS /api/ws/results`
- **説明**: WebSocketでリアルタイム集計結果を配信
- **接続時送信メッセージ**:
```json
{
  "type": "subscribe",
  "quiz_id": 1
}
```
- **受信メッセージ例**:
```json
{
  "type": "result_update",
  "data": {
    "quiz_id": 1,
    "total_answers": 151,
    "results": {
      "A": {
        "count": 121,
        "percentage": 80.1
      },
      "B": {
        "count": 20,
        "percentage": 13.2
      },
      "C": {
        "count": 5,
        "percentage": 3.3
      },
      "D": {
        "count": 5,
        "percentage": 3.3
      }
    },
    "updated_at": "2024-01-01T10:10:30Z"
  }
}
```

## 7. ランキング取得エンドポイント

### 7.1 総合ランキング
- **エンドポイント**: `GET /api/ranking/overall`
- **説明**: 全参加者の総合ランキングを取得
- **クエリパラメータ**:
  - `limit`: 取得件数（デフォルト: 100）
  - `offset`: 開始位置（デフォルト: 0）
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "ranking": [
      {
        "rank": 1,
        "participant_id": 123,
        "nickname": "GoマスターA",
        "total_answers": 10,
        "correct_answers": 10,
        "accuracy_rate": 1.0,
        "total_score": 100
      },
      {
        "rank": 2,
        "participant_id": 456,
        "nickname": "GoファンB",
        "total_answers": 10,
        "correct_answers": 9,
        "accuracy_rate": 0.9,
        "total_score": 90
      }
    ],
    "total_participants": 500,
    "updated_at": "2024-01-01T10:15:00Z"
  }
}
```

### 7.2 問題別正解率ランキング
- **エンドポイント**: `GET /api/ranking/quiz/{id}`
- **説明**: 指定された問題の正解者一覧
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "quiz_id": 1,
    "question_text": "Go言語の開発元は？",
    "correct_participants": [
      {
        "participant_id": 123,
        "nickname": "GoマスターA",
        "selected_option": "A",
        "answered_at": "2024-01-01T10:01:00Z"
      },
      {
        "participant_id": 789,
        "nickname": "GoファンC",
        "selected_option": "A",
        "answered_at": "2024-01-01T10:01:15Z"
      }
    ],
    "total_correct": 120,
    "total_answers": 150,
    "correct_percentage": 80.0
  }
}
```

### 7.3 参加者個人の順位
- **エンドポイント**: `GET /api/ranking/participant/{id}`
- **説明**: 指定された参加者の現在の順位と統計
- **レスポンス**:
```json
{
  "success": true,
  "data": {
    "participant_id": 123,
    "nickname": "GoファンA",
    "current_rank": 15,
    "total_participants": 500,
    "total_answers": 8,
    "correct_answers": 6,
    "accuracy_rate": 0.75,
    "total_score": 60,
    "percentile": 97.0
  }
}
```

## 8. エラーレスポンス

### 8.1 共通エラー形式
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "入力値が不正です",
    "details": [
      {
        "field": "nickname",
        "message": "ニックネームは必須です"
      }
    ]
  }
}
```

### 8.2 HTTPステータスコード
- `200`: 成功
- `201`: 作成成功
- `400`: リクエストエラー
- `401`: 認証エラー
- `403`: 権限エラー
- `404`: リソースが見つからない
- `409`: 競合エラー（重複など）
- `429`: レート制限
- `500`: サーバーエラー

## 9. 認証・セキュリティ

### 9.1 JWT認証
- 管理者用エンドポイントはJWT Bearer認証が必要
- トークンの有効期限: 24時間
- リフレッシュトークン機能なし（再ログインが必要）

### 9.2 レート制限
- 一般エンドポイント: 100リクエスト/分
- 管理者エンドポイント: 1000リクエスト/分
- 回答送信: 1リクエスト/秒

### 9.3 CORS設定
- 開発環境: `*`
- 本番環境: 指定されたドメインのみ

### 9.4 ファイルアップロード制限
- 画像ファイル: 最大5MB
- 対応形式: JPEG, PNG, GIF
- ウイルススキャン実行