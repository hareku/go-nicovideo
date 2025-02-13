package nicovideo

type Client struct {
	userSession string
}

func NewClient(opts ...NewClientOption) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type NewClientOption func(c *Client)

func WithUserSession(userSession string) NewClientOption {
	return func(c *Client) {
		c.userSession = userSession
	}
}
