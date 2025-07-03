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

## コード品質・リント・フォーマットコマンド

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

#### セキュリティとコード品質ガイドライン
- **ファイルパス検証**: `#nosec G304`コメントで適切にセキュリティリスクを文書化
- **未使用パラメータ**: `_ /*paramName*/`記法で明示的に無効化
- **空ブロック**: defer文では`_ = resource.Close()`で簡潔に記述
- **エクスポート関数**: 全てのpublic関数・変数に適切なコメントを追加
- **型アサーション**: チェック付きアサーション`value, ok := interface{}.(Type)`を使用
- **定数使用**: ハードコーディングを避け、定数を積極的に活用

### Nuxt3管理ダッシュボード
```
cd admin-dashboard
npm run lint       # ESLintチェック
npm run test       # Vitestテスト実行
npm run dev        # 開発サーバー起動
npm run build      # プロダクションビルド
```

#### テスト品質
- Chart.jsなど外部ライブラリの適切なモック
- Vue Test Utilsを使用したコンポーネントテスト
- WebSocketの統合テスト

### Next.js参加者アプリ
```
cd participant-app
npm run lint       # Next.jsリント
npm run test       # Jestテスト実行
npm run dev        # 開発サーバー起動
npm run build      # プロダクションビルド
```

#### 品質指標
- **Goバックエンド**: 主要サービス85%以上のテストカバレッジ
- **Nuxt3**: 全14テスト成功、Chart.js統合対応
- **Next.js**: ESLint 100%成功、React Hook適切な使用

## Git運用ルール

- 作業毎に必ずcommitする（まとめてcommitしない）
- 各タスク完了時に個別にcommitを実行する
- すべての作業完了後にプルリクエスト（PR）を作成する

## ドキュメント更新ルール

- **作業完了時の必須タスク**: 新機能追加、プロジェクト構成変更、重要な修正作業の完了後は必ず以下のドキュメントを確認・更新する
  - `CLAUDE.md`: 開発ルール、プロジェクト構成の変更を反映
  - `README.md`: メインプロジェクトの説明、セットアップ手順の更新
  - 各サブプロジェクトの`README.md`: 機能追加や変更内容を反映
- **更新対象となる作業例**:
  - 新しいプロジェクト・モジュールの追加
  - 技術スタックの変更
  - セットアップ手順の変更
  - 新機能の追加
  - アーキテクチャの変更

## tenntennさんの思想に基づくGoアーキテクチャ指針

### 基本原則

#### 1. シンプルさの追求 (Simplicity First)
- **Go言語の哲学**: "Less is more" - 複雑さを避け、誰でも理解できるコードを書く
- **最小限の抽象化**: 過度な抽象化を避け、必要な時にのみ抽象化を行う
- **明示的な依存関係**: 依存関係を隠さず、明確に表現する
- **標準ライブラリ優先**: サードパーティライブラリよりも標準ライブラリを積極的に活用

#### 2. テスタビリティの重視 (Testability Driven Design)
- **テストファースト**: 実装前にテストを書き、テスト可能な設計を心がける
- **依存関係の注入**: テストしやすいようにインターフェースと依存関係注入を活用
- **モックの適切な使用**: 外部依存をモック化し、単体テストの独立性を保つ
- **テーブル駆動テスト**: Go言語の慣習に従い、テーブル駆動テストを積極的に使用

#### 3. 明確性の追求 (Clarity Over Cleverness)
- **可読性優先**: 巧妙なコードよりも、読みやすく理解しやすいコードを優先
- **明示的エラーハンドリング**: 例外ではなく、明示的なエラー値を返してエラーを処理
- **名前付けの重要性**: 変数名、関数名、パッケージ名で意図を明確に表現
- **コメントは"なぜ"を説明**: コードが"何を"するかではなく、"なぜ"そうするかを説明

### アーキテクチャパターン

#### 1. Clean Architecture の適用
```go
// Domain Layer (ビジネスロジック)
type QuizService interface {
    CreateQuiz(ctx context.Context, quiz *Quiz) error
    GetQuiz(ctx context.Context, id int) (*Quiz, error)
}

// Infrastructure Layer (技術的詳細)
type QuizRepository interface {
    Save(ctx context.Context, quiz *Quiz) error
    FindByID(ctx context.Context, id int) (*Quiz, error)
}

// Application Layer (ユースケース)
type QuizUsecase struct {
    repo QuizRepository
    // 依存関係は明示的に注入
}
```

#### 2. インターフェース設計の原則
- **小さなインターフェース**: 1-3個のメソッドを持つ小さなインターフェースを作成
- **受け取る側が定義**: インターフェースは使用する側のパッケージで定義する
- **具象型で返す**: 関数は具象型を返し、インターフェースを受け取る

