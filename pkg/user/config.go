package user

type LimitRule struct {
	Prefix   string // redis key前缀
	Duration uint64 // 时间间隔单位s
	Limit    uint32 // 时间间隔内的最大数量
}

type BlockListConfig struct {
	BlockListPrefix string
	SignInLimit     *LimitRule
	RefreshLimit    *LimitRule
}

type Config struct {
	BlockList *BlockListConfig
}
