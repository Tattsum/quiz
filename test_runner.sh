#!/bin/bash

# クイズ大会システム テスト実行スクリプト
# 全てのテストを順次実行し、結果をレポートとして出力

set -e

echo "================================================"
echo "クイズ大会システム 総合テスト実行"
echo "================================================"

# カラー出力用の定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# テスト結果を記録
TEST_RESULTS_DIR="./test-results"
mkdir -p $TEST_RESULTS_DIR

# ログファイル
LOG_FILE="$TEST_RESULTS_DIR/test-execution-$(date +%Y%m%d-%H%M%S).log"

# テスト実行関数
run_test() {
    local test_name="$1"
    local test_command="$2"
    local working_dir="$3"
    
    echo -e "${BLUE}[実行中] $test_name${NC}"
    echo "[$(date)] 開始: $test_name" >> $LOG_FILE
    
    if [ -n "$working_dir" ]; then
        cd "$working_dir"
    fi
    
    if eval "$test_command" >> $LOG_FILE 2>&1; then
        echo -e "${GREEN}[成功] $test_name${NC}"
        echo "[$(date)] 成功: $test_name" >> $LOG_FILE
        return 0
    else
        echo -e "${RED}[失敗] $test_name${NC}"
        echo "[$(date)] 失敗: $test_name" >> $LOG_FILE
        return 1
    fi
    
    if [ -n "$working_dir" ]; then
        cd - > /dev/null
    fi
}

# メイン実行
main() {
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    echo "テスト実行ログ: $LOG_FILE"
    echo "開始時間: $(date)" | tee -a $LOG_FILE
    echo ""
    
    echo -e "${YELLOW}=== 1. Go言語バックエンドテスト ===${NC}"
    
    # Go単体テスト
    total_tests=$((total_tests + 1))
    if run_test "Go単体テスト" "go test ./internal/..." "."; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    # Go統合テスト
    total_tests=$((total_tests + 1))
    if run_test "Go統合テスト" "go test -tags=integration ./..." "."; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    # WebSocketテスト
    total_tests=$((total_tests + 1))
    if run_test "WebSocketテスト" "go test -run TestWebSocket ./internal/handlers/" "."; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    echo ""
    echo -e "${YELLOW}=== 2. フロントエンドテスト ===${NC}"
    
    # Nuxt3管理ダッシュボードテスト
    if [ -d "./admin-dashboard" ]; then
        total_tests=$((total_tests + 1))
        if run_test "Nuxt3管理ダッシュボードテスト" "npm run test" "./admin-dashboard"; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    # Next.js参加者アプリテスト
    if [ -d "./participant-app" ]; then
        total_tests=$((total_tests + 1))
        if run_test "Next.js参加者アプリテスト" "npm run test" "./participant-app"; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    echo ""
    echo -e "${YELLOW}=== 3. パフォーマンステスト ===${NC}"
    
    # 70人同時接続テスト
    total_tests=$((total_tests + 1))
    if run_test "70人同時接続パフォーマンステスト" "go test -run TestConcurrent -timeout 300s" "."; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    # システム負荷テスト
    total_tests=$((total_tests + 1))
    if run_test "システム負荷テスト" "go test -run TestSystemLoad -timeout 300s" "."; then
        passed_tests=$((passed_tests + 1))
    else
        failed_tests=$((failed_tests + 1))
    fi
    
    echo ""
    echo -e "${YELLOW}=== 4. E2Eテスト ===${NC}"
    
    # Cypressテスト（E2Eディレクトリが存在する場合）
    if [ -d "./e2e" ]; then
        total_tests=$((total_tests + 1))
        if run_test "CypressE2Eテスト" "npm run cypress:run" "./e2e"; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    echo ""
    echo "================================================"
    echo -e "${BLUE}テスト実行結果サマリー${NC}"
    echo "================================================"
    echo "実行時間: $(date)" | tee -a $LOG_FILE
    echo "総テスト数: $total_tests" | tee -a $LOG_FILE
    echo -e "成功: ${GREEN}$passed_tests${NC}" | tee -a $LOG_FILE
    echo -e "失敗: ${RED}$failed_tests${NC}" | tee -a $LOG_FILE
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}✅ 全てのテストが成功しました！${NC}" | tee -a $LOG_FILE
        exit_code=0
    else
        echo -e "${RED}❌ $failed_tests 個のテストが失敗しました${NC}" | tee -a $LOG_FILE
        exit_code=1
    fi
    
    # テストカバレッジレポート生成
    echo ""
    echo -e "${YELLOW}=== 5. テストカバレッジレポート ===${NC}"
    if run_test "Goテストカバレッジ" "go test -coverprofile=$TEST_RESULTS_DIR/coverage.out ./... && go tool cover -html=$TEST_RESULTS_DIR/coverage.out -o $TEST_RESULTS_DIR/coverage.html" "."; then
        echo "カバレッジレポート: $TEST_RESULTS_DIR/coverage.html"
    fi
    
    echo ""
    echo "詳細ログ: $LOG_FILE"
    echo "テスト結果ディレクトリ: $TEST_RESULTS_DIR"
    
    exit $exit_code
}

