package filtersegments

func NewTestClient(client client) *Client {
	return &Client{client: client}
}
