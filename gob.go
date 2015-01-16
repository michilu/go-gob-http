// Copyright 2015 ENDOH takanao.
// <https://github.com/MiCHiLU/go-gob-http>
//
// Copyright 2014 Orchestrate, Inc.
// <https://github.com/orchestrate-io/dvr>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gobhttp

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
)

//
// Error wrapper
//

// This is a safe type that will always be able to be gob encoded.
type gobSafeError string

// Error() for gobSafeError
func (g gobSafeError) Error() string {
	return string(g)
}

// This is the list of known encodable types saved as a map of name -> bool.
// This allows us to know if a given type will be decodable or not.. If not
// then we need to wrap the type in a gobSafeError structure.
var encodableTypes map[string]bool = map[string]bool{}

// Initialize encodableTypes
func init() {
	// Error return types.
	registerErrorType(new(gobSafeError))
	registerErrorType(new(http.ProtocolError))
	registerErrorType(new(net.AddrError))
	registerErrorType(new(net.DNSConfigError))
	registerErrorType(new(net.DNSError))
	registerErrorType(new(net.DNSError))
	registerErrorType(new(net.InvalidAddrError))
	registerErrorType(new(net.OpError))
	registerErrorType(new(net.ParseError))
	registerErrorType(new(net.UnknownNetworkError))
	registerErrorType(new(url.Error))
	registerErrorType(new(url.EscapeError))

	// Other objects that we might end up seeing.
	gob.Register(new(rsa.PublicKey))
	gob.Register(new(rsa.PrivateKey))
	gob.Register(new(net.TCPAddr))
}

// Adds an error interface object to the list of known types that this library
// will be able to encode. This is necessary due to the way that gob encodes
// interface object. The only error types here are those that will be returned
// from the RoundTripper object. Typical use cases should not need this at all.
// If you are using this you must do it via your modules init() otherwise
// results can be unpredictable.
func registerErrorType(err error) {
	// Walk the given interface all the way down to the raw object.
	value := reflect.ValueOf(err)
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		if _, ok := value.Interface().(error); ok {
			gob.Register(value.Interface())
		}
		value = value.Elem()
	}
	typ := value.Type()
	id := fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
	encodableTypes[id] = true
}

// This type is used to store errors. Since some errors might contain private
// fields we need to ensure that we can still convert them as best as possible.
// Specifically this will convert them to a string error.
type gobError struct {
	Error error
}

// The encoder that allows us to safely convert errors.
func (g *gobError) GobEncode() ([]byte, error) {
	if g.Error == nil {
		return []byte{}, nil
	}

	// If we are encoding a known safe type then we write the types name and
	// then encode it into the byte stream, otherwise we are forced to convert
	// it into a gobSafeError type so it can be safely stored.
	value := reflect.ValueOf(g.Error)
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Figure out the name of this object type.
	id := fmt.Sprintf("%s.%s", value.Type().PkgPath(), value.Type().Name())

	// Make a safe error object for us to encode with.
	rawError := gobRawError{
		Error:             g.Error,
		ErrorsErrorString: id == "errors.errorString",
	}

	// If the object that we are encoding is not safe then we need to change
	// it into one that actually is.
	if _, ok := encodableTypes[id]; !ok {
		rawError.Error = gobSafeError(g.Error.Error())
	}

	// Encode the safe object and return the byte array.
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(&rawError)
	return buffer.Bytes(), err
}

// Decodes a gobError struct
func (g *gobError) GobDecode(data []byte) error {
	if len(data) == 0 {
		g.Error = nil
		return nil
	}
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	var rawError gobRawError
	if err := decoder.Decode(&rawError); err != nil {
		return err
	}

	if rawError.ErrorsErrorString {
		g.Error = errors.New(rawError.Error.Error())
	} else {
		g.Error = rawError.Error
	}
	return nil
}

// This is the object type that gobError will use when encoding and decoding.
// The assumption here is that the Error field in this object will be populated
// only with safe error types.
type gobRawError struct {
	Error error

	// This is set to true if the error was initially a 'errors.errorString'
	// so we know that we can convert it back in the decoding process.
	ErrorsErrorString bool
}

//
// Request wrapper
//

// This is a gob encodable version of http.Request. It removes the possibility
// that a private field can be seen and therefor return an error.
type GobRequest struct {
	Method           string
	URL              string
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
	Form             url.Values
	PostForm         url.Values
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	TLS              *tls.ConnectionState

	// The request body and err returned when reading it.
	Body  []byte
	Error gobError
}

// This takes a Request object and returns a gob compatible GobRequest object.
func NewGobRequest(req *http.Request) *GobRequest {
	if req == nil {
		return nil
	}

	r := new(GobRequest)
	r.Method = req.Method
	r.URL = req.URL.String()
	r.Proto = req.Proto
	r.ProtoMajor = req.ProtoMajor
	r.ProtoMinor = req.ProtoMinor
	r.Header = req.Header
	r.ContentLength = req.ContentLength
	r.TransferEncoding = req.TransferEncoding
	r.Close = req.Close
	r.Host = req.Host
	r.Form = req.Form
	r.PostForm = req.PostForm
	r.Trailer = req.Trailer
	r.RemoteAddr = req.RemoteAddr
	r.RequestURI = req.RequestURI
	newGobRequestVS(req, r)

	return r
}

//
// Response wrapper
//

// This object wraps an http.Response object in order to make it possible to
// gob encode it.
type GobResponse struct {
	Status           string
	StatusCode       int
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Header           http.Header
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Trailer          http.Header
	TLS              *tls.ConnectionState

	// The response body and err returned when reading it.
	Body  []byte
	Error gobError
}

// This takes a Response object and returns a gob compatible GobResponse object.
func newGobResponse(resp *http.Response) *GobResponse {
	if resp == nil {
		return nil
	}

	r := new(GobResponse)
	r.Status = resp.Status
	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	r.ProtoMajor = resp.ProtoMajor
	r.ProtoMinor = resp.ProtoMinor
	r.Header = resp.Header
	r.ContentLength = resp.ContentLength
	r.TransferEncoding = resp.TransferEncoding
	r.Close = resp.Close
	r.Trailer = resp.Trailer
	newGobResponseVS(resp, r)

	return r
}
