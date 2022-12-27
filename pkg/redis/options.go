package redis

var DefaultOption = option{
	addr: "localhost:6379",
	db:   0,
}

type option struct {
	addr     string
	password string
	db       int
}

type optionFn func(*option)

func NewOption(opts ...optionFn) *option {
	opt := DefaultOption
	for _, item := range opts {
		item(&opt)
	}
	return &opt
}

func WithAddr(addr string) optionFn {
	return func(o *option) {
		o.addr = addr
	}
}

func WithPassword(password string) optionFn {
	return func(o *option) {
		o.password = password
	}
}

func WithDB(db int) optionFn {
	return func(o *option) {
		o.db = db
	}
}
