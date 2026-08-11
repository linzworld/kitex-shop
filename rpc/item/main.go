/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"log"
	"net"

	item "example_shop/kitex_gen/example/shop/item/itemservice"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/server"
)

func main() {
	itemServiceImpl := new(ItemServiceImpl)
	stockCli, err := NewStockClient()
	if err != nil {
		log.Fatal(err)
	}
	itemServiceImpl.stockCli = stockCli

	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:8891")
	if err != nil {
		log.Fatal(err)
	}

	svr := item.NewServer(itemServiceImpl,
		server.WithServiceAddr(addr),
		server.WithMetaHandler(transmeta.ServerTTHeaderHandler),
		server.WithEnableContextTimeout(true),
		server.WithServerBasicInfo(
			&rpcinfo.EndpointBasicInfo{
				ServiceName: "example.shop.item",
			}),
	)

	err = svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
