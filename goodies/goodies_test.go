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

	value, err := goodies.Get(key)
	if err != nil {
		testing.Errorf("Unexpected error on getting value: %v", err)
	}
	if value != expected {
		testing.Error("Unexpected value")
	}

	_, err = goodies.Get("Non-existent")
	if _, ok := err.(NotFoundError); !ok {
		testing.Errorf("Expected NotFoundError but received: %v", err)
	}

	err = goodies.Update(key, "new", ExpireDefault)
	if err != nil {
		testing.Error("Unexpected error on updating existing value")
	}

	value, err = goodies.Get(key)
	if err != nil || value != "new" {
		testing.Error("Update doesn't work")
	}

	err = goodies.Update("Non-existent", "newer", ExpireDefault)
	if _, ok := err.(NotFoundError); !ok {
		testing.Error("Update of non-existent is expected to throw a not found error")
	}

	keys := goodies.Keys()
	if len(keys) != 1 || keys[0] != key {
		testing.Error("Keys collection doesn't match expected")
	}
}

func TestGoodiesExpiry(testing *testing.T) {
	goodies := NewGoodies(25*time.Millisecond, "", 0)
	goodies.Set("nonexp", "1", ExpireNever)
	goodies.Set("exp", "1", ExpireDefault)
	<-time.After(10 * time.Millisecond)
	if _, err := goodies.Get("exp"); err != nil {
		testing.Error("Expired too soon")
	}
	<-time.After(30 * time.Millisecond)
	if _, err := goodies.Get("exp"); err == nil {
		testing.Error("Not expired but expected to have expired")
	}
	<-time.After(1000 * time.Millisecond)
	if _, err := goodies.Get("nonexp"); err != nil {
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
	received, err := goodies2.Get(key)
	if err != nil || (received != expected) {
		testing.Error("Basic persistence test failed")
	}
	goodies2.Stop()
}

func TestGoodiesNotAListError(testing *testing.T) {
	goodies := NewGoodies(25*time.Millisecond, "", 0)
	goodies.Set("value", "1", ExpireNever)
	err := goodies.ListPush("value", "1")
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
	_, err = goodies.Get(key)
	if _, ok := err.(NotFoundError); !ok {
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
	if _, ok := err6.(NotFoundError); !ok {
		testing.Error("Unexpected error when removing from non-existent list")
	}
}

func TestListRemoveByValue(testing *testing.T) {
	goodies := NewGoodies(ExpireNever, "", 0)
	key := "list"
	s1 := "s1"
	s2 := "s2"
	s3 := "s3"

	goodies.ListPush(key, s1)
	goodies.ListPush(key, s2)
	goodies.ListPush(key, s3)
	goodies.ListPush(key, s1)
	goodies.ListPush(key, s2)
	goodies.ListPush(key, s3)

	len, err := goodies.ListLen(key)
	if err != nil || len != 6 {
		testing.Error("List was created incorrectly")
	}
	err = goodies.ListRemoveValue(key, "s1")
	len, err = goodies.ListLen(key)
	if err != nil || len != 4 {
		testing.Error("List deletion failed")
		return
	}
	err = goodies.ListRemoveValue(key, s2)
	len, err = goodies.ListLen(key)
	if err != nil || len != 2 {
		testing.Error("List deletion of struct failed")
		return
	}
	err = goodies.ListRemoveValue(key, s3)
	len, err = goodies.ListLen(key)
	if err != nil || len != 0 {
		testing.Error("List deletion of integer failed")
		return
	}
	err = goodies.ListRemoveValue(key, "Non-existent")
	if err != nil {
		testing.Errorf("Unexpected error on removing inexistent value from list by value: %v", err)
	}

	goodies.Set("valkey", "3.14", ExpireDefault)
	err = goodies.ListRemoveValue("valkey", "value")
	if err == nil {
		testing.Error("Expected error on removing by value from non list")
	}
}

func TestDictOps(testing *testing.T) {
	goodies := NewGoodies(ExpireNever, "", 0)
	// key := "dict"
	// dictKey := "dictKey"
	goodies.Set("val", "1", ExpireDefault)
	err := goodies.DictSet("val", "val", "1")
	if err == nil {
		testing.Error("Expected error on setting dict value for a non dict item")
	}

	// goodies.DictSet(key, dictKey, 3.14)
	// var f float32

	// fval, err2 := goodies.DictGet(key, dictKey)
	// if err2 != nil {
	// 	testing.Error("Dictionary set/get didn't work")
	// }
	// //f = fval.(float32)
	// if f != 3.14 {
	// 	testing.Error("Dictionary set/get didn't return expected value")
	// }
}
