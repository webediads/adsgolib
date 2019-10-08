package middleware

type key int

var ContextKeyRequestIP = key(1)
var ContextKeyReferer = key(2)
var ContextKeyUserAgent = key(3)
