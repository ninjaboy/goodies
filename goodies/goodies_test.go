package goodies

import (
	"testing"
	"time"
)

func TestGoodiesCacheAdd(testing *testing.T) {
	goodies := NewGoodies(ExpireNever, "", 0)

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
	goodies := NewGoodies(ExpireNever, "", 0)
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
	goodies := NewGoodies(25*time.Millisecond, "", 0)
	goodies.Set("nonexp", 1, ExpireNever)
	goodies.Set("exp", 1, ExpireDefault)
	<-time.After(10 * time.Millisecond)
	if _, found := goodies.Get("exp"); !found {
		testing.Error("Expired too soon")
	}
	<-time.After(30 * time.Millisecond)
	if _, found := goodies.Get("exp"); found {
		testing.Error("Not expired but expected to have expired")
	}
	<-time.After(1000 * time.Millisecond)
	if _, found := goodies.Get("nonexp"); !found {
		testing.Error("Non expired item cannot be found")
	}
}

func TestGoodiesPersisted(testing *testing.T) {
	filename := "goodies_test.dat"
	goodies := NewGoodies(25*time.Second, filename, 50*time.Second)
	expected := "expected"
	goodies.Set("test", expected, ExpireDefault)
	goodies.Stop()
	<-time.After(1000 * time.Millisecond)
	goodies2 := NewGoodies(2*time.Second, filename, 30*time.Second)
	received, ok := goodies2.Get("test")
	if !ok || (received != expected) {
		testing.Error("Basic persistence test failed")
	}
	goodies2.Stop()
}

func TestGoodiesNotAListError(testing *testing.T) {
	goodies := NewGoodies(25*time.Millisecond, "", 0)
	goodies.Set("value", 1, ExpireNever)
	err := goodies.ListPush("value", 1, ExpireNever)
	if err == nil {
		testing.Error("Type check for list push doesn't work")
	}
	_, err2 := goodies.ListLen("value")
	if err2 == nil {
		testing.Error("Type check for list len doesn't work")
	}
}

func TestGoodiesSimpleListOps(testing *testing.T) {
	goodies := NewGoodies(25*time.Millisecond, "", 0)
	goodies.ListPush("list", "Join the", ExpireNever)
	goodies.ListPush("list", "Dark Side", ExpireNever) //TODO: Rubbish API design. Rethinking required (e.g. g.SetExpire(key))
	len, err := goodies.ListLen("list")
	if err != nil {
		testing.Error("List was not created on push")
	}
	if len != 2 {
		testing.Error("List of incorrect size")
	}
}
