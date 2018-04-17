package http_bridge

import (
	"github.com/crushedpixel/ferry"
	"net/http"
)

func BridgeRoot(f *ferry.Ferry, mux *http.ServeMux) {
	Bridge(f, mux, "", "")
}

func Bridge(f *ferry.Ferry, mux *http.ServeMux, namespace string, contentType string) {
	namespace = NormalizeNamespace(namespace)
	// the mux pattern must end on a slash to
	// match all subroutes
	pattern := namespace + "/"

	mux.HandleFunc(pattern, HandleFunc(f, namespace, contentType))
}

func HandleFunc(f *ferry.Ferry, namespace string, contentType string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// normalize namespace
		namespace = NormalizeNamespace(namespace)

		// create new connection for request
		cr := &ferry.ConnectionRequest{
			RemoteAddr: req.RemoteAddr,
			Header:     req.Header,
		}
		conn, res := f.NewConnection(cr)
		if res != nil {
			// connection was denied
			WriteResponse(rw, res, contentType)
			return
		}

		// get URI relative to namespace
		// by removing namespace prefix
		relativeURI := req.RequestURI[len(namespace):]

		r := &ferry.IncomingRequest{
			Method:     req.Method,
			RequestURI: relativeURI,
			Payload:    req.Body,
		}
		// handle request and write response
		WriteResponse(rw, conn.Handle(r), contentType)
	}
}

// WriteResponse writes a ferry.Response to the http.ResponseWriter.
func WriteResponse(rw http.ResponseWriter, res ferry.Response, contentType string) {
	if contentType != "" {
		rw.Header().Set("Content-Type", contentType)
	}
	status, payload := res.Response()
	rw.WriteHeader(status)
	rw.Write([]byte(payload))
}

// NormalizeNamespace ensures that namespace starts with a slash
// and does not end on a slash.
func NormalizeNamespace(namespace string) string {
	// prepend slash to namespace
	if len(namespace) < 1 || namespace[0] != '/' {
		namespace = "/" + namespace
	}
	// remove trailing slash from namespace
	if namespace[len(namespace)-1] == '/' {
		namespace = namespace[:len(namespace)-1]
	}

	return namespace
}
