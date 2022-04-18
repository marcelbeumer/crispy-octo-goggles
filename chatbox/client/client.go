package client

type Client struct{}

func (s *Client) Connect(serverAddr string) error {
	// r := s.initRouting()
	// err := http.ListenAndServe(addr, r)
	return nil
}

func NewClient() *Client {
	s := Client{}
	return &s
}
