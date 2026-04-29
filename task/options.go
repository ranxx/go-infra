package task

// Options options
type Options struct {
	MaxRetries int
}

// Option options
type Option func(*Options)

// WithMaxRetries with max retries
func WithMaxRetries(maxRetries int) Option {
	return func(o *Options) {
		o.MaxRetries = maxRetries
	}
}

func defaultOptions() Options {
	return Options{
		MaxRetries: 10,
	}
}
