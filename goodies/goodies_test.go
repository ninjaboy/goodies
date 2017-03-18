package goodies

import (
	"testing"
	"time"
)

func TestGoodiesAddSetUpdate(testing *testing.T) {
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

	_, ok := goodies.Get("Non-existent")
	if ok {
		testing.Error("Enxpected item found")
	}

	goodies.Update(key, "new", ExpireDefault)
	value, found := goodies.Get(key)
	if !found || value != "new" {
		testing.Error("Update doesn't work")
	}

	_, err := goodies.Update("Non-existent", "newer", ExpireDefault)
	if err == nil {
		testing.Error("Update of non-existent is expected to throw an error")
	}

	keys := goodies.Keys()
	if len(keys) != 1 || keys[0] != key {
		testing.Error("Keys collection doesn't match expected")
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
	len2, err2 := goodies.ListLen(key)
	if err2 != nil {
		testing.Errorf("Unexpected error %v", err)
	}
	if len2 != 3 {
		testing.Error("List of incorrect size")
	}

	len3, err3 := goodies.ListLen("Non-existent-list")
	if len3 != 0 || err3 != nil {
		testing.Error("Unexpected behaviour for non existent list length retreival")
	}

	err4 := goodies.ListRemoveIndex(key, 100)
	if err4 != nil {
		testing.Error("Unexpected behaviour for removing non-existent list item by index")
	}

	notListKey := "not a list"
	goodies.Set(notListKey, "I am string!", ExpireDefault)
	err5 := goodies.ListRemoveIndex(notListKey, 0)
	if err5 == nil {
		testing.Error("Removing by index from non list didn't report an error")
	}

	err6 := goodies.ListRemoveIndex("Non-existent list", 0)
	if err6 != nil {
		testing.Error("Unexpected error when removing from non-existent list")
	}
}

type Custom struct {
	i int
	f float32
	s string
}

func TestListRemoveByValue(testing *testing.T) {
	goodies := NewGoodies(ExpireNever, "", 0)
	key := "list"
	simpleStr := "simpleString"
	i := 100
	obj := Custom{3, 0.14, "pi"}

	goodies.ListPush(key, simpleStr)
	goodies.ListPush(key, obj)
	goodies.ListPush(key, i)
	goodies.ListPush(key, i)
	goodies.ListPush(key, obj)
	goodies.ListPush(key, simpleStr)

	len, err := goodies.ListLen(key)
	if err != nil || len != 6 {
		testing.Error("List was created incorrectly")
	}
	err = goodies.ListRemoveValue(key, "simpleString")
	len, err = goodies.ListLen(key)
	if err != nil || len != 4 {
		testing.Error("List deletion failed")
		return
	}
	err = goodies.ListRemoveValue(key, Custom{3, 0.14, "pi"})
	len, err = goodies.ListLen(key)
	if err != nil || len != 2 {
		testing.Error("List deletion of struct failed")
		return
	}
	err = goodies.ListRemoveValue(key, 100)
	len, err = goodies.ListLen(key)
	if err != nil || len != 0 {
		testing.Error("List deletion of integer failed")
		return
	}
	err = goodies.ListRemoveValue(key, "Non-existent")
	if err != nil {
		testing.Errorf("Unexpected error on removing inexistent value from list by value: %v", err)
	}

	goodies.Set("valkey", 3.14, ExpireDefault)
	err = goodies.ListRemoveValue("valkey", "value")
	if err == nil {
		testing.Error("Expected error on removing by value from non list")
	}
}
