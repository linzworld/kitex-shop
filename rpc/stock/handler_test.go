package main

import (
	"context"
	"testing"

	"example_shop/internal/bizerror"
	stock "example_shop/kitex_gen/example/shop/stock"

	"github.com/cloudwego/kitex/pkg/kerrors"
)

func TestStockServiceRejectsMissingItemID(t *testing.T) {
	tests := []struct {
		name string
		req  *stock.GetItemStockReq
	}{
		{name: "nil request", req: nil},
		{name: "zero item ID", req: &stock.GetItemStockReq{}},
		{name: "negative item ID", req: &stock.GetItemStockReq{ItemId: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := new(StockServiceImpl).GetItemStock(context.Background(), tt.req)
			if resp != nil {
				t.Fatalf("expected nil response, got %#v", resp)
			}
			bizErr, ok := kerrors.FromBizStatusError(err)
			if !ok {
				t.Fatalf("expected BizStatusError, got %v", err)
			}
			if bizErr.BizStatusCode() != bizerror.InvalidArgumentCode {
				t.Fatalf("unexpected business error code: %d", bizErr.BizStatusCode())
			}
		})
	}
}

func TestStockServiceReturnsStock(t *testing.T) {
	resp, err := new(StockServiceImpl).GetItemStock(
		context.Background(),
		&stock.GetItemStockReq{ItemId: 1024},
	)
	if err != nil {
		t.Fatalf("GetItemStock returned error: %v", err)
	}
	if resp.GetStock() != 1024 {
		t.Fatalf("stock = %d, want 1024", resp.GetStock())
	}
}
