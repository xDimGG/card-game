package main

type PacketType string

const (
	PacketTypeMessage PacketType = "message"
	PacketTypeError   PacketType = "error"
)

type Packet struct {
	Type  PacketType             `json:"type"`
	Moves []string               `json:"moves"`
	Game  string                 `json:"game"`
	State interface{}            `json:"state"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

func NewPacketError(message string) *Packet {
	return &Packet{
		Type: PacketTypeError,
		Data: map[string]interface{}{
			"message": message,
		},
	}
}

func Get[T any](data interface{}, key string) (T, bool) {
	obj, ok := data.(map[string]interface{})
	if ok {
		val, ok := obj[key]
		if ok {
			if cast, ok := val.(T); ok {
				return cast, true
			}
		}
	}

	return *new(T), false
}
