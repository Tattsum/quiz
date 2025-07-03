-- クイズ大会システム データベーススキーマ
-- PostgreSQL / MySQL 対応

-- 管理者テーブル（認証あり）
CREATE TABLE administrators (
    id BIGSERIAL PRIMARY KEY,  -- MySQL: BIGINT AUTO_INCREMENT PRIMARY KEY
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 参加者テーブル（匿名、ニックネームのみ）
CREATE TABLE participants (
    id BIGSERIAL PRIMARY KEY,  -- MySQL: BIGINT AUTO_INCREMENT PRIMARY KEY
    nickname VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- クイズテーブル（問題文、選択肢、正解、メディアURL）
CREATE TABLE quizzes (
    id BIGSERIAL PRIMARY KEY,  -- MySQL: BIGINT AUTO_INCREMENT PRIMARY KEY
    question_text TEXT NOT NULL,
    option_a VARCHAR(255) NOT NULL,
    option_b VARCHAR(255) NOT NULL,
    option_c VARCHAR(255) NOT NULL,
    option_d VARCHAR(255) NOT NULL,
    correct_answer CHAR(1) NOT NULL CHECK (correct_answer IN ('A', 'B', 'C', 'D')),
    image_url VARCHAR(500),
    video_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 回答記録テーブル（参加者、問題、選択肢、正解/不正解）
CREATE TABLE answers (
    id BIGSERIAL PRIMARY KEY,  -- MySQL: BIGINT AUTO_INCREMENT PRIMARY KEY
    participant_id BIGINT NOT NULL,
    quiz_id BIGINT NOT NULL,
    selected_option CHAR(1) NOT NULL CHECK (selected_option IN ('A', 'B', 'C', 'D')),
    is_correct BOOLEAN NOT NULL,
    answered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (participant_id) REFERENCES participants(id) ON DELETE CASCADE,
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE,
    UNIQUE(participant_id, quiz_id)  -- 一人の参加者が同じ問題に複数回答することを防ぐ
);

-- セッション管理テーブル（現在の問題番号、投票受付状態）
CREATE TABLE quiz_sessions (
    id BIGSERIAL PRIMARY KEY,  -- MySQL: BIGINT AUTO_INCREMENT PRIMARY KEY
    current_quiz_id BIGINT,
    is_accepting_answers BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (current_quiz_id) REFERENCES quizzes(id) ON DELETE SET NULL
);

-- インデックス作成（パフォーマンス向上）
CREATE INDEX idx_answers_participant_id ON answers(participant_id);
CREATE INDEX idx_answers_quiz_id ON answers(quiz_id);
CREATE INDEX idx_answers_answered_at ON answers(answered_at);
CREATE INDEX idx_quiz_sessions_current_quiz_id ON quiz_sessions(current_quiz_id);

-- MySQL用の自動更新トリガー（PostgreSQLでは不要）
-- MySQL使用時のみ以下を実行
/*
DELIMITER $$
CREATE TRIGGER administrators_updated_at
    BEFORE UPDATE ON administrators
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER quizzes_updated_at
    BEFORE UPDATE ON quizzes
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER quiz_sessions_updated_at
    BEFORE UPDATE ON quiz_sessions
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$
DELIMITER ;
*/