```go
// Good: 小さく、焦点の明確なインターフェース
type Reader interface {
    Read([]byte) (int, error)
}

// Good: 使用する側で定義
type UserService struct {
    repo UserRepository  // このパッケージで定義
}

type UserRepository interface {
    FindByID(int) (*User, error)
}
```

#### 3. エラーハンドリング戦略
- **エラーラッピング**: `fmt.Errorf`と`%w`動詞でエラーをラップ
- **カスタムエラー型**: ドメイン固有のエラーは独自の型を定義
- **エラーの境界**: レイヤー間でエラーを変換し、内部実装を隠蔽

```go
// ドメインエラーの定義
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed on field %s: %s", e.Field, e.Message)
}

// エラーラッピング
func (s *QuizService) CreateQuiz(ctx context.Context, quiz *Quiz) error {
    if err := quiz.Validate(); err != nil {
        return fmt.Errorf("quiz validation failed: %w", err)
    }
    
    if err := s.repo.Save(ctx, quiz); err != nil {
        return fmt.Errorf("failed to save quiz: %w", err)
    }
    
    return nil
}
```

#### 4. 並行処理のベストプラクティス
- **Goroutineの責任範囲**: 起動した側が責任を持って管理する
- **Context使用**: タイムアウトとキャンセレーションにcontextを活用
- **チャネルによる通信**: 共有メモリよりもチャネルによる通信を優先

```go
// WebSocket接続管理（本プロジェクトでの実装例）
func (h *WebSocketHandler) handleConnection(ctx context.Context, conn *websocket.Conn) {
    defer conn.Close()
    
    // Goroutineでハートビート監視
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    go h.heartbeat(ctx, conn)
    
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // メッセージ処理
        }
    }
}
```

### プロジェクト固有の実装ガイドライン

#### 1. ディレクトリ構造の原則
```
internal/
├── domain/          # ビジネスロジック（エンティティ、値オブジェクト）
├── usecase/         # ユースケース（アプリケーションロジック）
├── repository/      # データアクセスインターフェース
├── handler/         # HTTPハンドラー（インターフェースアダプター）
├── infrastructure/  # 技術的詳細（DB、外部API）
└── pkg/            # 共通パッケージ
```

#### 2. 依存関係の方向
- **外側→内側**: 外側のレイヤーは内側のレイヤーに依存
- **内側は外側を知らない**: ドメイン層はインフラストラクチャ層を知らない
- **インターフェースで境界を作る**: レイヤー間はインターフェースで通信

#### 3. テスト戦略
- **単体テスト**: 各レイヤーを独立してテスト
- **統合テスト**: レイヤー間の連携をテスト
- **モックの活用**: 外部依存をモック化してテストの安定性を確保

