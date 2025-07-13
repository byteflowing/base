CREATE TABLE user_basic
(
    id             BIGSERIAL PRIMARY KEY,
    number         VARCHAR(50),                      -- '[必选] 用户编号'
    name           VARCHAR(50)  NOT NULL DEFAULT '', -- '[必选] 用户名'
    alias          VARCHAR(50)  NOT NULL DEFAULT '', -- '[必选] 昵称'
    password       VARCHAR(128),                     -- '[可选] 密码'
    avatar         VARCHAR(255),                     -- '[可选] 头像'
    gender         SMALLINT,                         -- '[可选] 性别'
    phone          VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 手机号码'
    email          VARCHAR(100) NOT NULL DEFAULT '', -- '[必选] 邮箱'
    country        VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 国家编码'
    province       VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 省份编码'
    city           VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 城市编码'
    district       VARCHAR(20)  NOT NULL DEFAULT '', -- '[必选] 区县编码'
    addr           VARCHAR(255),                     -- '[可选] 详细地址'
    status         SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 用户状态枚举'
    source         SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 用户注册来源'
    level          SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 用户等级'
    phone_verified SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 手机是否验证'
    email_verified SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 邮箱是否验证'
    deleted_at     BIGINT,                           -- '[可选] 删除时间'
    updated_at     BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 更新时间'
    created_at     BIGINT       NOT NULL DEFAULT 0   -- '[必选] 创建时间'
);
CREATE INDEX idx_basic_number ON user_basic (number);
CREATE INDEX idx_basic_name ON user_basic (name);
CREATE INDEX idx_basic_phone ON user_basic (phone);
CREATE INDEX idx_basic_email ON user_basic (email);
CREATE INDEX idx_basic_admin_region ON user_basic (country, province, city, district);

CREATE TABLE user_auth
(
    id         BIGSERIAL PRIMARY KEY,
    uid        BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 用户id'
    type       SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 认证类型枚举'
    status     SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 状态枚举'
    identifier VARCHAR(100) NOT NULL DEFAULT '', -- '[必选] 认证id'
    token      VARCHAR(128) NOT NULL DEFAULT '', -- '[必选] 认证密文'
    deleted_at BIGINT,                           -- '[可选] 删除时间'
    updated_at BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 更新时间'
    created_at BIGINT       NOT NULL DEFAULT 0   -- '[必选] 创建时间'
);
CREATE INDEX idx_auth_uid ON user_auth (uid);
CREATE INDEX idx_auth_identifier ON user_auth (identifier);

CREATE TABLE user_login_log
(
    id         BIGSERIAL PRIMARY KEY,
    uid        BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 用户id'
    session_id VARCHAR(128) NOT NULL DEFAULT '', -- '[必选] session_id'
    type       SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 认证类型枚举'
    status     SMALLINT     NOT NULL DEFAULT 0,  -- '[必选] 状态枚举'
    ip         VARCHAR(128),                     -- '[可选] 登录ip'
    location   VARCHAR(100),                     -- '[可选] 位置'
    agent      VARCHAR(100),                     -- '[可选] 登录软件信息'
    device     VARCHAR(100),                     -- '[可选] 登录设备信息'
    deleted_at BIGINT,                           -- '[可选] 删除时间'
    updated_at BIGINT       NOT NULL DEFAULT 0,  -- '[必选] 更新时间'
    created_at BIGINT       NOT NULL DEFAULT 0   -- '[必选] 创建时间'
);
CREATE INDEX idx_login_log_uid ON user_login_log (uid);
CREATE INDEX idx_login_log_session_id ON user_login_log (session_id);