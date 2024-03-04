package fesgo

import (
	"encoding/base64"
	"net/http"
)

type Account struct {
	UnAuthHandler func(*Context) // 未授权处理
	Users         map[string]string
}

func (a *Account) BasicAuth(next HandlerFunc) HandlerFunc {
	return func(c *Context) {
		user, password, ok := c.R.BasicAuth()
		if !ok {
			c.W.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			if a.UnAuthHandler != nil {
				a.UnAuthHandler(c)
			} else {
				c.SetStatusCode(http.StatusUnauthorized)
			}
			return
		}
		if p, ok := a.Users[user]; !ok || p != password {
			c.W.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			c.SetStatusCode(http.StatusUnauthorized)
			return
		}
		c.Set("user", user)
		next(c)
	}
}

func BasicAuth(username string, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
