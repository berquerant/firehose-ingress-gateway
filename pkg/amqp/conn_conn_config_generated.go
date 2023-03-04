// Code generated by "goconfig -field ExchangeName string|ExchangeKind string -prefix Conn -option -output conn_conn_config_generated.go"; DO NOT EDIT.

package amqp

type ConnConfigItem[T any] struct {
	modified     bool
	value        T
	defaultValue T
}

func (s *ConnConfigItem[T]) Set(value T) {
	s.modified = true
	s.value = value
}
func (s *ConnConfigItem[T]) Get() T {
	if s.modified {
		return s.value
	}
	return s.defaultValue
}
func (s *ConnConfigItem[T]) Default() T {
	return s.defaultValue
}
func (s *ConnConfigItem[T]) IsModified() bool {
	return s.modified
}
func NewConnConfigItem[T any](defaultValue T) *ConnConfigItem[T] {
	return &ConnConfigItem[T]{
		defaultValue: defaultValue,
	}
}

type ConnConfig struct {
	ExchangeName *ConnConfigItem[string]
	ExchangeKind *ConnConfigItem[string]
}
type ConnConfigBuilder struct {
	exchangeName string
	exchangeKind string
}

func (s *ConnConfigBuilder) ExchangeName(v string) *ConnConfigBuilder {
	s.exchangeName = v
	return s
}
func (s *ConnConfigBuilder) ExchangeKind(v string) *ConnConfigBuilder {
	s.exchangeKind = v
	return s
}
func (s *ConnConfigBuilder) Build() *ConnConfig {
	return &ConnConfig{
		ExchangeName: NewConnConfigItem(s.exchangeName),
		ExchangeKind: NewConnConfigItem(s.exchangeKind),
	}
}

func NewConnConfigBuilder() *ConnConfigBuilder { return &ConnConfigBuilder{} }
func (s *ConnConfig) Apply(opt ...ConnConfigOption) {
	for _, x := range opt {
		x(s)
	}
}

type ConnConfigOption func(*ConnConfig)

func WithExchangeName(v string) ConnConfigOption {
	return func(c *ConnConfig) {
		c.ExchangeName.Set(v)
	}
}
func WithExchangeKind(v string) ConnConfigOption {
	return func(c *ConnConfig) {
		c.ExchangeKind.Set(v)
	}
}
