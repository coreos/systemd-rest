package cookiejar

import(
    "net/http"
    "net/url"
    "sync"
)

type CookieJar struct {
    data map[string][]*http.Cookie
    lock sync.Mutex
}

func (jar CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
    jar.lock.Lock()
    jar.data[u.Host] = cookies
    jar.lock.Unlock()
}

func (jar CookieJar) Cookies(u *url.URL) []*http.Cookie {
    // FIXME: This is a very naive implementation
    return jar.data[u.Host]
}

func NewCookieJar() CookieJar {
    return CookieJar{
        data: make(map[string][]*http.Cookie),
        lock: sync.Mutex{},
    }
}