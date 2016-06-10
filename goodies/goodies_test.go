package goodies

import "testing"

func TestGoodiesCache(testing *testing.T) {
	goodies := CreateGoodies()

	key := "test"
	expected := "expected"
	goodies.Set(key, expected)

	if value, found := goodies.Get(key); found {
		if value != expected {
			testing.Error("Value was found but was incorrect")
		}
	} else {
		testing.Error("Getting of even a simple string failed")
	}

	list := []int{1, 2, 3, 4, 5}
	goodies.Set("list", &list)
	if lst, found := goodies.Get("list"); found {
		expectedList := lst.(*[]int)
		if (*expectedList)[4] != 5 {
			testing.Error("List reading failed")

		}
	} else {
		testing.Error("List not found")
	}

}
