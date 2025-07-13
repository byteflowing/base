CREATE TABLE message_sms
(
    id                  BIGSERIAL PRIMARY KEY,
    msg_type            SMALLINT    NOT NULL DEFAULT 0,  -- '[必选] 消息类型'
    msg_status          SMALLINT    NOT NULL DEFAULT 0,  -- '[必选] 消息状态'
    captcha_type        SMALLINT,                        -- '[可选] 验证码类型'
    captcha_combination SMALLINT,                        -- '[可选] 验证码组合形式'
    provider            INT         NOT NULL DEFAULT 0,  -- '[必选] 供应商'
    template            VARCHAR(20) NOT NULL DEFAULT '', -- '[必选] 短信模板'
    sign                VARCHAR(20) NOT NULL DEFAULT '', -- '[必选] 短信签名'
    request_id          VARCHAR(50) NOT NULL DEFAULT '', -- '[必选] 供应商请求id'
    phone               VARCHAR(20) NOT NULL DEFAULT '', -- '[必选] 手机号'
    sender_id           BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 发送人id'
    msg                 VARCHAR(1000),                   -- '[可选] 短信内容'
    letter_count        INT         NOT NULL DEFAULT 0,  -- '[必选] 短信内容长度'
    msg_count           INT         NOT NULL DEFAULT 0,  -- '[必选] 短信占用条数'
    send_time           BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 短信发送时间'
    receive_time        BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 短信接收时间'
    deleted_at          BIGINT,                          -- '[可选] 删除时间'
    updated_at          BIGINT      NOT NULL DEFAULT 0,  -- '[必选] 更新时间'
    created_at          BIGINT      NOT NULL DEFAULT 0   -- '[必选] 创建时间'
);
CREATE INDEX idx_request_id ON message_sms(request_id);
CREATE INDEX idx_phone ON message_sms(phone);
CREATE INDEX idx_sender_id ON message_sms(sender_id);
CREATE INDEX idx_created_at ON message_sms(created_at);