package core

import (
	"encoding/json"
	"time"
)

// Client is the subset of the Huginn C ABI exposed by the Bot API.
type Client interface {
	Close() error
	GetMe() (json.RawMessage, error)
	GetPeers(query string) (json.RawMessage, error)
	GetMessages(chatID string, limit, offset int) (json.RawMessage, error)
	SendMessage(chatID, text string, ttlSeconds int) (json.RawMessage, error)
	SendFile(chatID, caption, path string, ttlSeconds int) (json.RawMessage, error)
	GetUpdate(timeout time.Duration) (json.RawMessage, error)
	GetGroups() (json.RawMessage, error)
	CreateGroup(name string) (json.RawMessage, error)
	InviteToGroup(groupID, userID string) (json.RawMessage, error)
	MarkRead(messageID string) (json.RawMessage, error)
}

type Options struct {
	LibraryPath string
	Username    string
	MuninnAddr  string
	Database    string
	ChunkTTL    string
	TURNAddr    string
	TURNUser    string
	TURNPass    string
}
