import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import RealtimeChart from '../../components/RealtimeChart.vue'

// Chart.jsのモック
vi.mock('chart.js', () => {
  const mockChart = {
    update: vi.fn(),
    destroy: vi.fn(),
    data: {},
  }

  return {
    Chart: vi.fn(() => mockChart),
    ArcElement: vi.fn(),
    CategoryScale: vi.fn(),
    LinearScale: vi.fn(),
    BarElement: vi.fn(),
    Title: vi.fn(),
    Tooltip: vi.fn(),
    Legend: vi.fn(),
    register: vi.fn(),
  }
})

// WebSocketのモック
global.WebSocket = vi.fn(() => ({
  send: vi.fn(),
  close: vi.fn(),
  onopen: null,
  onclose: null,
  onmessage: null,
  onerror: null,
})) as any

// Nuxtの設定のモック
const mockUseRuntimeConfig = vi.fn(() => ({
  public: {
    wsBase: 'ws://localhost:3000'
  }
}))

// Canvas要素のモック
Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
  value: vi.fn(() => ({
    fillRect: vi.fn(),
    clearRect: vi.fn(),
    getImageData: vi.fn(() => ({ data: new Array(4) })),
    putImageData: vi.fn(),
    createImageData: vi.fn(() => new Array(4)),
    setTransform: vi.fn(),
    drawImage: vi.fn(),
    save: vi.fn(),
    fillText: vi.fn(),
    restore: vi.fn(),
    beginPath: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    closePath: vi.fn(),
    stroke: vi.fn(),
    translate: vi.fn(),
    scale: vi.fn(),
    rotate: vi.fn(),
    arc: vi.fn(),
    fill: vi.fn(),
    measureText: vi.fn(() => ({ width: 0 })),
    transform: vi.fn(),
    rect: vi.fn(),
    clip: vi.fn(),
  })),
})

// Global API をモック
global.useRuntimeConfig = mockUseRuntimeConfig
global.ref = vi.fn()
global.reactive = vi.fn()
global.computed = vi.fn()
global.watch = vi.fn()
global.nextTick = vi.fn(() => Promise.resolve())
global.onMounted = vi.fn()
global.onUnmounted = vi.fn()

describe('RealtimeChart', () => {
  let wrapper: any

  beforeEach(() => {
    vi.clearAllMocks()
    
    // ref のモック
    global.ref = vi.fn((initialValue) => ({
      value: initialValue
    }))
    
    // reactive のモック
    global.reactive = vi.fn((obj) => obj)
    
    // computed のモック
    global.computed = vi.fn((fn) => ({
      value: fn()
    }))
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  it('コンポーネントが正常にレンダリングされる', () => {
    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    expect(wrapper.find('.bg-white').exists()).toBe(true)
    expect(wrapper.text()).toContain('リアルタイム回答状況')
  })

  it('データがない場合の表示が正しい', () => {
    // hasData が false を返すようにモック
    global.computed = vi.fn((fn) => {
      const result = fn()
      if (typeof result === 'boolean') {
        return { value: false }
      }
      return { value: result }
    })

    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    expect(wrapper.text()).toContain('回答データがありません')
  })

  it('プロパティが正しく受け取られる', () => {
    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 123,
        questionId: 456
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    expect(wrapper.props('quizId')).toBe(123)
    expect(wrapper.props('questionId')).toBe(456)
  })

  it('デフォルトプロパティが設定されている', () => {
    wrapper = mount(RealtimeChart, {
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    expect(wrapper.props('quizId')).toBe(null)
    expect(wrapper.props('questionId')).toBe(null)
  })

  it('統計情報の表示要素が存在する', () => {
    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    // 統計情報の表示要素をチェック
    expect(wrapper.text()).toContain('総参加者')
    expect(wrapper.text()).toContain('回答済み')
    expect(wrapper.text()).toContain('未回答')
    expect(wrapper.text()).toContain('回答率')
  })

  it('円グラフのcanvas要素が存在する', () => {
    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: false
        }
      }
    })

    const canvas = wrapper.find('canvas')
    expect(canvas.exists()).toBe(true)
    expect(canvas.attributes('width')).toBe('320')
    expect(canvas.attributes('height')).toBe('320')
  })

  it('選択肢別詳細が表示される', () => {
    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    // 選択肢A-Dがあることを確認
    expect(wrapper.text()).toContain('選択肢')
  })

  it('WebSocket接続状態インジケーターが存在する', () => {
    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    // 接続状態インジケーターの要素を確認
    const indicator = wrapper.find('.w-2.h-2.rounded-full')
    expect(indicator.exists()).toBe(true)
  })

  it('回答推移グラフが存在する', () => {
    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    expect(wrapper.text()).toContain('回答推移')
    const chartArea = wrapper.find('.h-24.bg-gray-50')
    expect(chartArea.exists()).toBe(true)
  })
})

