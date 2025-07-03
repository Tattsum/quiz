describe('参加者フロー', () => {
  beforeEach(() => {
    cy.clearLocalStorage()
    cy.clearSessionStorage()
  })

  it('参加者登録から回答送信までの基本フロー', () => {
    // 1. 参加者アプリにアクセス
    cy.visit('http://localhost:3002')
    cy.contains('クイズに参加').should('be.visible')

    // 2. ニックネーム入力
    cy.get('input[id="nickname"]').should('be.visible')
    cy.get('input[id="nickname"]').type('E2Eテストユーザー')
    
    // 3. 参加ボタンをクリック
    cy.get('button[type="submit"]').should('not.be.disabled')
    cy.get('button[type="submit"]').click()

    // 4. 待機画面に遷移
    cy.contains('クイズの開始をお待ちください').should('be.visible')
    cy.get('[data-testid="participant-id"]').should('be.visible')

    // 5. 管理者がセッションを開始したと仮定して、回答画面を確認
    // （実際のE2Eテストでは管理者操作も含める）
    cy.window().then((win) => {
      // WebSocket経由で問題データを受信したとシミュレート
      const mockQuestionData = {
        type: 'question_switch',
        quiz_id: 1,
        question_number: 1,
        total_questions: 5,
        question: {
          id: 1,
          question_text: 'テスト問題: 日本の首都はどこですか？',
          option_a: '東京',
          option_b: '大阪', 
          option_c: '京都',
          option_d: '名古屋'
        }
      }
      
      // カスタムイベントを発火して問題表示をトリガー
      win.dispatchEvent(new CustomEvent('mock-question-received', { 
        detail: mockQuestionData 
      }))
    })

    // 6. 回答画面の要素を確認
    cy.contains('問題 1 / 5').should('be.visible')
    cy.contains('日本の首都はどこですか？').should('be.visible')
    
    // 選択肢ボタンを確認
    cy.get('[data-testid="option-a"]').should('contain', '東京')
    cy.get('[data-testid="option-b"]').should('contain', '大阪')
    cy.get('[data-testid="option-c"]').should('contain', '京都')
    cy.get('[data-testid="option-d"]').should('contain', '名古屋')

    // 7. 回答を選択
    cy.get('[data-testid="option-a"]').click()
    cy.get('[data-testid="option-a"]').should('have.class', 'selected')

    // 8. 回答送信
    cy.get('[data-testid="submit-answer"]').click()
    cy.contains('回答を送信しました').should('be.visible')
  })

  it('ニックネーム入力のバリデーションテスト', () => {
    cy.visit('http://localhost:3002')

    // 空のニックネームでの送信
    cy.get('button[type="submit"]').should('be.disabled')

    // 短すぎるニックネーム
    cy.get('input[id="nickname"]').type('あ')
    cy.get('button[type="submit"]').click()
    cy.contains('ニックネームは2文字以上で入力してください').should('be.visible')

    // 長すぎるニックネーム
    cy.get('input[id="nickname"]').clear()
    cy.get('input[id="nickname"]').type('あ'.repeat(21))
    cy.get('button[type="submit"]').click()
    cy.contains('ニックネームは20文字以下で入力してください').should('be.visible')

    // 有効なニックネーム
    cy.get('input[id="nickname"]').clear()
    cy.get('input[id="nickname"]').type('有効なニックネーム')
    cy.get('button[type="submit"]').should('not.be.disabled')
  })

  it('回答変更機能のテスト', () => {
    // 参加者登録
    cy.registerParticipant('回答変更テストユーザー')

    // 模擬的に問題表示状態にする
    cy.window().then((win) => {
      const mockQuestionData = {
        type: 'question_switch',
        quiz_id: 1,
        question: {
          id: 1,
          question_text: 'テスト問題',
          option_a: '選択肢A',
          option_b: '選択肢B',
          option_c: '選択肢C',
          option_d: '選択肢D'
        }
      }
      win.dispatchEvent(new CustomEvent('mock-question-received', { 
        detail: mockQuestionData 
      }))
    })

    // 最初の回答を選択
    cy.get('[data-testid="option-a"]').click()
    cy.get('[data-testid="option-a"]').should('have.class', 'selected')

    // 別の回答に変更
    cy.get('[data-testid="option-b"]').click()
    cy.get('[data-testid="option-b"]').should('have.class', 'selected')
    cy.get('[data-testid="option-a"]').should('not.have.class', 'selected')

    // 回答送信
    cy.get('[data-testid="submit-answer"]').click()
  })

  it('回答結果表示のテスト', () => {
    cy.registerParticipant('結果表示テストユーザー')

    // 問題に回答後、結果表示をシミュレート
    cy.window().then((win) => {
      const mockResultData = {
        type: 'voting_end',
        quiz_id: 1,
        results: {
          correct_answer: 'A',
          your_answer: 'A',
          is_correct: true,
          total_answers: 50,
          correct_count: 35,
          answer_distribution: {
            'A': 35,
            'B': 10,
            'C': 3,
            'D': 2
          }
        }
      }
      win.dispatchEvent(new CustomEvent('mock-result-received', { 
        detail: mockResultData 
      }))
    })

    // 結果画面の要素を確認
    cy.contains('結果発表').should('be.visible')
    cy.contains('正解').should('be.visible')
    cy.get('[data-testid="correct-answer"]').should('contain', 'A')
    cy.get('[data-testid="your-answer"]').should('contain', 'A')
    cy.contains('正答率').should('be.visible')
  })

  it('待機画面のテスト', () => {
    cy.registerParticipant('待機画面テストユーザー')

    // 待機画面の要素を確認
    cy.contains('クイズの開始をお待ちください').should('be.visible')
    cy.get('[data-testid="participant-count"]').should('be.visible')
    cy.get('[data-testid="waiting-animation"]').should('be.visible')

    // 参加者IDが表示されることを確認
    cy.get('[data-testid="participant-id"]').should('contain', 'ID:')
  })

  it('最終結果画面のテスト', () => {
    cy.registerParticipant('最終結果テストユーザー')

    // 最終結果をシミュレート
    cy.window().then((win) => {
      const mockFinalResult = {
        type: 'quiz_completed',
        final_results: {
          total_questions: 5,
          correct_answers: 4,
          accuracy_rate: 80,
          rank: 5,
          total_participants: 20
        }
      }
      win.dispatchEvent(new CustomEvent('mock-final-result', { 
        detail: mockFinalResult 
      }))
    })

    // 最終結果画面の要素を確認
    cy.contains('クイズ終了').should('be.visible')
    cy.contains('お疲れ様でした').should('be.visible')
    cy.get('[data-testid="total-score"]').should('contain', '4/5')
    cy.get('[data-testid="accuracy-rate"]').should('contain', '80%')
    cy.get('[data-testid="final-rank"]').should('contain', '5位')
  })

  it('レスポンシブデザインのテスト', () => {
    // モバイルビューポートでのテスト
    cy.setMobileViewport()
    cy.visit('http://localhost:3002')
    
    // モバイルでの表示を確認
    cy.get('input[id="nickname"]').should('be.visible')
    cy.get('button[type="submit"]').should('be.visible')
    
    // ボタンサイズがモバイル用に適切かチェック
    cy.get('button[type="submit"]').should('have.css', 'padding')

    // タブレットビューポートでのテスト
    cy.setTabletViewport()
    cy.reload()
    cy.get('input[id="nickname"]').should('be.visible')

    // デスクトップビューポートでのテスト
    cy.setDesktopViewport()
    cy.reload()
    cy.get('input[id="nickname"]').should('be.visible')
  })
})

