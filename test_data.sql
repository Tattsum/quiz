-- テスト用のサンプルデータ
-- このファイルはCI環境で自動的に読み込まれます

-- 管理者テストデータ
INSERT INTO administrators (id, username, password_hash, email) VALUES
(1, 'admin', '$2a$10$1iUcDcN76V09xV2EHF8xyuL9m.soCXbkd7ip6U9DbfAkXQbyW6Ktm', 'admin@example.com');

-- 参加者テストデータ
INSERT INTO participants (id, nickname) VALUES
(1, 'TestUser1'),
(2, 'TestUser2'),
(3, 'TestUser3');

-- クイズテストデータ
INSERT INTO quizzes (id, question_text, option_a, option_b, option_c, option_d, correct_answer) VALUES
(1, 'What is 2+2?', '3', '4', '5', '6', 'B'),
(2, 'What is the capital of Japan?', 'Tokyo', 'Osaka', 'Kyoto', 'Nagoya', 'A'),
(3, 'What is 5*3?', '15', '12', '18', '20', 'A');

-- セッション管理テストデータ
INSERT INTO quiz_sessions (id, current_quiz_id, is_accepting_answers, created_at) VALUES
(1, 3, true, CURRENT_TIMESTAMP);

-- 回答テストデータ（UpdateAnswerテスト用）  
INSERT INTO answers (id, participant_id, quiz_id, selected_option, is_correct) VALUES
(1, 2, 2, 'A', true);

-- ID シーケンスの調整
SELECT setval('administrators_id_seq', 1, true);
SELECT setval('participants_id_seq', 3, true);
SELECT setval('quizzes_id_seq', 3, true);
SELECT setval('quiz_sessions_id_seq', 1, true);
SELECT setval('answers_id_seq', 1, true);