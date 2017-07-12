package session

import (
	"io"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"
	"sync"
)

type SSHInterface interface {
	Get(key string) *SSHInfo
	Set(key string, value *SSHInfo)
	Del(key string)
}

type SSHSessionManage struct {
	cookieName string
}

type SSHInfo struct {
	Addr     string
	Port     string
	Username string
	Password string
}

type SSHList struct {
	lock sync.RWMutex
	List map[string]*SSHInfo
}

var SSHListManage *SSHList

func NewSSHSessionManage(cookieName string) (*SSHSessionManage, error) {
	return &SSHSessionManage{cookieName: cookieName}, nil
}

//生成一个新的 id
func (manager *SSHSessionManage) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (manager *SSHSessionManage) Start(w http.ResponseWriter, r *http.Request, sid string) (string) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sid := manager.sessionId()
		cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: 0}
		http.SetCookie(w, &cookie)
		return sid
	}
	sessid, _ := url.QueryUnescape(cookie.Value)
	return sessid
}

//session 销毁
func (manager *SSHSessionManage) Destroy(w http.ResponseWriter, r *http.Request, sid string)  {
	if sid != "" {
		expiration := time.Now()
		ck := http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &ck)
		return
	}

	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	expiration := time.Now()
	ck := http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
	http.SetCookie(w, &ck)
}

func (sshList *SSHList) Get(key string) *SSHInfo {
	sshList.lock.RLock()
	defer sshList.lock.RUnlock()
	if _, ok := sshList.List[key]; !ok {
		return nil
	}
	return sshList.List[key]
}

func (sshList *SSHList) Set(key string, value *SSHInfo)  {
	sshList.lock.Lock()
	defer sshList.lock.Unlock()
	sshList.List[key] = value
}

func (sshList *SSHList) Del(key string)  {
	sshList.lock.Lock()
	defer sshList.lock.Unlock()
	delete(sshList.List, key)
}

func init()  {
	SSHListManage = new(SSHList)
	SSHListManage.List = make(map[string]*SSHInfo)
}