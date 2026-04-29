CREATE TABLE subscriptions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name TEXT NOT NULL,
    price        INT NOT NULL CHECK (price >= 0),
    user_id      UUID NOT NULL,
    start_date   DATE NOT NULL,      -- будем хранить первый день месяца (2025-07-01)
    end_date     DATE,               -- NULL = бессрочная, иначе последний день месяца
    created_at   TIMESTAMP DEFAULT NOW(),
    updated_at   TIMESTAMP DEFAULT NOW()
);

-- Индексы для фильтрации:
CREATE INDEX idx_user_id ON subscriptions(user_id);
CREATE INDEX idx_service_name ON subscriptions(service_name);
CREATE INDEX idx_dates ON subscriptions(start_date, end_date);