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
	legacy := CheckResult{Location: "order_items[2].user_url"}
	if got := legacy.InstancePath(); got != "/orderItems/2/userUrl" {
		t.Fatalf("legacy pointer: %s", got)
	}
}

func TestCheckResultStructuredPrefix(t *testing.T) {
	result := CheckResult{RuleID: "required", CanonicalLocation: Location().Property("product_name")}
	result = result.PrefixedBy(Location().Property("order_item_list").At(0))
	if result.ModelPath() != "order_item_list[0].product_name" ||
		result.NativePath() != "OrderItemList[0].ProductName" ||
		result.InstancePath() != "/orderItemList/0/productName" {
		t.Fatalf("unexpected prefixed paths: %s %s %s", result.ModelPath(), result.NativePath(), result.InstancePath())
	}
}
