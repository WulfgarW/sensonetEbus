package sensonetEbus

type ConnOption func(*Connection)

func WithLogger(logger Logger) ConnOption {
	return func(c *Connection) {
		c.logger = logger
	}
}

type EbusConnOption func(*EbusConnection)

func WithConnLogger(logger Logger) EbusConnOption {
	return func(c *EbusConnection) {
		c.logger = logger
	}
}
