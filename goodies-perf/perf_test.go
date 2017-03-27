package goodiesPerf

import (
	"goodies/goodies"
	"testing"
)

func BenchmarkGoodiesClientSet(b *testing.B) {
	// server := goodies.NewGoodiesHttpServer("9006", 1*time.Minute, "./goodies.dat", 30*time.Second)
	// fmt.Println("Before server")
	// go server.ListenAndServe()
	// fmt.Println("Serve started")
	client := goodies.NewGoodiesClient("http://127.0.0.1:9006/")
	for i := 0; i < b.N; i++ {
		client.Set(string(i), string(i), goodies.ExpireDefault)
	}
	// server.Close()
}

func BenchmarkGoodiesClientGet(b *testing.B) {
	// server := goodies.NewGoodiesHttpServer("9006", 1*time.Minute, "./goodies.dat", 30*time.Second)
	// fmt.Println("Before server")
	// go server.ListenAndServe()
	// fmt.Println("Serve started")
	client := goodies.NewGoodiesClient("http://127.0.0.1:9006/")
	for i := 0; i < b.N; i++ {
		client.Get(string(i))
	}
	// server.Close()
}
