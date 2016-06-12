package goodies

import "testing"
import "time"

func TestGoodiesCacheAdd(testing *testing.T) {
	goodies := NewGoodies(ExpireNever)

	key := "test"
	expected := "expected"
	goodies.Set(key, expected, ExpireNever)

	if value, found := goodies.Get(key); found {
		if value != expected {
			testing.Error("Value was found but was incorrect")
		}
	} else {
		testing.Error("Getting of even a simple string failed")
	}
}

func TestGoodiesAddList(testing *testing.T) {
	goodies := NewGoodies(ExpireNever)
	list := []int{1, 2, 3, 4, 5}
	goodies.Set("list", &list, ExpireNever)
	if lst, found := goodies.Get("list"); found {
		expectedList := lst.(*[]int)
		if (*expectedList)[4] != 5 {
			testing.Error("List reading failed")
		}
	} else {
		testing.Error("List not found")
	}
}

func TestGoodiesExpiry(testing *testing.T) {
	goodies := NewGoodies(25 * time.Millisecond)
	goodies.Set("nonexp", 1, ExpireNever)
	goodies.Set("exp", 1, ExpireDefault)
	<-time.After(10 * time.Millisecond)
	if _, found := goodies.Get("exp"); !found {
		testing.Error("Expired too soon")
	}
	<-time.After(15 * time.Millisecond)
	if _, found := goodies.Get("exp"); found {
		testing.Error("Not expired but expected to have expired")
	}
	<-time.After(1000 * time.Millisecond)
	if _, found := goodies.Get("nonexp"); !found {
		testing.Error("Non expired item has been removed")
	}

}

func TestGoodiesPersisted(testing *testing.T) {
	goodies := NewGoodiesPersisted(25 * time.Second, "goodies_test.dat")
	
}
