// ***********************************************************
// This example support/e2e.js is processed and
// loaded automatically before your test files.
//
// This is a great place to put global configuration and
// behavior that modifies Cypress.
//
// You can change the location of this file or turn off
// automatically serving support files with the
// 'supportFile' configuration option.
//
// You can read more here:
// https://on.cypress.io/configuration
// ***********************************************************

// Import commands.js using ES2015 syntax:
import './commands'

// Alternatively you can use CommonJS syntax:
// require('./commands')

// WebSocket接続のテスト用設定
Cypress.on('window:before:load', (win) => {
  // WebSocketモック用のカスタム実装
  const originalWebSocket = win.WebSocket
  
  win.WebSocket = function(url) {
    const ws = new originalWebSocket(url)
    
    // デバッグ用のログ出力
    ws.addEventListener('open', () => {
      cy.log(`WebSocket connected to: ${url}`)
    })
    
    ws.addEventListener('close', () => {
      cy.log(`WebSocket disconnected from: ${url}`)
    })
    
    ws.addEventListener('error', (error) => {
      cy.log(`WebSocket error: ${error}`)
    })
    
    return ws
  }
  
  // WebSocketクラスのプロパティを継承
  win.WebSocket.CONNECTING = originalWebSocket.CONNECTING
  win.WebSocket.OPEN = originalWebSocket.OPEN
  win.WebSocket.CLOSING = originalWebSocket.CLOSING
  win.WebSocket.CLOSED = originalWebSocket.CLOSED
})

// ページロード時のエラーを無視（開発環境用）
Cypress.on('uncaught:exception', (err, runnable) => {
  // WebSocket関連のエラーや開発環境でのHMRエラーを無視
  if (err.message.includes('WebSocket') || 
      err.message.includes('ChunkLoadError') ||
      err.message.includes('Loading chunk')) {
    return false
  }
  
  // 他のエラーは通常通り処理
  return true
})