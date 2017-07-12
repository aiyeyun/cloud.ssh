package session

import (
	"sync"
	"errors"
	"io"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"
)


type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

type Session interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Delete(ket interface{}) error
	SessionID() string
}

type Manager struct {
	cookieName  string
	lock        sync.Mutex
	provider    Provider
	maxLifeTime int64
}

var provides = make(map[string]Provider)

func NewSessionManager(provideName, cookieName string, maxLifeTime int64) (*Manager, error) {
	provide, ok := provides[provideName]
	if !ok {
		return nil, errors.New("session: unknown provide forgotten import? "+ provideName)
	}
	return &Manager{cookieName: cookieName, provider: provide, maxLifeTime: maxLifeTime}, nil
}

func Register(name string, provide Provider)  {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provide " + name)
	}
	provides[name] = provide
}

//生成一个新的 id
func (manager *Manager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request, sid string) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if sid != "" {
		session, _ = manager.provider.SessionRead(sid)
		return session
	}

	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sid := manager.sessionId()
		session, _ = manager.provider.SessionInit(sid)
		cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(manager.maxLifeTime)}
		http.SetCookie(w, &cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRead(sid)
	}
	return session
}

//session 销毁
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request, sid string)  {

	if sid != "" {
		manager.lock.Lock()
		defer manager.lock.Unlock()
		manager.provider.SessionDestroy(sid)
		expiration := time.Now()
		ck := http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &ck)
		return
	}

	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionDestroy(cookie.Value)
	expiration := time.Now()
	ck := http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
	http.SetCookie(w, &ck)
}

//session 回收
func (manager *Manager) GC()  {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionGC(manager.maxLifeTime)
	time.AfterFunc(time.Duration(manager.maxLifeTime), func() {
		manager.GC()
	})
}