describe('参加者のエラーハンドリングテスト', () => {
  it('ネットワークエラー時の表示', () => {
    // ネットワークエラーをシミュレート
    cy.intercept('POST', '**/api/public/participants', {
      statusCode: 500,
      body: { error: 'Internal Server Error' }
    }).as('registerError')

    cy.visit('http://localhost:3002')
    cy.get('input[id="nickname"]').type('エラーテストユーザー')
    cy.get('button[type="submit"]').click()

    cy.wait('@registerError')
    cy.contains('登録に失敗しました').should('be.visible')
  })

  it('WebSocket接続エラー時の表示', () => {
    cy.registerParticipant('WebSocketエラーテストユーザー')
    
    // WebSocket接続エラーをシミュレート
    cy.window().then((win) => {
      // WebSocket接続をモック（エラー状態）
      const mockWsError = new CustomEvent('websocket-error', {
        detail: { error: 'Connection failed' }
      })
      win.dispatchEvent(mockWsError)
    })

    // エラー表示を確認
    cy.contains('接続に問題が発生しました').should('be.visible')
    cy.get('[data-testid="retry-button"]').should('be.visible')
  })

  it('回答送信エラー時の再試行', () => {
    cy.registerParticipant('回答エラーテストユーザー')

    // 回答送信エラーをシミュレート
    cy.intercept('POST', '**/api/public/answers', {
      statusCode: 500,
      body: { error: 'Server Error' }
    }).as('answerError')

    // 模擬問題表示
    cy.window().then((win) => {
      const mockQuestionData = {
        type: 'question_switch',
        quiz_id: 1,
        question: {
          id: 1,
          question_text: 'エラーテスト問題',
          option_a: 'A', option_b: 'B', option_c: 'C', option_d: 'D'
        }
      }
      win.dispatchEvent(new CustomEvent('mock-question-received', { 
        detail: mockQuestionData 
      }))
    })

    cy.get('[data-testid="option-a"]').click()
    cy.get('[data-testid="submit-answer"]').click()

    cy.wait('@answerError')
    cy.contains('回答の送信に失敗しました').should('be.visible')
    cy.get('[data-testid="retry-submit"]').should('be.visible')
  })
})