```go
// テスト例：ユースケースのテスト
func TestQuizUsecase_CreateQuiz(t *testing.T) {
    tests := []struct {
        name    string
        quiz    *Quiz
        repoErr error
        wantErr bool
    }{
        {
            name: "success",
            quiz: &Quiz{Title: "Test Quiz"},
            repoErr: nil,
            wantErr: false,
        },
        {
            name: "repository error",
            quiz: &Quiz{Title: "Test Quiz"},
            repoErr: errors.New("db error"),
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockQuizRepository{
                SaveFunc: func(ctx context.Context, quiz *Quiz) error {
                    return tt.repoErr
                },
            }
            
            usecase := &QuizUsecase{repo: mockRepo}
            err := usecase.CreateQuiz(context.Background(), tt.quiz)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateQuiz() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 品質維持のための継続的改善

#### 1. 静的解析の活用
- **golangci-lint**: 包括的な静的解析
- **go vet**: 潜在的な問題の検出
- **go fmt/gofumpt**: 一貫したフォーマット

#### 2. パフォーマンス監視
- **pprof**: プロファイリングツールの活用
- **ベンチマーク**: 性能向上のためのベンチマークテスト
- **メトリクス**: 実行時メトリクスの監視

#### 3. レビュー文化
- **コードレビュー**: 設計思想の共有と品質向上
- **ペアプログラミング**: 知識共有と品質向上
- **リファクタリング**: 継続的な改善

## 個人開発でのGoアーキテクチャ指針

### 個人開発における現実的なアプローチ

#### 1. 段階的成長戦略 (Progressive Growth)
- **MVP (Minimum Viable Product) 重視**: 最小限の機能で早期リリース
- **段階的複雑化**: 必要に応じて機能とアーキテクチャを拡張
- **一人でメンテナンスできる規模**: 過度な抽象化を避ける
- **将来の拡張性**: チーム開発への移行を考慮した構造

#### 2. コスト効率的な技術選択
- **標準ライブラリ最大限活用**: サードパーティライブラリを最小限に
- **無料サービスの活用**: GitHub Actions、無料監視ツール等の積極利用
- **VPS vs PaaS**: 学習・コスト面からVPSを選択
- **モノリス vs マイクロサービス**: 個人開発ではモノリスアーキテクチャを推奨

#### 3. 個人開発向けディレクトリ構成
```
quiz/
├── cmd/
│   └── server/         # アプリケーション エントリーポイント
├── internal/
│   ├── handler/        # HTTPハンドラー（シンプルな構成）
│   ├── service/        # ビジネスロジック（軽量）
│   ├── repository/     # データアクセス層
│   ├── model/          # ドメインモデル
│   └── middleware/     # 認証、ログ等の横断関心事
├── web/               # フロントエンド（Nuxt3、Next.js）
├── scripts/           # デプロイ、DB migrate等のスクリプト
├── docker/            # Docker関連ファイル
├── docs/              # 最小限のドキュメント
└── README.md          # セットアップ・運用手順
```

#### 4. 個人開発でのテスト戦略
- **30-40%カバレッジ目標**: 重要部分に集中したテスト
- **統合テストの重視**: 個人開発では統合テストが効率的
- **手動テストとの併用**: 自動化しきれない部分は手動で補完
- **CI/CDでの自動実行**: GitHub Actionsでのテスト自動化
```go
// 個人開発向けテストの例（実用性重視）
func TestQuizAPI_Integration(t *testing.T) {
    // データベースを含む統合テスト
    testDB := setupTestDB(t)
    defer testDB.Close()
    
    server := setupTestServer(testDB)
    defer server.Close()
    
    // 実際のAPIエンドポイントをテスト
    resp, err := http.Post(server.URL+"/api/quiz", "application/json", 
        strings.NewReader(`{"title":"Test Quiz","questions":[...]}`))
    
    assert.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
```

#### 5. 個人開発でのパフォーマンス考慮
- **VPS環境での最適化**: 限られたリソースでの効率的な動作
- **適切なDB設計**: インデックス設定、クエリ最適化
- **Redis活用**: セッション管理、簡単なキャッシュ
- **静的ファイル最適化**: Nginx でのgzip、キャッシュ設定

#### 6. 運用の簡素化
- **ログ管理**: Docker logs + logrotate（シンプルな構成）
- **監視**: UptimeRobot（無料）+ Grafana Cloud（無料枠）
- **バックアップ**: VPSスナップショット + 日次DBダンプ
- **デプロイ自動化**: GitHub Actions による完全自動化

#### 7. セキュリティ基本対策（個人開発）
```go
// 基本的なセキュリティミドルウェア
func securityHeaders() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Next()
    })
}

// JWTによる認証（シンプルな実装）
func authMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        // JWT検証処理...
        c.Next()
    }
}
```

#### 8. 技術的負債管理（個人開発版）
- **リファクタリングのタイミング**: 新機能追加前の適切なタイミング
- **コメント最小限**: 自己文書化されたコードを心がける
- **依存関係の見直し**: 四半期に1回のセキュリティアップデート
- **将来の自分への配慮**: 半年後でも理解できるシンプルな設計

#### 9. 個人開発でのGoベストプラクティス
```go
// シンプルなエラーハンドリング
func (s *QuizService) CreateQuiz(quiz *Quiz) error {
    if err := quiz.Validate(); err != nil {
        return fmt.Errorf("無効なクイズデータ: %w", err)
    }
    
    return s.repo.Save(quiz)
}

// 設定の環境変数管理
type Config struct {
    Port        string `env:"PORT" envDefault:"8080"`
    DatabaseURL string `env:"DATABASE_URL,required"`
    JWTSecret   string `env:"JWT_SECRET,required"`
    RedisURL    string `env:"REDIS_URL" envDefault:"localhost:6379"`
}

// グレースフルシャットダウン（個人開発でも重要）
func (s *Server) Start() error {
    srv := &http.Server{
        Addr:    ":" + s.config.Port,
        Handler: s.router,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("サーバー起動エラー: %v", err)
        }
    }()
    
    // グレースフルシャットダウンの処理
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return srv.Shutdown(ctx)
}
```

#### 10. スケーリング考慮事項（個人開発）
- **垂直スケーリング優先**: まずはVPSのスペックアップを検討
- **水平スケーリングの準備**: 将来的な複数インスタンス対応を考慮
- **ステートレス設計**: セッション情報はRedisに保存
- **データベース最適化**: 適切なインデックス設定とクエリ改善

このアーキテクチャ指針により、個人開発者でも **tenntennさんの思想** に基づいた **シンプルで保守性の高いGo言語アプリケーション** を構築・運用できます。
