CREATE TABLE user_basic
(
    id                 BIGINT PRIMARY KEY,
    number             VARCHAR(50)  NOT NULL DEFAULT '',  -- '[必选] 用户编号'
    name               VARCHAR(50),                       -- '[可选] 用户名称'
    alias              VARCHAR(50),                       -- '[可选] 昵称'
    password           VARCHAR(100),                      -- '[可选] 密码'
    avatar             VARCHAR(255),                      -- '[可选] 头像'
    gender             SMALLINT,                          -- '[可选] 性别'
    birthday           DATE,                              -- '[可选] 生日'
    phone_country_code VARCHAR(10)  NOT NULL DEFAULT '',  -- '[必选] 手机国家编码'
    phone              VARCHAR(20)  NOT NULL DEFAULT '',  -- '[必选] 手机号码'
    email              VARCHAR(100) NOT NULL DEFAULT '',  -- '[必选] 邮箱'
    country_code       VARCHAR(20)  NOT NULL DEFAULT '',  -- '[必选] 国家编码'
    province_code      VARCHAR(20)  NOT NULL DEFAULT '',  -- '[必选] 省份编码'
    city_code          VARCHAR(20)  NOT NULL DEFAULT '',  -- '[必选] 城市编码'
    district_code      VARCHAR(20)  NOT NULL DEFAULT '',  -- '[必选] 区县编码'
    addr               VARCHAR(255),                      -- '[可选] 详细地址'
    status             SMALLINT     NOT NULL DEFAULT 0,   -- '[必选] 用户状态枚举'
    source             SMALLINT     NOT NULL DEFAULT 0,   -- '[必选] 用户注册来源'
    signup_type        SMALLINT     NOT NULL DEFAULT 0,   -- '[必选] 注册方式'
    phone_verified     SMALLINT     NOT NULL DEFAULT 0,   -- '[必选] 手机是否验证'
    email_verified     SMALLINT     NOT NULL DEFAULT 0,   -- '[必选] 邮箱是否验证'
    deleted_at         BIGINT,                            -- '[可选] 删除时间，毫秒时间戳'
    updated_at         BIGINT       NOT NULL DEFAULT 0,   -- '[必选] 更新时间，毫秒时间戳'
    created_at         BIGINT       NOT NULL DEFAULT 0,   -- '[必选] 创建时间，毫秒时间戳'
    ext                JSONB                 DEFAULT '{}' -- '[可选] 扩展字段'
);
CREATE INDEX idx_basic_number ON user_basic (number);
CREATE INDEX idx_basic_phone ON user_basic (phone);
CREATE INDEX idx_basic_email ON user_basic (email);
CREATE INDEX idx_user_ext ON user_basic USING GIN (ext jsonb_path_ops);
CREATE INDEX idx_basic_admin_region ON user_basic (country_code, province_code, city_code, district_code);

CREATE TABLE user_auth
(
    id         BIGSERIAL PRIMARY KEY,
    uid        BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 用户id'
    type       SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 认证类型枚举'
    status     SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 状态枚举'
    appid      VARCHAR(50)  NOT NULL DEFAULT '', -- '[必选] appid，如微信小程序的AppId'
    identifier VARCHAR(100) NOT NULL DEFAULT '', -- '[必选] 认证id, openid等'
    credential VARCHAR(128) NOT NULL DEFAULT '', -- '[必选] 认证密文，session_key等'
    union_id   VARCHAR(100),                     -- '[可选] 微信登录的unionid'
    deleted_at BIGINT,                           -- '[可选] 删除时间，毫秒时间戳'
    updated_at BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 更新时间，毫秒时间戳'
    created_at BIGINT       NOT NULL DEFAULT 0   -- '[必选] 创建时间，毫秒时间戳'
);
CREATE INDEX idx_auth_uid ON user_auth (uid);
CREATE INDEX idx_auth_appid ON user_auth (appid);
CREATE INDEX idx_auth_identifier ON user_auth (identifier);

CREATE TABLE user_sign_log
(
    id                 BIGSERIAL PRIMARY KEY,
    uid                BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 用户id'
    type               SMALLINT    NOT NULL DEFAULT 0,  -- '[必选] 认证类型枚举'
    status             SMALLINT    NOT NULL DEFAULT 0,  -- '[必选] 状态枚举'
    identifier         VARCHAR(50) NOT NULL DEFAULT '', -- '[必选] 登录的账号信息appid 邮箱 手机号等'
    ip                 VARCHAR(128),                    -- '[可选] 登录ip'
    location           VARCHAR(100),                    -- '[可选] 位置'
    agent              VARCHAR(255),                    -- '[可选] 登录软件信息'
    device             VARCHAR(255),                    -- '[可选] 登录设备信息'
    access_session_id  CHAR(36)    NOT NULL DEFAULT '', -- '[必选] access_session_id'
    refresh_session_id CHAR(36)    NOT NULL DEFAULT '', -- '[必选] refresh_session_id'
    access_expired_at  BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 过期时间，毫秒时间戳'
    refresh_expired_at BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 刷新截止时间，毫秒时间戳'
    deleted_at         BIGINT,                          -- '[可选] 删除时间，毫秒时间戳'
    updated_at         BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 更新时间，毫秒时间戳'
    created_at         BIGINT      NOT NULL DEFAULT 0   -- '[必选] 创建时间，毫秒时间戳'
);
CREATE INDEX idx_sign_log_uid ON user_sign_log (uid);
CREATE INDEX idx_sign_log_access_session_id ON user_sign_log (access_session_id);
CREATE INDEX idx_sign_log_refresh_session_id ON user_sign_log (refresh_session_id);
