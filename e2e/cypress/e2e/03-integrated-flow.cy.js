describe('統合フロー', () => {
  beforeEach(() => {
    cy.clearLocalStorage()
    cy.clearSessionStorage()
  })

  it('管理者と参加者の統合フロー', () => {
    // 1. 管理者操作: ログインとクイズ作成
    cy.adminLogin()
    
    // クイズを作成
    cy.createQuiz({
      question_text: '統合テスト問題: 富士山の標高は何メートルですか？',
      option_a: '3776m',
      option_b: '3000m', 
      option_c: '4000m',
      option_d: '3500m',
      correct_answer: 'A'
    })

    // 作成されたクイズのIDを取得
    cy.url().then((url) => {
      const quizId = url.split('/').pop()
      cy.wrap(quizId).as('quizId')
    })

    // 2. セッション開始
    cy.visit('http://localhost:3001/quiz-control')
    cy.get('@quizId').then((quizId) => {
      cy.get('[data-testid="quiz-select"]').select(quizId.toString())
    })
    cy.get('[data-testid="start-session-button"]').click()
    cy.contains('セッション開始').should('be.visible')

    // 3. 別ウィンドウで参加者操作
    cy.window().then((adminWin) => {
      // 参加者ウィンドウを開く
      const participantWin = adminWin.open('http://localhost:3002', 'participant')
      
      // 参加者の操作をシミュレート
      cy.wrap(participantWin).then((pWin) => {
        // 参加者のニックネーム登録
        const nicknameInput = pWin.document.querySelector('#nickname')
        const submitButton = pWin.document.querySelector('button[type="submit"]')
        
        if (nicknameInput && submitButton) {
          nicknameInput.value = '統合テスト参加者'
          submitButton.click()
        }
      })
    })

    // 4. 管理者側でセッション状況を確認
    cy.visit('http://localhost:3001/quiz-control/realtime')
    cy.contains('総参加者').should('be.visible')
    cy.get('[data-testid="participant-count"]').should('contain', '1')

    // 5. 問題を開始
    cy.visit('http://localhost:3001/quiz-control')
    cy.get('@quizId').then((quizId) => {
      cy.get('[data-testid="start-quiz-button"]').click()
    })

    // 6. リアルタイム統計で回答状況を確認
    cy.visit('http://localhost:3001/quiz-control/realtime')
    cy.waitForWebSocket()
    
    // WebSocket経由でデータが更新されることを確認
    cy.get('[data-testid="ws-status"]').should('contain', '接続中')
    cy.get('[data-testid="realtime-chart"]').should('be.visible')

    // 7. 結果確認
    cy.visit('http://localhost:3001/ranking')
    cy.get('[data-testid="ranking-table"]').should('be.visible')
    cy.contains('統合テスト参加者').should('be.visible')
  })

  it('複数参加者同時接続テスト', () => {
    // 管理者でセッション準備
    cy.adminLogin()
    cy.visit('http://localhost:3001/quiz-control')
    
    // セッション開始
    cy.get('[data-testid="start-session-button"]').click()

    // 複数の参加者ウィンドウを開く
    const participantCount = 5
    const participantNicknames = []

    for (let i = 0; i < participantCount; i++) {
      const nickname = `参加者${i + 1}`
      participantNicknames.push(nickname)
      
      cy.window().then((win) => {
        const participantWin = win.open('http://localhost:3002', `participant-${i}`)
        
        // 各参加者の登録をシミュレート
        cy.wrap(participantWin).then((pWin) => {
          // 参加者登録の処理
          setTimeout(() => {
            const nicknameInput = pWin.document.querySelector('#nickname')
            const submitButton = pWin.document.querySelector('button[type="submit"]')
            
            if (nicknameInput && submitButton) {
              nicknameInput.value = nickname
              submitButton.click()
            }
          }, i * 1000) // 時間差で登録
        })
      })
    }

    // 管理者側で参加者数を確認
    cy.visit('http://localhost:3001/quiz-control/realtime')
    cy.wait(6000) // 全参加者の登録を待つ
    
    cy.get('[data-testid="participant-count"]').should('contain', participantCount.toString())
    
    // 各参加者が表示されることを確認
    participantNicknames.forEach(nickname => {
      cy.contains(nickname).should('be.visible')
    })
  })

  it('WebSocketリアルタイム通信テスト', () => {
    cy.adminLogin()

    // リアルタイム統計ページを開く
    cy.visit('http://localhost:3001/quiz-control/realtime')
    
    // WebSocket接続を確認
    cy.waitForWebSocket()
    cy.get('[data-testid="ws-status"]').should('contain', '接続中')

    // 別ウィンドウで参加者を登録
    cy.window().then((win) => {
      const participantWin = win.open('http://localhost:3002', 'participant')
      
      cy.wrap(participantWin).then((pWin) => {
        setTimeout(() => {
          const nicknameInput = pWin.document.querySelector('#nickname')
          const submitButton = pWin.document.querySelector('button[type="submit"]')
          
          if (nicknameInput && submitButton) {
            nicknameInput.value = 'WebSocketテスト参加者'
            submitButton.click()
          }
        }, 1000)
      })
    })

    // リアルタイム更新を確認
    cy.wait(3000)
    cy.get('[data-testid="participant-count"]').should('not.contain', '0')

    // クイズ開始とリアルタイム回答状況の確認
    cy.visit('http://localhost:3001/quiz-control')
    cy.get('[data-testid="quiz-select"]').select(0)
    cy.get('[data-testid="start-quiz-button"]').click()

    // リアルタイム統計に戻って回答状況を確認
    cy.visit('http://localhost:3001/quiz-control/realtime')
    cy.get('[data-testid="answer-chart"]').should('be.visible')
    cy.contains('回答済み').should('be.visible')
  })

  it('プロジェクター表示との連携テスト', () => {
    cy.adminLogin()

    // プロジェクター画面を開く
    cy.visit('http://localhost:3001/projector')
    cy.contains('クイズ大会').should('be.visible')
    cy.get('[data-testid="countdown-timer"]').should('be.visible')

    // 管理者ウィンドウでセッション操作
    cy.window().then((projectorWin) => {
      const adminWin = projectorWin.open('http://localhost:3001/quiz-control', 'admin')
      
      cy.wrap(adminWin).then((aWin) => {
        // 管理者操作をシミュレート
        setTimeout(() => {
          const startButton = aWin.document.querySelector('[data-testid="start-session-button"]')
          if (startButton) {
            startButton.click()
          }
        }, 1000)
      })
    })

    // プロジェクター画面でセッション開始が反映されることを確認
    cy.wait(3000)
    cy.contains('進行中').should('be.visible')

    // 問題表示の確認
    cy.get('[data-testid="question-display"]').should('be.visible')
    cy.get('[data-testid="option-buttons"]').should('be.visible')
  })

  it('エラー処理と回復機能のテスト', () => {
    cy.adminLogin()

    // ネットワークエラーをシミュレート
    cy.intercept('GET', '**/api/admin/session/status', {
      statusCode: 500,
      body: { error: 'Server Error' }
    }).as('sessionError')

    cy.visit('http://localhost:3001/quiz-control')
    cy.wait('@sessionError')

    // エラー表示の確認
    cy.contains('データの取得に失敗しました').should('be.visible')
    cy.get('[data-testid="retry-button"]').should('be.visible')

    // 回復後の動作確認
    cy.intercept('GET', '**/api/admin/session/status', {
      statusCode: 200,
      body: {
        session_id: 1,
        is_accepting_answers: false,
        total_participants: 0
      }
    }).as('sessionSuccess')

    cy.get('[data-testid="retry-button"]').click()
    cy.wait('@sessionSuccess')
    cy.contains('クイズ制御').should('be.visible')
  })

  it('セッション終了フロー', () => {
    cy.adminLogin()

    // セッション開始
    cy.visit('http://localhost:3001/quiz-control')
    cy.get('[data-testid="start-session-button"]').click()

    // 参加者登録
    cy.window().then((win) => {
      const participantWin = win.open('http://localhost:3002', 'participant')
      cy.wrap(participantWin).then((pWin) => {
        setTimeout(() => {
          const nicknameInput = pWin.document.querySelector('#nickname')
          const submitButton = pWin.document.querySelector('button[type="submit"]')
          
          if (nicknameInput && submitButton) {
            nicknameInput.value = '終了テスト参加者'
            submitButton.click()
          }
        }, 1000)
      })
    })

    // クイズ実行
    cy.get('[data-testid="quiz-select"]').select(0)
    cy.get('[data-testid="start-quiz-button"]').click()
    cy.wait(2000)

    // セッション終了
    cy.get('[data-testid="end-session-button"]').click()
    cy.contains('セッション終了').should('be.visible')

    // 最終結果の確認
    cy.visit('http://localhost:3001/ranking')
    cy.contains('終了テスト参加者').should('be.visible')
    cy.get('[data-testid="final-ranking"]').should('be.visible')

    // 参加者側で終了画面が表示されることを確認
    cy.window().then((win) => {
      const participantWin = win.open('http://localhost:3002', 'participant-check')
      cy.wrap(participantWin).then((pWin) => {
        // 終了画面の確認
        setTimeout(() => {
          const endMessage = pWin.document.querySelector('[data-testid="quiz-ended"]')
          expect(endMessage).to.exist
        }, 2000)
      })
    })
  })
})