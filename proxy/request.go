package proxy

type Request struct {
	Action  string
	Payload Payload
	Token   string
	Client  Client
}

type Payload struct {
	TxtFQDN  string
	TxtValue string
}

type Client struct {
	RemoteAddr string
	Name       string
}
