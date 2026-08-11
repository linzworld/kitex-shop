package main

import (
	"context"
	"errors"
	"testing"

	"example_shop/internal/bizerror"
	item "example_shop/kitex_gen/example/shop/item"
	"example_shop/kitex_gen/example/shop/stock"

	"github.com/cloudwego/kitex/client/callopt"
	"github.com/cloudwego/kitex/pkg/kerrors"
)

type stockClientStub struct {
	calls int
	resp  *stock.GetItemStockResp
	err   error
}

func (s *stockClientStub) GetItemStock(
	ctx context.Context,
	req *stock.GetItemStockReq,
	callOptions ...callopt.Option,
) (*stock.GetItemStockResp, error) {
	s.calls++
	return s.resp, s.err
}

func TestItemServiceRejectsMissingID(t *testing.T) {
	tests := []struct {
		name string
		req  *item.GetItemReq
	}{
		{name: "nil request", req: nil},
		{name: "zero ID", req: &item.GetItemReq{}},
		{name: "negative ID", req: &item.GetItemReq{Id: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stockClient := &stockClientStub{}
			handler := &ItemServiceImpl{stockCli: stockClient}

			resp, err := handler.GetItem(context.Background(), tt.req)
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
			if stockClient.calls != 0 {
				t.Fatalf("stock client called %d times for invalid request", stockClient.calls)
			}
		})
	}
}

func TestItemServiceBuildsResponseFromStock(t *testing.T) {
	stockClient := &stockClientStub{resp: &stock.GetItemStockResp{Stock: 12}}
	handler := &ItemServiceImpl{stockCli: stockClient}

	resp, err := handler.GetItem(context.Background(), &item.GetItemReq{Id: 1024})
	if err != nil {
		t.Fatalf("GetItem returned error: %v", err)
	}
	if resp.GetItem().GetId() != 1024 || resp.GetItem().GetStock() != 12 {
		t.Fatalf("unexpected item response: %#v", resp.GetItem())
	}
	if stockClient.calls != 1 {
		t.Fatalf("stock client called %d times, want 1", stockClient.calls)
	}
}

func TestItemServiceFallsBackToZeroStock(t *testing.T) {
	stockClient := &stockClientStub{err: errors.New("stock unavailable")}
	handler := &ItemServiceImpl{stockCli: stockClient}

	resp, err := handler.GetItem(context.Background(), &item.GetItemReq{Id: 1024})
	if err != nil {
		t.Fatalf("GetItem returned error: %v", err)
	}
	if resp.GetItem().GetStock() != 0 {
		t.Fatalf("stock = %d, want 0", resp.GetItem().GetStock())
	}
}
