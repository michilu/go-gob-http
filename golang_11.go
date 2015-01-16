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

// +build !go1.3

package dvr

import (
	"net/http"
)

// This file contains functions calls that will be put in place with golang's
// prior to 1.3.

// This call does nothing since golang's prior to 1.3 do not have TLS fields.
func newGobRequestVS(req *http.Request, r *gobRequest) {
}

// This call does nothing since golang's prior to 1.3 do not have TLS fields.
func newGobResponseVS(resp *http.Response, r *gobResponse) {
}

// This call does nothing since golang's prior to 1.3 do not have TLS fields.
func (g *gobQuery) requestResponseVS(r *RequestResponse) {
}
