package wmiddleware

type key int

// ContextKeyRequestIP is the context key for our middleware that adds the request ip
var ContextKeyRequestIP = key(1)

// ContextKeyReferer is the context key for our middleware that adds the referer
var ContextKeyReferer = key(2)

// ContextKeyUserAgent is the context key for our middleware that adds the user agent
var ContextKeyUserAgent = key(3)
