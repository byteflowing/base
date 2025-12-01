CREATE TABLE user_account
(
    id                  BIGINT PRIMARY KEY,
    tenant_id           VARCHAR(50)  NOT NULL DEFAULT '',    -- '租户id'
    number              VARCHAR(50)  NOT NULL DEFAULT '',    -- '用户编号'
    name                VARCHAR(50),                         -- '用户名称'
    alias               VARCHAR(50),                         -- '昵称'
    password            VARCHAR(100),                        -- '密码'
    avatar              VARCHAR(255),                        -- '头像'
    gender              SMALLINT,                            -- '性别'
    birthday            DATE,                                -- '生日'
    phone_country_code  VARCHAR(10)  NOT NULL DEFAULT '',    -- '手机国家编码'
    phone               VARCHAR(20)  NOT NULL DEFAULT '',    -- '手机号码'
    email               VARCHAR(100) NOT NULL DEFAULT '',    -- '邮箱'
    country_code        VARCHAR(20)  NOT NULL DEFAULT '',    -- '国家编码'
    province_code       VARCHAR(20)  NOT NULL DEFAULT '',    -- '省份编码'
    city_code           VARCHAR(20)  NOT NULL DEFAULT '',    -- '城市编码'
    district_code       VARCHAR(20)  NOT NULL DEFAULT '',    -- '区县编码'
    addr                VARCHAR(255),                        -- '详细地址'
    status              SMALLINT     NOT NULL DEFAULT 0,     -- '用户状态枚举'
    source              SMALLINT     NOT NULL DEFAULT 0,     -- '用户注册来源'
    signup_type         SMALLINT     NOT NULL DEFAULT 0,     -- '用户注册类型'
    phone_verified      BOOLEAN      NOT NULL DEFAULT FALSE, -- '手机是否验证'
    email_verified      BOOLEAN      NOT NULL DEFAULT FALSE, -- '邮箱是否验证'
    type                SMALLINT,                            -- '用户类型，预留字段，由业务方维护'
    level               INT,                                 -- '用户等级，预留字段，由业务方维护'
    register_ip         VARCHAR(50),                         -- '注册ip'
    register_device     VARCHAR(255),                        -- '注册设备'
    register_agent      VARCHAR(255),                        -- '注册UA'
    register_location   VARCHAR(50),                         -- '注册地'
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now(), -- '更新时间'
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(), -- '创建时间'
    deleted_at          TIMESTAMPTZ,                         -- '删除时间'
    password_updated_at TIMESTAMPTZ,                         -- '密码更改时间'
    ext                 JSON                                 -- '扩展字段'
);
CREATE UNIQUE INDEX idx_user_account_number ON user_account (tenant_id, number);
CREATE UNIQUE INDEX idx_user_account_email ON user_account (tenant_id, email);
CREATE UNIQUE INDEX idx_user_account_phone ON user_account (tenant_id, phone_country_code, phone);
CREATE INDEX idx_user_account_region ON user_account (country_code, province_code, city_code, district_code);

CREATE TABLE user_auth
(
    id         BIGSERIAL PRIMARY KEY,
    tenant_id  VARCHAR(50)  NOT NULL DEFAULT '',    -- '租户id'
    uid        BIGINT       NOT NULL DEFAULT 0,     -- '[必选] 用户id'
    type       SMALLINT     NOT NULL DEFAULT 0,     -- '[必选] 认证类型枚举'
    status     SMALLINT     NOT NULL DEFAULT 0,     -- '[必选] 状态枚举'
    appid      VARCHAR(100) NOT NULL DEFAULT '',    -- '[必选] appid，如微信小程序的AppId'
    open_id    VARCHAR(100) NOT NULL DEFAULT '',    -- '[必选] open_id'
    union_id   VARCHAR(100) NOT NULL DEFAULT '',    -- '[必选] unionid'
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(), -- '[必选] 更新时间'
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(), -- '[必选] 创建时间'
    deleted_at TIMESTAMPTZ                          -- '[可选] 删除时间'
);
CREATE INDEX idx_user_auth_uid ON user_auth (uid);
CREATE INDEX idx_user_auth_app_id ON user_auth (appid);
CREATE INDEX idx_user_auth_tenant_id ON user_auth (tenant_id);
CREATE INDEX idx_user_auth_union_id ON user_auth (union_id);
CREATE UNIQUE INDEX idx_uni_user_auth_open_id ON user_auth (open_id);


CREATE TABLE user_sign_log
(
    id                 BIGSERIAL PRIMARY KEY,
    tenant_id          VARCHAR(50) NOT NULL DEFAULT '',    -- '租户id'
    uid                BIGINT      NOT NULL DEFAULT 0,     -- '用户id'
    type               SMALLINT    NOT NULL DEFAULT 0,     -- '认证类型枚举'
    status             SMALLINT    NOT NULL DEFAULT 0,     -- '状态'
    identifier         VARCHAR(50) NOT NULL DEFAULT '',    -- '登录的账号信息appid 邮箱 手机号等'
    ip                 VARCHAR(50),                        -- '登录ip'
    location           VARCHAR(100),                       -- '位置'
    agent              VARCHAR(255),                       -- '登录软件信息'
    device             VARCHAR(255),                       -- '登录设备信息'
    access_jti         VARCHAR(64) NOT NULL DEFAULT '',    -- 'access token jti'
    refresh_jti        VARCHAR(64) NOT NULL DEFAULT '',    -- 'refresh token jti'
    access_expired_at  TIMESTAMPTZ,                        -- 'access token 过期时间'
    refresh_expired_at TIMESTAMPTZ,                        -- 'refresh token 过期时间'
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(), -- '更新时间'
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(), -- '创建时间'
    deleted_at         TIMESTAMPTZ                         -- '删除时间'
);
CREATE INDEX idx_user_sign_log_uni_uid_created_at ON user_sign_log (tenant_id, uid, created_at DESC);
CREATE UNIQUE INDEX idx_user_sign_log_access_token_id ON user_sign_log (access_jti);
CREATE UNIQUE INDEX idx_user_sign_log_refresh_token_id ON user_sign_log (refresh_jti);
