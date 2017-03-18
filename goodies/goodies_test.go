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
	key := "test"
	goodies.Set(key, expected, ExpireDefault)
	goodies.Stop()
	<-time.After(1000 * time.Millisecond)
	goodies2 := NewGoodies(2*time.Second, filename, 30*time.Second)
	received, ok := goodies2.Get(key)
	if !ok || (received != expected) {
		testing.Error("Basic persistence test failed")
	}
	goodies2.Stop()
}

func TestGoodiesNotAListError(testing *testing.T) {
	goodies := NewGoodies(25*time.Millisecond, "", 0)
	goodies.Set("value", 1, ExpireNever)
	err := goodies.ListPush("value", 1)
	if err == nil {
		testing.Error("Type check for list push doesn't work")
	}
	_, err2 := goodies.ListLen("value")
	if err2 == nil {
		testing.Error("Type check for list len doesn't work")
	}
}

func TestExpiryApi(testing *testing.T) {
	goodies := NewGoodies(50*time.Millisecond, "", 0)
	key := "list"
	err := goodies.ListPush(key, "val")
	defer goodies.Remove(key)
	if err != nil {
		testing.Error("List already exists")
	}

	<-time.After(100 * time.Millisecond)
	_, ok := goodies.Get(key)
	if ok {
		testing.Error("List expiration doesn't work")
	}

	err = goodies.ListPush(key, "val2")
	err = goodies.ListPush(key, "val3")
	if err != nil {
		testing.Error("List already exists")
	}
	goodies.SetExpiry(key, 150*time.Millisecond)
	<-time.After(100 * time.Millisecond)
	len, err2 := goodies.ListLen(key)
	if err2 != nil {
		testing.Error("Set new expiration doesn't work")
	}
	if len != 2 {
		testing.Error("Returned list len doesn't match")
	}

	<-time.After(100 * time.Millisecond)
	err = goodies.SetExpiry(key, 100*time.Millisecond)
	if err == nil {
		testing.Error("Error should be thrown as item is expected to become outdated already")
	}
}

func TestGoodiesSimpleListOps(testing *testing.T) {
	goodies := NewGoodies(ExpireNever, "", 0)
	key := "list"
	//ListPush test
	goodies.ListPush(key, "Where is")
	goodies.ListPush(key, "the")
	goodies.ListPush(key, "Money")
	goodies.ListPush(key, "Lebowski")

	//ListLen test
	len, err := goodies.ListLen("list")
	if err != nil {
		testing.Error("List was not created on push")
	}
	if len != 4 {
		testing.Error("List of incorrect size")
	}

	//ListRemove test
	err = goodies.ListRemoveIndex("list", 0)
	if err != nil {
		testing.Error("List doesn't exist")
	}
	len, err = goodies.ListLen(key)
	if err != nil {
		testing.Errorf("Unexpected error %v", err)
	}
	if len != 3 {
		testing.Error("List of incorrect size")
	}

}
