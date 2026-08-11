package protocoltest

import (
	"testing"

	"example_shop/kitex_gen/example/shop/base"
	"example_shop/kitex_gen/example/shop/item"
	"example_shop/kitex_gen/example/shop/stock"
)

func TestItemResponseProtobufRoundTrip(t *testing.T) {
	want := &item.GetItemResp{
		Item: &item.Item{
			Id:          1024,
			Title:       "Kitex",
			Description: "Kitex is an excellent framework!",
			Stock:       12,
		},
		BaseResp: &base.BaseResp{Code: "0", Msg: "ok"},
	}

	data, err := want.Marshal(nil)
	if err != nil {
		t.Fatalf("marshal item response: %v", err)
	}

	got := new(item.GetItemResp)
	if err := got.Unmarshal(data); err != nil {
		t.Fatalf("unmarshal item response: %v", err)
	}
	if got.GetItem().GetId() != want.GetItem().GetId() ||
		got.GetItem().GetStock() != want.GetItem().GetStock() ||
		got.GetBaseResp().GetCode() != want.GetBaseResp().GetCode() {
		t.Fatalf("round-trip mismatch: got %#v, want %#v", got, want)
	}
}

func TestStockRequestProtobufRoundTrip(t *testing.T) {
	want := &stock.GetItemStockReq{ItemId: 1024}

	data, err := want.Marshal(nil)
	if err != nil {
		t.Fatalf("marshal stock request: %v", err)
	}

	got := new(stock.GetItemStockReq)
	if err := got.Unmarshal(data); err != nil {
		t.Fatalf("unmarshal stock request: %v", err)
	}
	if got.GetItemId() != want.GetItemId() {
		t.Fatalf("item ID = %d, want %d", got.GetItemId(), want.GetItemId())
	}
}
