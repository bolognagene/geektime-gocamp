package key_expired_event

type KeyExpiredEvent interface {
	Process(key string) error
}
