package fesgo

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc
