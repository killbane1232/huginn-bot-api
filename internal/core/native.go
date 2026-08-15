package core

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <string.h>

typedef long (*create_fn)(char*, char*, char*, char*, char*, char*, char*);
typedef long (*create_with_peer_flag_fn)(char*, char*, char*, char*, char*, char*, char*, char*);
typedef void (*destroy_fn)(long);
typedef char* (*get_me_fn)(long);
typedef char* (*get_peers_fn)(long);
typedef char* (*search_peers_fn)(long, char*);
typedef char* (*get_messages_paginated_fn)(long, char*, int, int);
typedef char* (*send_message_fn)(long, char*, char*, int);
typedef char* (*send_file_fn)(long, char*, char*, char*, int);
typedef char* (*get_event_fn)(long, int);
typedef char* (*get_groups_fn)(long);
typedef char* (*create_group_fn)(long, char*);
typedef char* (*invite_to_group_fn)(long, char*, char*);
typedef char* (*mark_read_fn)(long, char*);
typedef void (*free_string_fn)(char*);

typedef struct {
	void *library;
	create_fn create;
	create_with_peer_flag_fn create_with_peer_flag;
	destroy_fn destroy;
	get_me_fn get_me;
	get_peers_fn get_peers;
	search_peers_fn search_peers;
	get_messages_paginated_fn get_messages_paginated;
	send_message_fn send_message;
	send_file_fn send_file;
	get_event_fn get_event;
	get_groups_fn get_groups;
	create_group_fn create_group;
	invite_to_group_fn invite_to_group;
	mark_read_fn mark_read;
	free_string_fn free_string;
} huginn_api;

static void api_fail(huginn_api *api, char **error, const char *message) {
	*error = strdup(message == NULL ? "unknown dynamic loader error" : message);
	if (api != NULL) {
		if (api->library != NULL) dlclose(api->library);
		free(api);
	}
}

#define LOAD_SYMBOL(api, field, symbol, type, error) do { \
	dlerror(); \
	(api)->field = (type)dlsym((api)->library, symbol); \
	const char *load_error = dlerror(); \
	if (load_error != NULL) { api_fail((api), (error), load_error); return NULL; } \
} while (0)

static huginn_api* api_open(const char *path, char **error) {
	huginn_api *api = (huginn_api*)calloc(1, sizeof(huginn_api));
	if (api == NULL) {
		*error = strdup("out of memory");
		return NULL;
	}
	api->library = dlopen(path, RTLD_NOW | RTLD_LOCAL);
	if (api->library == NULL) {
		api_fail(api, error, dlerror());
		return NULL;
	}
	LOAD_SYMBOL(api, create, "messenger_create", create_fn, error);
	LOAD_SYMBOL(api, create_with_peer_flag, "messenger_create_with_peer_flag", create_with_peer_flag_fn, error);
	LOAD_SYMBOL(api, destroy, "messenger_destroy", destroy_fn, error);
	LOAD_SYMBOL(api, get_me, "messenger_get_me", get_me_fn, error);
	LOAD_SYMBOL(api, get_peers, "messenger_get_peers", get_peers_fn, error);
	LOAD_SYMBOL(api, search_peers, "messenger_search_peers", search_peers_fn, error);
	LOAD_SYMBOL(api, get_messages_paginated, "messenger_get_messages_paginated", get_messages_paginated_fn, error);
	LOAD_SYMBOL(api, send_message, "messenger_send_message", send_message_fn, error);
	LOAD_SYMBOL(api, send_file, "messenger_send_file", send_file_fn, error);
	LOAD_SYMBOL(api, get_event, "messenger_get_event", get_event_fn, error);
	LOAD_SYMBOL(api, get_groups, "messenger_get_groups", get_groups_fn, error);
	LOAD_SYMBOL(api, create_group, "messenger_create_group", create_group_fn, error);
	LOAD_SYMBOL(api, invite_to_group, "messenger_invite_to_group", invite_to_group_fn, error);
	LOAD_SYMBOL(api, mark_read, "messenger_mark_read", mark_read_fn, error);
	LOAD_SYMBOL(api, free_string, "messenger_free_string", free_string_fn, error);
	return api;
}

