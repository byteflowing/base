CREATE TABLE user_basic
(
    id                  BIGINT PRIMARY KEY,
    biz                 VARCHAR(50)  NOT NULL DEFAULT '', -- '[必选] 用于隔离不同应用'
    number              VARCHAR(50)  NOT NULL DEFAULT '', -- '[必选] 用户编号'
    name                VARCHAR(50),                      -- '[可选] 用户名称'
    alias               VARCHAR(50),                      -- '[可选] 昵称'
    password            VARCHAR(100),                     -- '[可选] 密码'
    avatar              VARCHAR(255),                     -- '[可选] 头像'
    gender              SMALLINT,                         -- '[可选] 性别'
    birthday            DATE,                             -- '[可选] 生日'
    phone_country_code  VARCHAR(10)  NOT NULL DEFAULT '', -- '[必选] 手机国家编码'
    phone               VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 手机号码'
    email               VARCHAR(100) NOT NULL DEFAULT '', -- '[必选] 邮箱'
    country_code        VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 国家编码'
    province_code       VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 省份编码'
    city_code           VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 城市编码'
    district_code       VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 区县编码'
    addr                VARCHAR(255),                     -- '[可选] 详细地址'
    status              SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 用户状态枚举'
    source              SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 用户注册来源'
    signup_type         SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 用户注册类型'
    phone_verified      SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 手机是否验证'
    email_verified      SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 邮箱是否验证'
    type                SMALLINT,                         -- '[可选] 用户类型，预留字段，由业务方维护'
    level               INT,                              -- '[可选] 用户等级，预留字段，由业务方维护'
    register_ip         VARCHAR(50),                      -- '[可选] 注册ip'
    register_device     VARCHAR(255),                     -- '[可选] 注册设备'
    register_agent      VARCHAR(255),                     -- '[可选] 注册UA'
    password_updated_at BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 密码更改时间，毫秒时间戳'
    deleted_at          BIGINT,                           -- '[可选] 删除时间，毫秒时间戳'
    updated_at          BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 更新时间，毫秒时间戳'
    created_at          BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 创建时间，毫秒时间戳'
    ext                 JSON                              -- '[可选] 扩展字段'
);
CREATE UNIQUE INDEX idx_basic_uni_number ON user_basic (number);
CREATE UNIQUE INDEX idx_basic_uni_biz_phone ON user_basic (biz, phone_country_code, phone);
CREATE UNIQUE INDEX idx_basic_uni_biz_email ON user_basic (biz, email);
CREATE INDEX idx_basic_admin_region ON user_basic (country_code, province_code, city_code, district_code);

CREATE TABLE user_auth
(
    id         BIGSERIAL PRIMARY KEY,
    uid        BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 用户id'
    type       SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 认证类型枚举'
    status     SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 状态枚举'
    appid      VARCHAR(50)  NOT NULL DEFAULT '', -- '[必选] appid，如微信小程序的AppId'
    biz        VARCHAR(50)  NOT NULL DEFAULT '', -- '[必选] 用于隔离不同应用'
    identifier VARCHAR(100) NOT NULL DEFAULT '', -- '[必选] 认证id, openid等'
    credential VARCHAR(128) NOT NULL DEFAULT '', -- '[必选] 认证密文，session_key等'
    union_id   VARCHAR(100),                     -- '[可选] 微信登录的unionid'
    deleted_at BIGINT,                           -- '[可选] 删除时间，毫秒时间戳'
    updated_at BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 更新时间，毫秒时间戳'
    created_at BIGINT       NOT NULL DEFAULT 0   -- '[必选] 创建时间，毫秒时间戳'
);
CREATE INDEX idx_auth_uid ON user_auth (uid);
CREATE UNIQUE INDEX idx_auth_uni_appid ON user_auth (appid);
CREATE INDEX idx_auth_identifier ON user_auth (identifier);
CREATE INDEX idx_auth_union_id ON user_auth (union_id);

CREATE TABLE user_sign_log
(
    id                 BIGSERIAL PRIMARY KEY,
    uid                BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 用户id'
    type               SMALLINT    NOT NULL DEFAULT 0,  -- '[必选] 认证类型枚举'
    status             SMALLINT    NOT NULL DEFAULT 0,  -- '[必选] 状态枚举'
    identifier         VARCHAR(50) NOT NULL DEFAULT '', -- '[必选] 登录的账号信息appid 邮箱 手机号等'
    biz                VARCHAR(50) NOT NULL DEFAULT '', -- '[必选] 用于隔离不同应用'
    ip                 VARCHAR(50),                     -- '[可选] 登录ip'
    location           VARCHAR(100),                    -- '[可选] 位置'
    agent              VARCHAR(255),                    -- '[可选] 登录软件信息'
    device             VARCHAR(255),                    -- '[可选] 登录设备信息'
    access_session_id  CHAR(36)    NOT NULL DEFAULT '', -- '[必选] access_session_id'
    refresh_session_id CHAR(36)    NOT NULL DEFAULT '', -- '[必选] refresh_session_id'
    access_expired_at  BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 过期时间，秒时间戳'
    refresh_expired_at BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 刷新截止时间，秒时间戳'
    deleted_at         BIGINT,                          -- '[可选] 删除时间，毫秒时间戳'
    updated_at         BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 更新时间，毫秒时间戳'
    created_at         BIGINT      NOT NULL DEFAULT 0   -- '[必选] 创建时间，毫秒时间戳'
);
CREATE INDEX idx_sign_log_uid ON user_sign_log (uid);
CREATE UNIQUE INDEX idx_sign_uni_log_access_session_id ON user_sign_log (access_session_id);
CREATE UNIQUE INDEX idx_sign_uni_log_refresh_session_id ON user_sign_log (refresh_session_id);
