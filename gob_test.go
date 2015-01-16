// Copyright 2014 Orchestrate, Inc.
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

package dvr

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/liquidgecka/testlib"
)

// An error implementation that is unknown to the decoder.
type customError string

func (p customError) Error() string {
	return string(p)
}

func TestGobError_GobEncode(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()

	// Test 1: An unencodable error.

	g := &gobError{Error: customError("Unexpected")}

	// Encode the error.
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	T.ExpectSuccess(encoder.Encode(g))

	// Decode the byte array to see what was returned.
	g2 := new(gobError)
	decoder := gob.NewDecoder(buffer)
	T.ExpectSuccess(decoder.Decode(g2))

	// Make sure that the two encoded objects types are NOT the same.
	T.NotEqual(g, g2)

	// Ensure that the type of g2.Error is actually a gobSafeError
	if _, ok := g2.Error.(*gobSafeError); !ok {
		T.Fatalf("g2.Error is not a gobSafeEror, its a %T", g2.Error)
	}

	// Test 2: An encodable error.

	g = &gobError{Error: &http.ProtocolError{ErrorString: "Expected"}}

	// Encode the error.
	buffer = &bytes.Buffer{}
	encoder = gob.NewEncoder(buffer)
	T.ExpectSuccess(encoder.Encode(g))

	// Decode the byte array to see what was returned.
	g2 = new(gobError)
	decoder = gob.NewDecoder(buffer)
	T.ExpectSuccess(decoder.Decode(g2))

	// Ensure that the two encoded objects ARE the same.
	T.Equal(g, g2)

	// Ensure that the type of g2.Error is actually a *http.ProtocolError
	if _, ok := g2.Error.(*http.ProtocolError); !ok {
		T.Fatalf("g2.Error is not a *http.ProtocolError, its a %T", g2.Error)
	}

	// Test 3: An errors.stringError object.

	g = &gobError{Error: errors.New("Expected")}

	// Encode the error.
	buffer = &bytes.Buffer{}
	encoder = gob.NewEncoder(buffer)
	T.ExpectSuccess(encoder.Encode(g))

	// Decode the byte array to see what was returned.
	g2 = new(gobError)
	decoder = gob.NewDecoder(buffer)
	T.ExpectSuccess(decoder.Decode(g2))

	// Ensure that the two encoded objects ARE the same.
	T.Equal(g, g2)

	// Ensure that the type of g2.Error is actually a *http.ProtocolError
	if reflect.TypeOf(g.Error) != reflect.TypeOf(g2.Error) {
		T.Fatalf("g2.Error is not a *http.ProtocolError, its a %T", g2.Error)
	}
}

func TestGobError_GobDecode(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()
	// Most of the Decoding tests are in TestGobError_GobEncode other than
	// a simple bogus data check.
	g := &gobError{}
	T.ExpectError(g.GobDecode([]byte{0, 1, 2, 3}))
}

func TestGobQuery_RequestResponse(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()
	defer func() {
		err := recover()
		if err == nil {
			T.Fatalf("The expected panic didn't happen")
		} else if _, ok := err.(*dvrFailure); !ok {
			T.Fatalf("An unexpected panic happened: %#v", err)
		}
	}()
	panicOutput = ioutil.Discard
	gq := &gobQuery{Request: new(gobRequest)}
	gq.Request.URL = "://"
	gq.RequestResponse()
	T.Fatalf("The above call should never return.")
}

func TestSimpleCoverage(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()
	T.Equal(newGobRequest(nil), nil)
	T.Equal(newGobResponse(nil), nil)
}
