package logger

import (
	"crypto/sha1"
	"encoding/hex"
	"go.uber.org/zap"
	"hash"
	"strings"
)

type ZapLogger struct {
	l            *zap.Logger
	h            hash.Hash
	allowEncrypt bool
}

func NewZapLogger(l *zap.Logger, encrypt bool) *ZapLogger {
	return &ZapLogger{
		l:            l,
		h:            sha1.New(),
		allowEncrypt: encrypt,
	}
}

func (z *ZapLogger) Debug(msg string, args ...Field) {
	z.l.Debug(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Info(msg string, args ...Field) {
	z.l.Info(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Warn(msg string, args ...Field) {
	z.l.Warn(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Error(msg string, args ...Field) {
	z.l.Error(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) toZapFields(args []Field) []zap.Field {
	res := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		// 脱敏，碰到phone, password, email进行加密
		// https://md5decrypt.net/en/Sha1/
		if z.allowEncrypt && (strings.EqualFold(arg.Key, "phone") ||
			strings.EqualFold(arg.Key, "email") ||
			strings.Contains(strings.ToLower(arg.Key), "password")) {
			z.h.Write([]byte(arg.Value.(string)))
			arg.Value = hex.EncodeToString(z.h.Sum(nil))
		}
		res = append(res, zap.Any(arg.Key, arg.Value))
	}
	return res
}
