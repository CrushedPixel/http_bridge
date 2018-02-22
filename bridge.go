package http_bridge

import (
	"github.com/crushedpixel/ferry"
	"net/http"
)

// NormalizeNamespace ensures that namespace starts with a slash
// and does not end on a slash.
func NormalizeNamespace(namespace string) string {
	// prepend slash to namespace
	if len(namespace) < 1 || namespace[0] != '/' {
		namespace = "/" + namespace
	}
	// remove trailing slash from namespace
	if len(namespace) > 1 && namespace[len(namespace)-1] == '/' {
		namespace = namespace[:len(namespace)-1]
	}

	return namespace
}

func Bridge(f *ferry.Ferry, mux *http.ServeMux, namespace string) {
	namespace = NormalizeNamespace(namespace)
	// the mux pattern must end on a slash to
	// match all subroutes
	pattern := namespace + "/"

	mux.HandleFunc(pattern, HandleFunc(f, namespace))
}

func HandleFunc(f *ferry.Ferry, namespace string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// create new connection for request
		cr := &ferry.ConnectionRequest{
			RemoteAddr: req.RemoteAddr,
			Header:     req.Header,
		}
		conn, res := f.NewConnection(cr)
		if res != nil {
			// connection was denied
			writeResponse(rw, res)
			return
		}

		// get URI relative to namespace
		// by removing namespace prefix
		relativeURI := req.RequestURI[len(namespace):]

		r := &ferry.Request{
			Method:     req.Method,
			RequestURI: relativeURI,
			Payload:    req.Body,
		}
		// handle request and write response
		writeResponse(rw, conn.Handle(r))
	}
}

func writeResponse(rw http.ResponseWriter, res ferry.Response) {
	status, payload := res.Response()
	rw.WriteHeader(status)
	rw.Write([]byte(payload))
}