static void api_close(huginn_api *api) {
	if (api == NULL) return;
	if (api->library != NULL) dlclose(api->library);
	free(api);
}

static long api_create(huginn_api *api, char *username, char *muninn, char *db, char *ttl, char *turn_addr, char *turn_user, char *turn_pass) {
	return api->create(username, muninn, db, ttl, turn_addr, turn_user, turn_pass);
}
static long api_create_with_peer_flag(huginn_api *api, char *username, char *muninn, char *db, char *ttl, char *turn_addr, char *turn_user, char *turn_pass, char *peer_flag) {
	return api->create_with_peer_flag(username, muninn, db, ttl, turn_addr, turn_user, turn_pass, peer_flag);
}
static void api_destroy(huginn_api *api, long handle) { api->destroy(handle); }
static char* api_get_me(huginn_api *api, long handle) { return api->get_me(handle); }
static char* api_get_peers(huginn_api *api, long handle) { return api->get_peers(handle); }
static char* api_search_peers(huginn_api *api, long handle, char *query) { return api->search_peers(handle, query); }
static char* api_get_messages_paginated(huginn_api *api, long handle, char *chat, int limit, int offset) { return api->get_messages_paginated(handle, chat, limit, offset); }
static char* api_send_message(huginn_api *api, long handle, char *chat, char *text, int ttl) { return api->send_message(handle, chat, text, ttl); }
static char* api_send_file(huginn_api *api, long handle, char *chat, char *caption, char *path, int ttl) { return api->send_file(handle, chat, caption, path, ttl); }
static char* api_get_event(huginn_api *api, long handle, int timeout_ms) { return api->get_event(handle, timeout_ms); }
static char* api_get_groups(huginn_api *api, long handle) { return api->get_groups(handle); }
static char* api_create_group(huginn_api *api, long handle, char *name) { return api->create_group(handle, name); }
static char* api_invite_to_group(huginn_api *api, long handle, char *group, char *user) { return api->invite_to_group(handle, group, user); }
static char* api_mark_read(huginn_api *api, long handle, char *message) { return api->mark_read(handle, message); }
static void api_free_string(huginn_api *api, char *value) { api->free_string(value); }
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

const botPeerFlag = "very_thick"

type Native struct {
	mu     sync.RWMutex
	api    *C.huginn_api
	handle C.long
	closed bool
}

func Open(options Options) (*Native, error) {
	if options.LibraryPath == "" {
		return nil, errors.New("core library path is required")
	}
	if options.Username == "" {
		return nil, errors.New("HUGINN_USERNAME is required")
	}

	path := C.CString(options.LibraryPath)
	defer C.free(unsafe.Pointer(path))
	var loadError *C.char
	api := C.api_open(path, &loadError)
	if api == nil {
		defer C.free(unsafe.Pointer(loadError))
		return nil, fmt.Errorf("load Huginn core: %s", C.GoString(loadError))
	}

	values := []string{options.Username, options.MuninnAddr, options.Database, options.ChunkTTL, options.TURNAddr, options.TURNUser, options.TURNPass, botPeerFlag}
	cvalues := make([]*C.char, len(values))
	for i, value := range values {
		cvalues[i] = C.CString(value)
		defer C.free(unsafe.Pointer(cvalues[i]))
	}
	handle := C.api_create_with_peer_flag(api, cvalues[0], cvalues[1], cvalues[2], cvalues[3], cvalues[4], cvalues[5], cvalues[6], cvalues[7])
	if handle < 0 {
		C.api_close(api)
		return nil, fmt.Errorf("create Huginn messenger: native error %d", int64(handle))
	}
	return &Native{api: api, handle: handle}, nil
}

