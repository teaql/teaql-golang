package runtime

import "testing"

func TestObjectLocationPathDialects(t *testing.T) {
	location := Location().Property("order_items").At(2).Property("user_url")
	if got := location.ModelPath(); got != "order_items[2].user_url" {
		t.Fatalf("model path: %s", got)
	}
	if got := location.NativePath(); got != "OrderItems[2].UserUrl" {
		t.Fatalf("native path: %s", got)
	}
	if got := location.InstancePath(); got != "/orderItems/2/userUrl" {
		t.Fatalf("instance path: %s", got)
	}
	escaped := Location().Property("a/b~c")
	if got := escaped.InstancePath(); got != "/a~1b~0c" {
		t.Fatalf("escaped pointer: %s", got)
	}
}
