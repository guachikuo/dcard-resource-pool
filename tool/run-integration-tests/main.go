package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pl "github.com/guachikuo/dcard-resource-pool"
)

const (
	maxIdleSize = 20
	maxIdleTime = time.Duration(5 * time.Second)

	remoteAddr = "127.0.0.1:9999"
)

func runMockServer(ln net.Listener) {
	fmt.Println("------- start running mock TCP server -------")

	for {
		ln.Accept()
	}
}

func main() {
	ctx := context.Background()

	ln, err := net.Listen("tcp", remoteAddr)
	if err != nil {
		log.Fatal(err)
		return
	}

	go runMockServer(ln)

	time.Sleep(100 * time.Millisecond)

	// ------- start creating TCP connection pool------- //

	fmt.Println("------- start creating TCP connection pool -------")

	pool, err := pl.New(
		func(ctx context.Context) (net.Conn, error) {
			conn, err := net.Dial("tcp", remoteAddr)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		func(ctx context.Context, conn net.Conn) {
			conn.Close()
		},
		maxIdleSize,
		maxIdleTime,
	)
	if err != nil {
		log.Fatal(err)
		return
	}

	time.Sleep(100 * time.Millisecond)

	// ------- simple test ------- //

	fmt.Println("------- start simple test first -------")

	if num := pool.NumIdle(); num != maxIdleSize {
		log.Fatal(fmt.Errorf("unexpected NumIdle : %d", num))
		return
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	if conn == nil || conn.RemoteAddr().String() != remoteAddr {
		log.Fatal(fmt.Errorf("unexpected connection"))
		return
	}

	if num := pool.NumIdle(); num != maxIdleSize-1 {
		log.Fatal(fmt.Errorf("unexpected NumIdle : %d", num))
		return
	}

	pool.Release(ctx, conn)

	if num := pool.NumIdle(); num != maxIdleSize {
		log.Fatal(fmt.Errorf("unexpected NumIdle : %d", num))
		return
	}

	fmt.Println("success")

	time.Sleep(100 * time.Millisecond)

	// ------- integration test ------- //

	fmt.Println("------- start integration test -------")

	fn := func(i int, wg *sync.WaitGroup) {
		defer wg.Done()

		conn, err := pool.Acquire(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}
		if conn == nil || conn.RemoteAddr().String() != remoteAddr {
			log.Fatal(fmt.Errorf("unexpected connection"))
			return
		}

		if i%3 == 0 {
			time.Sleep(100 * time.Millisecond)
			pool.Release(ctx, conn)
		}
	}

	fmt.Println("test 1")

	wg := new(sync.WaitGroup)
	goCnt := 17
	for i := 0; i < goCnt; i++ {
		wg.Add(1)
		go fn(i, wg)
	}
	wg.Wait()

	// acquire but not release : 17*(2/3) -> 11
	// acquire and release : 17*(1/3) -> 6
	// and there are still 3 connections in the pool
	if num := pool.NumIdle(); num != 9 {
		log.Fatal(fmt.Errorf("unexpected NumIdle : %d", num))
		return
	}

	fmt.Println("success")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("test 2")

	wg = new(sync.WaitGroup)
	goCnt = 100
	for i := 0; i < goCnt; i++ {
		wg.Add(1)
		go fn(i, wg)
	}
	wg.Wait()

	// acquire but not release : 100*(2/3)
	// acquire and release : 100*(1/3), but only 20 resource would be in the pool
	if num := pool.NumIdle(); num != 20 {
		log.Fatal(fmt.Errorf("unexpected NumIdle : %d", num))
		return
	}

	fmt.Println("success")

	return
}
