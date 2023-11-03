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
	"encoding/json"

	pkgHTTP "github.com/apache/apisix-go-plugin-runner/pkg/http"
	"github.com/apache/apisix-go-plugin-runner/pkg/log"
	"github.com/apache/apisix-go-plugin-runner/pkg/plugin"
	"github.com/google/uuid"
)

func init() {
	err := plugin.RegisterPlugin(&MyRewriteResponse{})
	if err != nil {
		log.Fatalf("failed to register plugin MyRewriteResponse: %s", err)
	}
}

// it to the upstream.
type MyRewriteResponse struct {
	// Embed the default plugin here,
	// so that we don't need to reimplement all the methods.
	plugin.DefaultPlugin
}

func (p *MyRewriteResponse) Name() string {
	return "my-response-rewrite"
}

func (p *MyRewriteResponse) ResponseFilter(conf interface{}, w pkgHTTP.Response) {
	w.Header().Set("responseid", uuid.New().String())

	tag := conf.(MyRewriteResponseConf).Tag
	if len(tag) > 0 {
		_, err := w.Write([]byte(tag))
		if err != nil {
			log.Errorf("failed to write: %s", err)
		}
	}
}

type MyRewriteResponseConf struct {
	Tag string `json:"tag"`
}

func (p *MyRewriteResponse) ParseConf(in []byte) (interface{}, error) {
	conf := MyRewriteResponseConf{}
	err := json.Unmarshal(in, &conf)
	return conf, err
}
