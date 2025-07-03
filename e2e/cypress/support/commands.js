// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************

// 管理者ログイン用のカスタムコマンド
Cypress.Commands.add('adminLogin', (username = 'testadmin', password = 'password') => {
  cy.visit('http://localhost:3001/login')
  cy.get('input[name="username"]').type(username)
  cy.get('input[name="password"]').type(password)
  cy.get('button[type="submit"]').click()
  cy.url().should('include', '/dashboard')
})

// 参加者登録用のカスタムコマンド
Cypress.Commands.add('registerParticipant', (nickname) => {
  cy.visit('http://localhost:3002')
  cy.get('input[id="nickname"]').type(nickname)
  cy.get('button[type="submit"]').click()
})

// クイズ作成用のカスタムコマンド
Cypress.Commands.add('createQuiz', (quizData) => {
  const defaultQuizData = {
    question_text: 'テスト問題',
    option_a: '選択肢A',
    option_b: '選択肢B', 
    option_c: '選択肢C',
    option_d: '選択肢D',
    correct_answer: 'A',
    ...quizData
  }
  
  cy.visit('http://localhost:3001/questions/create')
  cy.get('input[name="question_text"]').type(defaultQuizData.question_text)
  cy.get('input[name="option_a"]').type(defaultQuizData.option_a)
  cy.get('input[name="option_b"]').type(defaultQuizData.option_b)
  cy.get('input[name="option_c"]').type(defaultQuizData.option_c)
  cy.get('input[name="option_d"]').type(defaultQuizData.option_d)
  cy.get('select[name="correct_answer"]').select(defaultQuizData.correct_answer)
  cy.get('button[type="submit"]').click()
})

// WebSocket接続の待機用コマンド
Cypress.Commands.add('waitForWebSocket', (timeout = 5000) => {
  cy.window().its('WebSocket').should('exist')
  cy.wait(1000) // WebSocket接続の安定化を待つ
})

// APIレスポンスの待機用コマンド
Cypress.Commands.add('waitForApiResponse', (alias, timeout = 10000) => {
  cy.wait(alias, { timeout })
})

// ローカルストレージのクリア
Cypress.Commands.add('clearLocalStorage', () => {
  cy.window().then((win) => {
    win.localStorage.clear()
  })
})

// セッションストレージのクリア
Cypress.Commands.add('clearSessionStorage', () => {
  cy.window().then((win) => {
    win.sessionStorage.clear()
  })
})

// データベースのリセット（テスト用）
Cypress.Commands.add('resetDatabase', () => {
  cy.request({
    method: 'POST',
    url: 'http://localhost:8080/api/test/reset-db',
    failOnStatusCode: false
  })
})

// 複数参加者のシミュレーション
Cypress.Commands.add('simulateMultipleParticipants', (count = 5) => {
  for (let i = 0; i < count; i++) {
    cy.window().then((win) => {
      const newWindow = win.open('http://localhost:3002', `participant-${i}`)
      cy.wrap(newWindow).as(`participant-${i}`)
    })
  }
})

// スクリーンショット撮影（カスタム）
Cypress.Commands.add('captureScreenshot', (name) => {
  cy.screenshot(name, {
    capture: 'viewport',
    clip: { x: 0, y: 0, width: 1280, height: 720 }
  })
})

// エラーログのチェック
Cypress.Commands.add('checkConsoleErrors', () => {
  cy.window().then((win) => {
    const logs = []
    const originalError = win.console.error
    
    win.console.error = function(...args) {
      logs.push(args.join(' '))
      originalError.apply(win.console, args)
    }
    
    cy.wrap(logs).as('consoleLogs')
  })
})

// レスポンシブテスト用のビューポート設定
Cypress.Commands.add('setMobileViewport', () => {
  cy.viewport(375, 667) // iPhone SE
})

Cypress.Commands.add('setTabletViewport', () => {
  cy.viewport(768, 1024) // iPad
})

Cypress.Commands.add('setDesktopViewport', () => {
  cy.viewport(1920, 1080) // Desktop
})