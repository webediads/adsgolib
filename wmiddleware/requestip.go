package wmiddleware

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/webediads/adsgolib/wcontext"
)

// ipRange : a structure that holds the start and end of a range of ip addresses
type ipRange struct {
	start net.IP
	end   net.IP
}

// inRange : check to see if a given ip address is within a given range
func inRange(r ipRange, ipAddress net.IP) bool {
	// strcmp type byte comparison
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

var privateRanges = []ipRange{
	{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
	// my use case is only concerned with ipv4 atm
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		// iterate over all our ranges
		for _, r := range privateRanges {
			// check if this ip is in a private range
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

// RequestIP handles reading the input data and sets the IP of the request in the context
func RequestIP(contextKeyIP wcontext.Key) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// default without proxy
			fromIP, _, _ := net.SplitHostPort(r.RemoteAddr)

			for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
				addresses := strings.Split(r.Header.Get(h), ",")
				// march from right to left until we get a public address
				// that will be the address right before our proxy.
				for i := len(addresses) - 1; i >= 0; i-- {
					ip := strings.TrimSpace(addresses[i])
					if len(ip) > 1 {
						if !strings.Contains(ip, ":") {
							ip = ip + ":"
						}
						// header can contain spaces too, strip those out.
						ip, _, err := net.SplitHostPort(ip)
						if err == nil {
							realIP := net.ParseIP(ip)
							if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
								// bad address, go to next
								continue
							}
							fromIP = ip
						} else {
							continue
						}
					} else {
						continue
					}
				}
			}

			ctx = context.WithValue(r.Context(), contextKeyIP, fromIP)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
