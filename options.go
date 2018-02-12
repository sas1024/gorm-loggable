package loggable

type options struct {
	lazyUpdate       bool
	lazyUpdateFields []string
}

func LazyUpdateOption(fields ...string) func(options *options) {
	return func(options *options) {
		options.lazyUpdate = true
		options.lazyUpdateFields = fields
	}
}
