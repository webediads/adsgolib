package middlewares

type key int

var ContextKeyIP = key(1)
var ContextKeyReferer = key(2)
var ContextKeyUserAgent = key(3)