func (n *Native) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed {
		return nil
	}
	C.api_destroy(n.api, n.handle)
	C.api_close(n.api)
	n.closed = true
	return nil
}

func (n *Native) call(invoke func() *C.char) (json.RawMessage, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.closed {
		return nil, errors.New("Huginn core is closed")
	}
	value := invoke()
	if value == nil {
		return nil, errors.New("Huginn core returned a null response")
	}
	defer C.api_free_string(n.api, value)
	data := json.RawMessage(C.GoString(value))
	if len(data) == 0 {
		return nil, nil
	}
	if !json.Valid(data) {
		return nil, errors.New("Huginn core returned invalid JSON")
	}
	var nativeError struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &nativeError); err == nil && nativeError.Error != "" {
		return nil, errors.New(nativeError.Error)
	}
	return data, nil
}

func (n *Native) GetMe() (json.RawMessage, error) {
	return n.call(func() *C.char { return C.api_get_me(n.api, n.handle) })
}

func (n *Native) GetPeers(query string) (json.RawMessage, error) {
	if query == "" {
		return n.call(func() *C.char { return C.api_get_peers(n.api, n.handle) })
	}
	cquery := C.CString(query)
	defer C.free(unsafe.Pointer(cquery))
	return n.call(func() *C.char { return C.api_search_peers(n.api, n.handle, cquery) })
}

func (n *Native) GetMessages(chatID string, limit, offset int) (json.RawMessage, error) {
	chat := C.CString(chatID)
	defer C.free(unsafe.Pointer(chat))
	return n.call(func() *C.char {
		return C.api_get_messages_paginated(n.api, n.handle, chat, C.int(limit), C.int(offset))
	})
}

func (n *Native) SendMessage(chatID, text string, ttlSeconds int) (json.RawMessage, error) {
	chat := C.CString(chatID)
	textValue := C.CString(text)
	defer C.free(unsafe.Pointer(chat))
	defer C.free(unsafe.Pointer(textValue))
	return n.call(func() *C.char {
		return C.api_send_message(n.api, n.handle, chat, textValue, C.int(ttlSeconds))
	})
}

func (n *Native) SendFile(chatID, caption, path string, ttlSeconds int) (json.RawMessage, error) {
	chat := C.CString(chatID)
	captionValue := C.CString(caption)
	pathValue := C.CString(path)
	defer C.free(unsafe.Pointer(chat))
	defer C.free(unsafe.Pointer(captionValue))
	defer C.free(unsafe.Pointer(pathValue))
	return n.call(func() *C.char {
		return C.api_send_file(n.api, n.handle, chat, captionValue, pathValue, C.int(ttlSeconds))
	})
}

func (n *Native) GetUpdate(timeout time.Duration) (json.RawMessage, error) {
	return n.call(func() *C.char {
		return C.api_get_event(n.api, n.handle, C.int(timeout.Milliseconds()))
	})
}

func (n *Native) GetGroups() (json.RawMessage, error) {
	return n.call(func() *C.char { return C.api_get_groups(n.api, n.handle) })
}

func (n *Native) CreateGroup(name string) (json.RawMessage, error) {
	value := C.CString(name)
	defer C.free(unsafe.Pointer(value))
	return n.call(func() *C.char { return C.api_create_group(n.api, n.handle, value) })
}

func (n *Native) InviteToGroup(groupID, userID string) (json.RawMessage, error) {
	group := C.CString(groupID)
	user := C.CString(userID)
	defer C.free(unsafe.Pointer(group))
	defer C.free(unsafe.Pointer(user))
	return n.call(func() *C.char { return C.api_invite_to_group(n.api, n.handle, group, user) })
}

func (n *Native) MarkRead(messageID string) (json.RawMessage, error) {
	value := C.CString(messageID)
	defer C.free(unsafe.Pointer(value))
	return n.call(func() *C.char { return C.api_mark_read(n.api, n.handle, value) })
}
