package audit

const (
	DefaultMaxDetailsBytes = 4 * 1024
	DefaultPageSize        = 50
	DefaultMaxPageSize     = 200
)

type Configuration struct {
	MaxDetailsBytes int
	DefaultPageSize int
	MaxPageSize     int
}

func NewConfiguration() *Configuration {
	return &Configuration{
		MaxDetailsBytes: DefaultMaxDetailsBytes,
		DefaultPageSize: DefaultPageSize,
		MaxPageSize:     DefaultMaxPageSize,
	}
}

func (c *Configuration) normalizePageLimit(limit int) int {
	if c == nil {
		c = NewConfiguration()
	}

	if limit <= 0 {
		limit = c.DefaultPageSize
	}

	if c.MaxPageSize > 0 && limit > c.MaxPageSize {
		return c.MaxPageSize
	}

	return limit
}
