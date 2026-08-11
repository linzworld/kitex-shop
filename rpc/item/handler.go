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
	"context"
	"log"

	"example_shop/internal/bizerror"
	item "example_shop/kitex_gen/example/shop/item"
	"example_shop/kitex_gen/example/shop/stock"
	"example_shop/kitex_gen/example/shop/stock/stockservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/transmeta"
	"github.com/cloudwego/kitex/transport"
)

// ItemServiceImpl implements the last service interface defined in the IDL.
type ItemServiceImpl struct {
	stockCli stockservice.Client
}

func NewStockClient() (stockservice.Client, error) {
	return stockservice.NewClient(
		"example.shop.stock",
		client.WithHostPorts("stock:8890"),
		client.WithTransportProtocol(transport.TTHeader),
		client.WithMetaHandler(transmeta.ClientTTHeaderHandler),
	)
}

// GetItem implements the ItemServiceImpl interface.
func (s *ItemServiceImpl) GetItem(ctx context.Context, req *item.GetItemReq) (resp *item.GetItemResp, err error) {
	if req == nil || req.GetId() <= 0 {
		return nil, bizerror.InvalidArgument("id")
	}

	resp = &item.GetItemResp{
		Item: &item.Item{
			Id:          req.GetId(),
			Title:       "Kitex",
			Description: "Kitex is an excellent framework!",
		},
	}

	stockReq := &stock.GetItemStockReq{ItemId: req.GetId()}
	stockResp, err := s.stockCli.GetItemStock(ctx, stockReq)
	if err != nil {
		log.Printf("get item stock failed: %v", err)
		return resp, nil
	}
	if stockResp != nil {
		resp.Item.Stock = stockResp.GetStock()
	}
	return
}
