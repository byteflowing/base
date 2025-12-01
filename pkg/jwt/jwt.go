package jwt

import (
	"time"

	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/utils/idx"
	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenTypeKey = "token_type"
)

type Jwt struct {
	issuer     string
	secretKey  string
	signMethod jwt.SigningMethod
}

type Token struct {
	Token  string
	Jti    string
	Exp    time.Time
	Claims jwt.MapClaims
}

func New(issuer, secret string) *Jwt {
	return &Jwt{
		issuer:     issuer,
		secretKey:  secret,
		signMethod: jwt.GetSigningMethod("HS256"),
	}
}

// Generate 签发token
// @param subject jwt claims中sub的值
// @param ttl jwt的有效时长
// @param extra 自定义claims
// token中claim字段
//
//	iss: 签发人
//	sub: subject
//	exp: 过期时间
//	iat: 签发时间
//	nbf: 生效时间
//	jti: token对应的uuid可作为sessionID
//	token_type: 生成token的类型，业务自定义
//	其他：extra中传递自定义字段
func (j *Jwt) Generate(subject, tokenType string, ttl time.Duration, extra map[string]any) (*Token, error) {
	now := jwt.NewNumericDate(time.Now())
	jti, err := idx.UUIDv7()
	if err != nil {
		return nil, err
	}
	claims := jwt.MapClaims{
		"iss":        j.issuer,
		"sub":        subject,
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
		"jti":        jti,
		TokenTypeKey: tokenType,
	}
	for k, v := range extra {
		claims[k] = v
	}
	signed, err := jwt.NewWithClaims(j.signMethod, claims).SignedString(j.secretKey)
	return &Token{
		Token:  signed,
		Jti:    jti,
		Exp:    now.Add(ttl),
		Claims: claims,
	}, nil
}

// Parse 解析token，若err不为nil则token验证失败
// token中claim字段
//
//	iss: 签发人
//	sub: subject
//	exp: 过期时间
//	iat: 签发时间
//	nbf: 生效时间
//	jti: token对应的uuid可作为sessionID
//	其他：签发token时传递extra中的字段
func (j *Jwt) Parse(tokenString, tokenType string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != j.signMethod.Alg() {
			return nil, ecode.ErrJwtSignMethodMismatch
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		if iss, _ := claims.GetIssuer(); iss != j.issuer {
			return nil, ecode.ErrJwtIssuerMismatch
		}
		if j.getTokenType(claims) != tokenType {
			return nil, ecode.ErrJwtTokenTypeMismatch
		}
		return claims, nil
	}
	return nil, ecode.ErrJwtInvalidToken
}

func (j *Jwt) getTokenType(claims jwt.MapClaims) string {
	tokenType, _ := claims[TokenTypeKey].(string)
	return tokenType
}
