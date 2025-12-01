package captcha

import (
	"context"
	"fmt"
	"time"

	"github.com/byteflowing/base/pkg/redis"
	"github.com/byteflowing/base/pkg/utils/idx"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

const (
	tokenKeyFormat = "%s:%d:%d:{%s}:%s" // "prefix:sender:scene:{target}:token"
)

type Config struct {
	Prefix              string
	MaxTries            int
	Length              int
	CaptchaTTL          time.Duration
	CaptchaCombinations []enumv1.CaptchaTypeMask
}

type MessageCaptcha struct {
	cfg      *Config
	verifier *ValueVerifier
}

func NewMessageCaptcha(rdb *redis.Redis, cfg *Config) *MessageCaptcha {
	v := NewValueVerifier(rdb, cfg.CaptchaTTL, cfg.MaxTries)
	return &MessageCaptcha{
		cfg:      cfg,
		verifier: v,
	}
}

// Save : 生成验证码并存储在redis中，返回一个token
func (c *MessageCaptcha) Save(ctx context.Context, target string, sender enumv1.MessageSenderType, scene enumv1.MessageSceneType) (code, token string, err error) {
	code = GenerateCaptcha(c.cfg.Length, c.cfg.CaptchaCombinations)
	token = idx.UUIDv4()
	key := c.getKey(target, token, sender, scene)
	if err := c.verifier.Save(ctx, key, code); err != nil {
		return "", "", err
	}
	return
}

// Verify 带着Save接口返回的token及收到的验证码value来验证，验证成功以及达到最大尝试次数会删除验证码
func (c *MessageCaptcha) Verify(ctx context.Context, target, token, code string, sender enumv1.MessageSenderType, scene enumv1.MessageSceneType) (result *ValueVerifyResult, err error) {
	key := c.getKey(target, token, sender, scene)
	return c.verifier.Verify(ctx, key, code)
}

func (c *MessageCaptcha) getKey(target, token string, sender enumv1.MessageSenderType, scene enumv1.MessageSceneType) string {
	return fmt.Sprintf(tokenKeyFormat, c.cfg.Prefix, sender, scene, target, token)
}