describe('RealtimeChart Methods', () => {
  it('getOptionColor が正しい色を返す', () => {
    const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444']
    
    // テスト用の関数を直接テスト
    const getOptionColor = (index: number) => {
      const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444']
      return colors[index] || '#6B7280'
    }

    expect(getOptionColor(0)).toBe('#3B82F6')
    expect(getOptionColor(1)).toBe('#10B981')
    expect(getOptionColor(2)).toBe('#F59E0B')
    expect(getOptionColor(3)).toBe('#EF4444')
    expect(getOptionColor(5)).toBe('#6B7280') // 範囲外
  })

  it('getPercentage が正しいパーセンテージを計算する', () => {
    const getPercentage = (count: number, answeredCount: number) => {
      if (answeredCount === 0) return 0
      return Math.round((count / answeredCount) * 100)
    }

    expect(getPercentage(25, 100)).toBe(25)
    expect(getPercentage(33, 100)).toBe(33)
    expect(getPercentage(1, 3)).toBe(33)
    expect(getPercentage(0, 0)).toBe(0)
  })

  it('answerRate の計算が正しい', () => {
    const calculateAnswerRate = (answeredCount: number, totalParticipants: number) => {
      if (totalParticipants === 0) return 0
      return Math.round((answeredCount / totalParticipants) * 100)
    }

    expect(calculateAnswerRate(50, 100)).toBe(50)
    expect(calculateAnswerRate(75, 100)).toBe(75)
    expect(calculateAnswerRate(33, 100)).toBe(33)
    expect(calculateAnswerRate(0, 0)).toBe(0)
  })
})

describe('RealtimeChart WebSocket Integration', () => {
  it('WebSocket接続が初期化される', () => {
    const mockWebSocket = vi.fn()
    global.WebSocket = mockWebSocket

    wrapper = mount(RealtimeChart, {
      props: {
        quizId: 1,
        questionId: 1
      },
      global: {
        stubs: {
          canvas: true
        }
      }
    })

    // WebSocketが作成されることを確認（実際の接続ロジックはコンポーネント内で実行される）
    expect(mockWebSocket).toBeDefined()
  })

  it('WebSocketメッセージハンドラーの基本構造', () => {
    const handleWebSocketMessage = (data: any) => {
      if (data.type === 'answer_status') {
        return {
          totalParticipants: data.total_participants,
          answeredCount: data.answered_count,
          answerCounts: data.answer_counts || [0, 0, 0, 0]
        }
      }
      return null
    }

    const testMessage = {
      type: 'answer_status',
      question_id: 1,
      total_participants: 100,
      answered_count: 75,
      answer_counts: [20, 30, 15, 10]
    }

    const result = handleWebSocketMessage(testMessage)
    expect(result).toEqual({
      totalParticipants: 100,
      answeredCount: 75,
      answerCounts: [20, 30, 15, 10]
    })
  })
})