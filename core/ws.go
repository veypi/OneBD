package core

type WebSocketFunc = func(m Meta) (WebSocketConn, error)

type WebSocketConn interface {
	Wait()
	Disconnect(error) error
}
