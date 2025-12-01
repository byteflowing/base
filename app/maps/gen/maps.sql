CREATE TABLE map_account
(
    id         BIGSERIAL PRIMARY KEY,            -- '主键'
    name       VARCHAR(100) NOT NULL DEFAULT '', -- '账号名'
    map_source SMALLINT     NOT NULL DEFAULT 0,  -- '地图来源'
    map_type   SMALLINT     NOT NULL DEFAULT 0,  -- '地图类型'
    key        VARCHAR(100) NOT NULL DEFAULT '', -- '地图key'
    status     SMALLINT     NOT NULL DEFAULT 0,  -- '地图状态'
    owner_type SMALLINT     NOT NULL DEFAULT 0,  -- '地图拥有者类型'
    object_id  BIGINT       NOT NULL DEFAULT 0,  -- '地图拥有者id，自有地图为0'
    comment    VARCHAR(255),                     -- '备注'
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX map_account_object_name ON map_account (object_id, name);
CREATE INDEX map_account_uni_created_status ON map_account (created_at DESC, status);

CREATE TABLE map_interface
(
    id             BIGSERIAL PRIMARY KEY,          -- '主键'
    map_id         BIGINT      NOT NULL DEFAULT 0, -- 'map id'
    interface_type INT         NOT NULL DEFAULT 0, -- '接口类型'
    second_limit   INT,                            -- 'qps'
    daily_limit    INT,                            -- '日限额'
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX map_interface_unique_map_id_interface_type ON map_interface (map_id, interface_type);

CREATE TABLE map_interface_count
(
    id             BIGSERIAL PRIMARY KEY,              -- '主键'
    map_id         BIGINT      NOT NULL DEFAULT 0,     -- 'map id'
    interface_type INT         NOT NULL DEFAULT 0,     -- '接口类型'
    day            DATE        NOT NULL DEFAULT now(), -- '日期'
    count          INT         NOT NULL DEFAULT 0,     -- '调用数量'
    err_count      INT         NOT NULL DEFAULT 0,     -- '错误次数'
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX map_interface_count_unique_count_interface_type ON map_interface_count (map_id, interface_type);
CREATE INDEX map_interface_count_day ON map_interface_count (day);
CREATE INDEX map_interface_count_created_at ON map_interface_count (created_at);