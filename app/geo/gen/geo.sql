CREATE TABLE geo_country
(
    id            BIGSERIAL PRIMARY KEY,              -- 主键
    cca2          CHAR(2)     NOT NULL UNIQUE,        -- ISO 3166-1 alpha-2 代码，例如 "US"
    cca3          CHAR(3)     NOT NULL UNIQUE,        -- ISO 3166-1 alpha-3 代码，例如 "USA"
    ccn3          CHAR(3)     NOT NULL UNIQUE,        -- 国家数字代码，例如：美国 840
    flag          VARCHAR(20) NOT NULL DEFAULT '',    -- 国旗 emoji格式
    continent     VARCHAR(50) NOT NULL DEFAULT '',    -- 所属洲
    sub_continent VARCHAR(50) NOT NULL DEFAULT '',    -- 细分洲
    multi_lang    JSONB,                              -- 多语言
    independent   BOOLEAN     NOT NULL,               -- 是否为独立国家
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,  -- 是否有效
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(), -- 创建时间
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(), -- 更新时间
    deleted_at    TIMESTAMPTZ                         -- 删除时间
);
CREATE UNIQUE INDEX idx_uniq_geo_region_country_cca2 ON geo_country (cca2);
CREATE UNIQUE INDEX idx_uniq_geo_region_country_cca3 ON geo_country (cca3);
CREATE UNIQUE INDEX idx_uniq_geo_region_country_ccn3 ON geo_country (ccn3);
CREATE INDEX idx_geo_country_multi_lang ON geo_country (multi_lang);

CREATE TABLE geo_region
(
    id           BIGSERIAL PRIMARY KEY,              -- 主键
    country_cca2 CHAR(2)     NOT NULL,               -- cca2
    source       SMALLINT    NOT NULL,               -- 来源枚举
    parent_code  VARCHAR(20) NOT NULL DEFAULT '',    -- 上级行政区编码
    code         VARCHAR(20) NOT NULL,               -- 行政区代码（国家代码/ISO码/自定义）
    level        SMALLINT    NOT NULL,               -- 层级（1=省/州, 2=市, 3=区县...）
    multi_lang   JSONB,                              -- 多语言
    is_active    BOOLEAN     NOT NULL DEFAULT TRUE,  -- 是否有效
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(), -- 创建时间
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(), -- 更新时间
    deleted_at   TIMESTAMPTZ                         -- 删除时间
);
CREATE INDEX idx_geo_region_parent_code ON geo_region (parent_code);
CREATE UNIQUE INDEX idx_uniq_geo_region_country_code_source ON geo_region (country_cca2, code, source);
CREATE INDEX idx_geo_region_multi_lang ON geo_region (multi_lang);

CREATE TABLE geo_phone_code
(
    id         BIGSERIAL PRIMARY KEY,              -- 主键
    name       VARCHAR(100) NOT NULL,              -- 英文名称
    phone_code VARCHAR(10)  NOT NULL,              -- E.164 前缀，带 '+'，例如 '+86', '+1', '+852'
    multi_lang JSONB,                              -- 多语言
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE, -- 是否有效
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_uniq_geo_phone_cca2 ON geo_phone_code (phone_code, name);
CREATE INDEX idx_geo_phone_code_multi_lang ON geo_phone_code (multi_lang);



