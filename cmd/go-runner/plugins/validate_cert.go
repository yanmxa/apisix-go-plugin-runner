/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugins

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/google/uuid"
)

const (
	CERT_PATH    = "apisix-go-plugin-runner/resource"
	CN_TYPE_CODE = "2.5.4.3"
)

var certDir string

func init() {
	err := plugin.RegisterPlugin(&ValidateCert{})
	if err != nil {
		log.Fatalf("failed to register plugin Validate Cert: %s", err)
	}
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get the current dir: %s", err)
	}
	certDir = filepath.Join(currentDir, CERT_PATH)
}

// it to the upstream.
type ValidateCert struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

func (p *ValidateCert) Name() string {
	return "validate-cert"
}

func (p *ValidateCert) RequestFilter(conf interface{}, w http.ResponseWriter, r pkgHTTP.Request) {
	w.Header().Add("X-Resp-A6-Runner", "Go")
	w.Header().Set("X-Gateway-Verifid", "false")
	w.Header().Set("Traceid", uuid.New().String())

	var (
		err        error
		statusCode int
	)

	defer func() {
		if err != nil {
			resp := p.response(statusCode, fmt.Sprintf("authentication failed: %s", err.Error()))
			_, e := w.Write(resp)
			if e != nil {
				log.Errorf("failed to write: %s", e)
			}
		}
	}()

	// clientCertFile := filepath.Join(certDir, "client.crt")
	// clientCertPEM, err := os.ReadFile(clientCertFile)
	// if err != nil {
	// 	log.Errorf("Failed to read client certificate file:", err)
	// 	statusCode = 400
	// 	return
	// }
	// bodyBytes, err := r.Body()
	// if err != nil {
	// 	log.Errorf("Failed to read request body:", err)
	// 	statusCode = 400
	// 	return
	// }
	// body := &Body{}
	// if err == json.Unmarshal(bodyBytes, body) {
	// 	log.Errorf("Failed to unmarshal request body:", err)
	// 	statusCode = 400
	// 	return
	// }
	decodedBytes, err := base64.StdEncoding.DecodeString(r.Header().Get("Client-Certificate"))
	if err != nil {
		log.Errorf("Failed to read client certificate file:", err)
		statusCode = 400 // Bad Request: The request is malformed, and the server cannot understand it.
		return
	}

	clientCertPEM := decodedBytes
	clientName := r.Header().Get("Source")

	caCertFile := filepath.Join(certDir, "ca.crt")
	caCertPEM, err := os.ReadFile(caCertFile)
	if err != nil {
		log.Errorf("Failed to read CA certificate file:", err)
		statusCode = 400 // Bad Request: The request is malformed, and the server cannot understand it.
		return
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCertPEM)
	if !ok {
		err = fmt.Errorf("Failed to parse CA certificate")
		statusCode = 400
		return
	}

	// Decode the certificate
	block, _ := pem.Decode(clientCertPEM)
	if block == nil {
		err = fmt.Errorf("Failed to decode certificate PEM. Header['Client-Certificate'] %s", decodedBytes)
		statusCode = 400
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Errorf("failed to parse certificate: ", err)
		statusCode = 400
		return
	}

	// Verify Certificate
	if _, err := cert.Verify(x509.VerifyOptions{
		Roots: caCertPool,
	}); err != nil {
		log.Errorf("failed to verify certificate: ", err)
		return
	}

	// Verify CN
	cn := ""
	for _, name := range cert.Subject.Names {
		if name.Type.String() == CN_TYPE_CODE {
			cn = name.Value.(string)
			break
		}
		// fmt.Println("name.Type.String()", name.Type.String(), "name.Value.(string)", name.Value.(string))
	}

	if strings.EqualFold(cn, clientName) {
		w.Header().Set("X-Gateway-Verifid", "true")
	} else {
		statusCode = 401
		err = fmt.Errorf("Unauthorized to identity: %s", clientName)
	}
}

func (p *ValidateCert) ParseConf(in []byte) (interface{}, error) {
	return nil, nil
}

func (p *ValidateCert) response(code int, message interface{}) (resp []byte) {
	resp, _ = json.Marshal(map[string]interface{}{
		"code":    code,
		"message": message,
	})
	return
}

type Body struct {
	Source  string `json:"source"`
	Cert    string `json:"cert"`
	Payload []byte `json:"payload"`
}
