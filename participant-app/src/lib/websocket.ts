import { WebSocketMessage } from '@/types'

export type WebSocketEventHandler = (message: WebSocketMessage) => void

export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private handlers: Map<string, WebSocketEventHandler[]> = new Map()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private isConnecting = false
  private heartbeatInterval: NodeJS.Timeout | null = null

  constructor(url?: string) {
    const wsUrl = url || (
      typeof window !== 'undefined' 
        ? `ws://${window.location.hostname}:8080/ws`
        : 'ws://localhost:8080/ws'
    )
    this.url = wsUrl
  }

  connect(): Promise<void> {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return Promise.resolve()
    }

    this.isConnecting = true

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.isConnecting = false
          this.reconnectAttempts = 0
          this.startHeartbeat()
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error)
          }
        }

        this.ws.onclose = () => {
          console.log('WebSocket disconnected')
          this.isConnecting = false
          this.stopHeartbeat()
          this.handleReconnect()
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error)
          this.isConnecting = false
          reject(error)
        }
      } catch (error) {
        this.isConnecting = false
        reject(error)
      }
    })
  }

  disconnect() {
    this.stopHeartbeat()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  subscribe(quizId: string) {
    this.send({
      type: 'subscribe',
      data: { quiz_id: quizId }
    })
  }

  unsubscribe(quizId: string) {
    this.send({
      type: 'unsubscribe',
      data: { quiz_id: quizId }
    })
  }

  on(event: string, handler: WebSocketEventHandler) {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, [])
    }
    this.handlers.get(event)!.push(handler)
  }

  off(event: string, handler: WebSocketEventHandler) {
    const handlers = this.handlers.get(event)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index !== -1) {
        handlers.splice(index, 1)
      }
    }
  }

  private send(message: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket is not connected. Message not sent:', message)
    }
  }

  private handleMessage(message: WebSocketMessage) {
    const handlers = this.handlers.get(message.type)
    if (handlers) {
      handlers.forEach(handler => handler(message))
    }

    // 全てのメッセージを受信する汎用ハンドラ
    const allHandlers = this.handlers.get('*')
    if (allHandlers) {
      allHandlers.forEach(handler => handler(message))
    }
  }

  private startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      this.send({ type: 'heartbeat', data: {} })
    }, 30000) // 30秒ごとにハートビート送信
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = this.reconnectDelay * this.reconnectAttempts
      
      console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
      
      setTimeout(() => {
        this.connect().catch(error => {
          console.error('Reconnection failed:', error)
        })
      }, delay)
    } else {
      console.error('Max reconnection attempts reached')
    }
  }

  get isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

export const wsClient = new WebSocketClient()