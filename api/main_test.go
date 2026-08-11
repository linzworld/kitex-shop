package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"example_shop/internal/bizerror"
	"example_shop/kitex_gen/example/shop/item"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/client/callopt"
)

type itemClientStub struct {
	resp *item.GetItemResp
	err  error
}

func (s *itemClientStub) GetItem(
	ctx context.Context,
	req *item.GetItemReq,
	callOptions ...callopt.Option,
) (*item.GetItemResp, error) {
	return s.resp, s.err
}

func TestHandlerReturnsProtobufResponseAsJSON(t *testing.T) {
	previousClient := cli
	t.Cleanup(func() { cli = previousClient })
	cli = &itemClientStub{resp: &item.GetItemResp{
		Item: &item.Item{Id: 1024, Title: "Kitex", Stock: 12},
	}}

	reqCtx := app.NewContext(0)
	Handler(context.Background(), reqCtx)

	if reqCtx.Response.StatusCode() != 200 {
		t.Fatalf("status = %d, want 200", reqCtx.Response.StatusCode())
	}
	body := string(reqCtx.Response.Body())
	if !strings.Contains(body, `"id":1024`) || !strings.Contains(body, `"stock":12`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestHandlerMapsBusinessError(t *testing.T) {
	previousClient := cli
	t.Cleanup(func() { cli = previousClient })
	cli = &itemClientStub{err: bizerror.InvalidArgument("id")}

	reqCtx := app.NewContext(0)
	Handler(context.Background(), reqCtx)

	if reqCtx.Response.StatusCode() != 400 {
		t.Fatalf("status = %d, want 400", reqCtx.Response.StatusCode())
	}
	if !strings.Contains(string(reqCtx.Response.Body()), "40001") {
		t.Fatalf("business error code missing from response: %s", reqCtx.Response.Body())
	}
}

func TestHandlerMapsSystemError(t *testing.T) {
	previousClient := cli
	t.Cleanup(func() { cli = previousClient })
	cli = &itemClientStub{err: errors.New("connection refused")}

	reqCtx := app.NewContext(0)
	Handler(context.Background(), reqCtx)

	if reqCtx.Response.StatusCode() != 502 {
		t.Fatalf("status = %d, want 502", reqCtx.Response.StatusCode())
	}
}
