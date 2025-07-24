CREATE TABLE message_sms
(
    id           BIGSERIAL PRIMARY KEY,
    msg_type     SMALLINT      NOT NULL DEFAULT 0,  -- '[必选] 消息类型'
    msg_status   SMALLINT      NOT NULL DEFAULT 0,  -- '[必选] 消息状态'
    captcha_type SMALLINT,                          -- '[可选] 验证码类型'
    provider     INT           NOT NULL DEFAULT 0,  -- '[必选] 供应商'
    template     VARCHAR(20)   NOT NULL DEFAULT '', -- '[必选] 短信模板'
    sign         VARCHAR(20)   NOT NULL DEFAULT '', -- '[必选] 短信签名'
    request_id   VARCHAR(50),                       -- '[必选] 供应商请求id'
    biz_id       VARCHAR(50),                       -- '[可选] 流水id'
    phone        VARCHAR(20)   NOT NULL DEFAULT '', -- '[必选] 手机号'
    sender_id    BIGINT        NOT NULL DEFAULT 0,  -- '[必选] 发送人id'
    params       VARCHAR(1000) NOT NULL DEFAULT '', -- '[必选] 发送模板参数'
    content      VARCHAR(1000),                     -- '[可选] 短信内容'
    err_code     VARCHAR(50),                       -- '[可选] 错误代码'
    err_msg      TEXT,                              -- '[可选] 错误信息'
    send_date    VARCHAR(50),                       -- '[必选] 短信发送时间'
    receive_date VARCHAR(50),                       -- '[必选] 短信接收时间'
    deleted_at   BIGINT,                            -- '[可选] 删除时间'
    updated_at   BIGINT        NOT NULL DEFAULT 0,  -- '[必选] 更新时间'
    created_at   BIGINT        NOT NULL DEFAULT 0   -- '[必选] 创建时间'
);
CREATE INDEX idx_phone ON message_sms (phone);
CREATE INDEX idx_biz_id ON message_sms(biz_id);
CREATE INDEX idx_sender_id ON message_sms (sender_id);
CREATE INDEX idx_msg_status_created_at ON message_sms (msg_status, created_at);