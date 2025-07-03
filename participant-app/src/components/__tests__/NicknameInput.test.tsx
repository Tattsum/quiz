import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import NicknameInput from '../NicknameInput'

describe('NicknameInput', () => {
  const mockOnSubmit = jest.fn()

  beforeEach(() => {
    mockOnSubmit.mockClear()
  })

  it('正常にレンダリングされる', () => {
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    expect(screen.getByText('クイズに参加')).toBeInTheDocument()
    expect(screen.getByText('ニックネームを入力してください')).toBeInTheDocument()
    expect(screen.getByLabelText('ニックネーム')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'クイズに参加する' })).toBeInTheDocument()
  })

  it('有効なニックネームで送信できる', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    await user.type(input, 'テストユーザー')
    await user.click(submitButton)
    
    expect(mockOnSubmit).toHaveBeenCalledWith('テストユーザー')
  })

  it('空のニックネームではエラーが表示される', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    await user.click(submitButton)
    
    expect(screen.getByText('ニックネームを入力してください')).toBeInTheDocument()
    expect(mockOnSubmit).not.toHaveBeenCalled()
  })

  it('2文字未満のニックネームではエラーが表示される', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    await user.type(input, 'あ')
    await user.click(submitButton)
    
    expect(screen.getByText('ニックネームは2文字以上で入力してください')).toBeInTheDocument()
    expect(mockOnSubmit).not.toHaveBeenCalled()
  })

  it('20文字を超えるニックネームではエラーが表示される', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    const longNickname = 'あ'.repeat(21)
    await user.type(input, longNickname)
    await user.click(submitButton)
    
    expect(screen.getByText('ニックネームは20文字以下で入力してください')).toBeInTheDocument()
    expect(mockOnSubmit).not.toHaveBeenCalled()
  })

  it('前後の空白がトリムされる', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    await user.type(input, '  テストユーザー  ')
    await user.click(submitButton)
    
    expect(mockOnSubmit).toHaveBeenCalledWith('テストユーザー')
  })

  it('ローディング状態では入力とボタンが無効になる', () => {
    render(<NicknameInput onSubmit={mockOnSubmit} loading={true} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: '参加中...' })
    
    expect(input).toBeDisabled()
    expect(submitButton).toBeDisabled()
  })

  it('空の入力ではボタンが無効になる', () => {
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    expect(submitButton).toBeDisabled()
  })

  it('入力後にボタンが有効になる', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    expect(submitButton).toBeDisabled()
    
    await user.type(input, 'テスト')
    
    expect(submitButton).not.toBeDisabled()
  })

  it('Enterキーで送信できる', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    
    await user.type(input, 'テストユーザー')
    await user.keyboard('{Enter}')
    
    expect(mockOnSubmit).toHaveBeenCalledWith('テストユーザー')
  })

  it('入力フィールドが正しく設定されている', () => {
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    
    expect(input).toHaveAttribute('type', 'text')
    expect(input).toHaveAttribute('id', 'nickname')
  })

  it('placeholderが表示される', () => {
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    
    expect(input).toHaveAttribute('placeholder', '例: たろう')
  })

  it('エラー時にエラーメッセージが表示される', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    // 無効な短い入力でエラーを発生させる
    await act(async () => {
      await user.type(input, 'あ')
      await user.click(submitButton)
    })
    
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
  })

  it('参加者数の上限が表示される', () => {
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    expect(screen.getByText('参加者は最大70名まで')).toBeInTheDocument()
  })

  it('autofocusが設定されている', () => {
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    
    expect(input).toHaveFocus()
  })

  it('有効な入力でエラーがクリアされる', async () => {
    const user = userEvent.setup()
    render(<NicknameInput onSubmit={mockOnSubmit} />)
    
    const input = screen.getByLabelText('ニックネーム')
    const submitButton = screen.getByRole('button', { name: 'クイズに参加する' })
    
    // 短い入力でエラーを発生させる
    await act(async () => {
      await user.type(input, 'あ')
      await user.click(submitButton)
    })
    
    // エラーメッセージが表示されることを確認
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument()
    })
    
    // 有効な入力に変更
    await act(async () => {
      await user.clear(input)
      await user.type(input, 'テストユーザー')
    })
    
    // エラーがクリアされることを確認 
    await waitFor(() => {
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })
    
    // フォーム送信
    await act(async () => {
      await user.click(submitButton)
    })
    
    expect(mockOnSubmit).toHaveBeenCalledWith('テストユーザー')
  })
})