CREATE TYPE room_visibility AS ENUM ('public', 'private');

CREATE TYPE room_theme AS ENUM ('default', 'farm', 'horror', 'fun', 'space');

-- Tabela de salas
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(25) NOT NULL,
    theme room_theme NOT NULL DEFAULT 'default',
    visibility room_visibility NOT NULL DEFAULT 'public',
    access_code VARCHAR(10),
    max_seats INTEGER NOT NULL DEFAULT 16,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Índice para buscar salas por dono
CREATE INDEX idx_rooms_owner_id ON rooms(owner_id);

-- Índice para buscar salas públicas não deletadas
CREATE INDEX idx_rooms_public_active ON rooms(visibility, deleted_at) 
    WHERE visibility = 'public' AND deleted_at IS NULL;

-- Índice para buscar por código de acesso
CREATE INDEX idx_rooms_access_code ON rooms(access_code) 
    WHERE access_code IS NOT NULL;

-- Índice para ordenação por data
CREATE INDEX idx_rooms_created_at ON rooms(created_at DESC);