# スクリプトの前処理
setup() {
    echo "テスト環境の準備..."
    
    # 必要なディレクトリの存在確認
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}エラー: go.modが見つかりません。正しいディレクトリで実行してください。${NC}"
        exit 1
    fi
    
    # Go依存関係の確認
    echo "Go依存関係の確認..."
    go mod tidy
    
    # フロントエンド依存関係の確認
    if [ -d "./admin-dashboard" ] && [ -f "./admin-dashboard/package.json" ]; then
        echo "管理ダッシュボードの依存関係確認..."
        cd admin-dashboard && npm install && cd ..
    fi
    
    if [ -d "./participant-app" ] && [ -f "./participant-app/package.json" ]; then
        echo "参加者アプリの依存関係確認..."
        cd participant-app && npm install && cd ..
    fi
    
    if [ -d "./e2e" ] && [ -f "./e2e/package.json" ]; then
        echo "E2Eテストの依存関係確認..."
        cd e2e && npm install && cd ..
    fi
}

# クリーンアップ関数
cleanup() {
    echo "テスト後のクリーンアップ..."
    # 一時ファイルやプロセスのクリーンアップ
    pkill -f "quiz" 2>/dev/null || true
    rm -f ./quiz 2>/dev/null || true
}

# シグナルハンドリング
trap cleanup EXIT

# ヘルプ表示
show_help() {
    echo "クイズ大会システム テスト実行スクリプト"
    echo ""
    echo "使用方法:"
    echo "  $0 [オプション]"
    echo ""
    echo "オプション:"
    echo "  -h, --help      このヘルプを表示"
    echo "  -s, --setup     テスト環境のセットアップのみ実行"
    echo "  -u, --unit      単体テストのみ実行"
    echo "  -i, --integration 統合テストのみ実行"
    echo "  -p, --performance パフォーマンステストのみ実行"
    echo "  -e, --e2e       E2Eテストのみ実行"
    echo "  -f, --frontend  フロントエンドテストのみ実行"
    echo ""
    echo "例:"
    echo "  $0                # 全てのテストを実行"
    echo "  $0 -u             # 単体テストのみ実行"
    echo "  $0 -p             # パフォーマンステストのみ実行"
}

# コマンドライン引数の処理
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    -s|--setup)
        setup
        echo "セットアップが完了しました。"
        exit 0
        ;;
    -u|--unit)
        setup
        echo -e "${YELLOW}=== 単体テストのみ実行 ===${NC}"
        run_test "Go単体テスト" "go test ./internal/..." "."
        run_test "フロントエンド単体テスト" "npm run test" "./admin-dashboard"
        run_test "参加者アプリ単体テスト" "npm run test" "./participant-app"
        exit $?
        ;;
    -i|--integration)
        setup
        echo -e "${YELLOW}=== 統合テストのみ実行 ===${NC}"
        run_test "Go統合テスト" "go test -tags=integration ./..." "."
        exit $?
        ;;
    -p|--performance)
        setup
        echo -e "${YELLOW}=== パフォーマンステストのみ実行 ===${NC}"
        run_test "パフォーマンステスト" "go test -run TestConcurrent -timeout 300s" "."
        run_test "システム負荷テスト" "go test -run TestSystemLoad -timeout 300s" "."
        exit $?
        ;;
    -e|--e2e)
        setup
        echo -e "${YELLOW}=== E2Eテストのみ実行 ===${NC}"
        if [ -d "./e2e" ]; then
            run_test "CypressE2Eテスト" "npm run cypress:run" "./e2e"
        else
            echo "E2Eテストディレクトリが見つかりません。"
        fi
        exit $?
        ;;
    -f|--frontend)
        setup
        echo -e "${YELLOW}=== フロントエンドテストのみ実行 ===${NC}"
        if [ -d "./admin-dashboard" ]; then
            run_test "管理ダッシュボードテスト" "npm run test" "./admin-dashboard"
        fi
        if [ -d "./participant-app" ]; then
            run_test "参加者アプリテスト" "npm run test" "./participant-app"
        fi
        exit $?
        ;;
    "")
        # 引数なしの場合は全テスト実行
        setup
        main
        ;;
    *)
        echo "不明なオプション: $1"
        echo "ヘルプを表示するには -h オプションを使用してください。"
        exit 1
        ;;
esac