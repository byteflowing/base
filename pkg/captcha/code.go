package captcha

import (
	"strings"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/bytedance/gopkg/lang/stringx"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

var (
	numberCharset string
	lowerCharset  string
	upperCharset  string
	symbolCharset string
)

func init() {
	numberCharset = stringx.Shuffle("0123456789")
	lowerCharset = stringx.Shuffle("abcdefghijklmnopqrstuvwxyz")
	upperCharset = stringx.Shuffle("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	symbolCharset = stringx.Shuffle("!@#$%&*")
}

func GenerateCaptcha(length int, captchaMask []enumv1.CaptchaTypeMask) string {
	if len(captchaMask) == 0 {
		captchaMask = []enumv1.CaptchaTypeMask{enumv1.CaptchaTypeMask_CAPTCHA_TYPE_MASK_NUMBER}
	}
	var mask enumv1.CaptchaTypeMask
	for _, c := range captchaMask {
		mask = mask | c
	}
	captchaType := int32(mask)
	var pools []string
	if captchaType&int32(enumv1.CaptchaTypeMask_CAPTCHA_TYPE_MASK_NUMBER) != 0 {
		pools = append(pools, numberCharset)
	}
	if captchaType&int32(enumv1.CaptchaTypeMask_CAPTCHA_TYPE_MASK_UPPERCASE) != 0 {
		pools = append(pools, upperCharset)
	}
	if captchaType&int32(enumv1.CaptchaTypeMask_CAPTCHA_TYPE_MASK_LOWERCASE) != 0 {
		pools = append(pools, lowerCharset)
	}
	if captchaType&int32(enumv1.CaptchaTypeMask_CAPTCHA_TYPE_MASK_SYMBOL) != 0 {
		pools = append(pools, symbolCharset)
	}
	if len(pools) == 0 {
		pools = append(pools, numberCharset)
	}
	var code []byte
	poolLen := len(pools)
	for _, pool := range pools {
		code = append(code, pool[fastrand.Intn(poolLen)])
	}

	allCharset := strings.Join(pools, "")
	allCharsetsLen := len(allCharset)
	codeLen := len(code)
	for codeLen < length {
		code = append(code, allCharset[fastrand.Intn(allCharsetsLen)])
	}
	return stringx.Shuffle(string(code))
}
