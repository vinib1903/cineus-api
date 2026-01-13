-- Mensagens de sala
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Índice para listar mensagens de uma sala
CREATE INDEX idx_chat_messages_room_id ON chat_messages(room_id, created_at DESC);

-- Mensagens diretas
CREATE TABLE direct_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content VARCHAR(500) NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Índice para listar conversa entre dois usuários
CREATE INDEX idx_direct_messages_conversation ON direct_messages(
    LEAST(from_user_id, to_user_id),
    GREATEST(from_user_id, to_user_id),
    created_at DESC
);

-- Índice para contar mensagens não lidas
CREATE INDEX idx_direct_messages_unread ON direct_messages(to_user_id, read_at)
    WHERE read_at IS NULL;
