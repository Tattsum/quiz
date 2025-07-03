describe('管理者フロー', () => {
  beforeEach(() => {
    // テスト前にデータをリセット
    cy.clearLocalStorage()
    cy.clearSessionStorage()
  })

  it('管理者ログインからクイズ作成までの基本フロー', () => {
    // 1. 管理者ログイン
    cy.visit('http://localhost:3001/login')
    cy.get('input[name="username"]').should('be.visible')
    cy.get('input[name="username"]').type('testadmin')
    cy.get('input[name="password"]').type('password')
    cy.get('button[type="submit"]').click()
    
    // ダッシュボードに遷移することを確認
    cy.url().should('include', '/dashboard')
    cy.contains('ダッシュボード').should('be.visible')

    // 2. クイズ作成ページに移動
    cy.visit('http://localhost:3001/questions/create')
    cy.contains('問題作成').should('be.visible')

    // 3. クイズを作成
    cy.get('input[name="question_text"]').type('E2Eテスト用問題: 日本の首都はどこですか？')
    cy.get('input[name="option_a"]').type('東京')
    cy.get('input[name="option_b"]').type('大阪')
    cy.get('input[name="option_c"]').type('京都')
    cy.get('input[name="option_d"]').type('名古屋')
    cy.get('select[name="correct_answer"]').select('A')
    
    // 4. クイズを保存
    cy.get('button[type="submit"]').click()
    
    // 成功メッセージまたはリダイレクトを確認
    cy.url().should('include', '/questions')
    cy.contains('E2Eテスト用問題').should('be.visible')
  })

  it('クイズ管理機能のテスト', () => {
    // 事前に管理者ログイン
    cy.adminLogin()

    // クイズ一覧ページに移動
    cy.visit('http://localhost:3001/questions')
    cy.contains('問題管理').should('be.visible')

    // 既存のクイズがあることを確認
    cy.get('[data-testid="quiz-item"]').should('have.length.at.least', 1)

    // クイズ編集テスト
    cy.get('[data-testid="edit-quiz-button"]').first().click()
    cy.url().should('include', '/questions/')
    cy.get('input[name="question_text"]').should('have.value')
    
    // 問題文を変更
    cy.get('input[name="question_text"]').clear().type('編集されたテスト問題')
    cy.get('button[type="submit"]').click()
    
    // 変更が反映されることを確認
    cy.visit('http://localhost:3001/questions')
    cy.contains('編集されたテスト問題').should('be.visible')
  })

  it('セッション管理機能のテスト', () => {
    cy.adminLogin()

    // セッション制御ページに移動
    cy.visit('http://localhost:3001/quiz-control')
    cy.contains('クイズ制御').should('be.visible')

    // セッション開始
    cy.get('[data-testid="start-session-button"]').click()
    cy.contains('セッション開始').should('be.visible')

    // 問題選択
    cy.get('[data-testid="quiz-select"]').select(0) // 最初のクイズを選択
    cy.get('[data-testid="start-quiz-button"]').click()

    // セッション状態の確認
    cy.contains('進行中').should('be.visible')
    cy.get('[data-testid="session-status"]').should('contain', '進行中')

    // 回答受付切り替えテスト
    cy.get('[data-testid="toggle-answers-button"]').click()
    cy.contains('回答受付停止').should('be.visible')

    cy.get('[data-testid="toggle-answers-button"]').click()
    cy.contains('回答受付開始').should('be.visible')
  })

  it('リアルタイム統計表示のテスト', () => {
    cy.adminLogin()

    // リアルタイム統計ページに移動
    cy.visit('http://localhost:3001/quiz-control/realtime')
    cy.contains('リアルタイム統計').should('be.visible')

    // チャートコンポーネントが表示されることを確認
    cy.get('[data-testid="realtime-chart"]').should('be.visible')
    
    // WebSocket接続状態インジケーターを確認
    cy.get('[data-testid="ws-status"]').should('be.visible')
    
    // 統計情報の表示要素を確認
    cy.contains('総参加者').should('be.visible')
    cy.contains('回答済み').should('be.visible')
    cy.contains('回答率').should('be.visible')
  })

  it('ランキング表示機能のテスト', () => {
    cy.adminLogin()

    // ランキングページに移動
    cy.visit('http://localhost:3001/ranking')
    cy.contains('ランキング').should('be.visible')

    // ランキングテーブルが表示されることを確認
    cy.get('[data-testid="ranking-table"]').should('be.visible')
    
    // ランキング項目を確認
    cy.contains('順位').should('be.visible')
    cy.contains('ニックネーム').should('be.visible')
    cy.contains('正答率').should('be.visible')
  })

  it('プロジェクター画面への遷移テスト', () => {
    cy.adminLogin()

    // ダッシュボードからプロジェクター画面へのリンクを確認
    cy.visit('http://localhost:3001/dashboard')
    cy.get('[data-testid="projector-link"]').should('be.visible')
    
    // プロジェクター画面に移動（新しいタブで開く）
    cy.get('[data-testid="projector-link"]').invoke('removeAttr', 'target').click()
    
    // プロジェクター画面の要素を確認
    cy.url().should('include', '/projector')
    cy.contains('問題').should('be.visible')
    cy.get('[data-testid="countdown-timer"]').should('be.visible')
  })
})

describe('管理者認証テスト', () => {
  it('無効な認証情報でのログイン失敗', () => {
    cy.visit('http://localhost:3001/login')
    
    cy.get('input[name="username"]').type('invaliduser')
    cy.get('input[name="password"]').type('wrongpassword')
    cy.get('button[type="submit"]').click()
    
    // エラーメッセージが表示されることを確認
    cy.contains('ログインに失敗しました').should('be.visible')
    cy.url().should('include', '/login')
  })

  it('未認証でのダッシュボードアクセス防止', () => {
    // 認証なしでダッシュボードにアクセス
    cy.visit('http://localhost:3001/dashboard')
    
    // ログインページにリダイレクトされることを確認
    cy.url().should('include', '/login')
  })

  it('セッション期限切れ後のリダイレクト', () => {
    // 管理者ログイン
    cy.adminLogin()
    
    // ローカルストレージのトークンを無効化
    cy.window().then((win) => {
      win.localStorage.setItem('admin_token', 'invalid_token')
    })
    
    // 保護されたページにアクセス
    cy.visit('http://localhost:3001/questions')
    
    // ログインページにリダイレクトされることを確認
    cy.url().should('include', '/login')
  })
})