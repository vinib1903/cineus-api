-- Tabela de banimentos
CREATE TABLE room_bans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    banned_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(200),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Impede banir o mesmo usuário duas vezes na mesma sala
    CONSTRAINT unique_active_ban UNIQUE (room_id, user_id)
);

-- Índice para verificar se usuário está banido
CREATE INDEX idx_room_bans_room_user ON room_bans(room_id, user_id);

-- Índice para listar bans de uma sala
CREATE INDEX idx_room_bans_room_id ON room_bans(room_id);

