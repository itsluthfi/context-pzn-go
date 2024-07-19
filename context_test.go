package contextpzngo

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	background := context.Background() // dibuat di awal, isinya kosongan
	fmt.Println(background)

	todo := context.TODO() // sama2 kosongan, tapi yg ini requirementnya belum jelas
	fmt.Println(todo)
}

func TestContextWithValue(t *testing.T) {
	contextA := context.Background() // context yg udah dibuat sifatnya immutable, setiap ada perubahan dari context awal akan membuat context baru (child)

	// context parent-child konsepnya sama kayak inheritance di oop, parent nurun ke child, tapi child gabisa naik ke parent

	contextB := context.WithValue(contextA, "b", "B")
	contextC := context.WithValue(contextA, "c", "C")

	contextD := context.WithValue(contextB, "d", "D")
	contextE := context.WithValue(contextB, "e", "E")

	contextF := context.WithValue(contextC, "f", "F")

	contextG := context.WithValue(contextF, "g", "G")

	fmt.Println(contextA)
	fmt.Println(contextB)
	fmt.Println(contextC)
	fmt.Println(contextD)
	fmt.Println(contextE)
	fmt.Println(contextF)
	fmt.Println(contextG)

	fmt.Println(contextF.Value("f")) // dapet valuenya
	fmt.Println(contextF.Value("c")) // dapet value dari parentnya
	fmt.Println(contextF.Value("b")) // gadapet value karena beda parent

	fmt.Println(contextA.Value("b")) // parent gabisa ambil data dari child, bisanya child ke parent
}

func CreateCounter() chan int { // contoh goroutine leak (berjalan terus2an tanpa henti dan kita gabisa stop)
	destination := make(chan int)

	go func() {
		defer close(destination)
		counter := 1
		for {
			destination <- counter
			counter++
		}
	}()

	return destination
}

func TestContextWithoutCancel(t *testing.T) {
	fmt.Println("Total Goroutine", runtime.NumGoroutine()) // awalnya 2

	destination := CreateCounter()
	for n := range destination {
		fmt.Println("Counter", n)
		if n == 10 {
			break
		}
	}

	fmt.Println("Total Goroutine", runtime.NumGoroutine()) // outputnya 3, harusnya 2 kayak di awal, 1 lebihnya jalan terus walaupun programnya udah berhenti

	// goroutine leak bisa nyebabin konsumsi memori yang berlebihan dan bisa bikin sistem down kalo goroutine leaknya banyak
}

func CreateCounterWithCancel(ctx context.Context) chan int {
	destination := make(chan int)

	go func() {
		defer close(destination)
		counter := 1
		for {
			select {
			case <-ctx.Done():
				return
				// kalo pake break bakal stop select, return bakal stop perulangannya
			default:
				destination <- counter
				counter++
			}
		}
	}()

	return destination
}

func TestContextWithCancel(t *testing.T) {
	fmt.Println("Total Goroutine", runtime.NumGoroutine()) // 2

	parent := context.Background()
	ctx, cancel := context.WithCancel(parent)

	destination := CreateCounterWithCancel(ctx)

	fmt.Println("Total Goroutine", runtime.NumGoroutine()) // 3

	for n := range destination {
		fmt.Println("Counter", n)
		if n == 10 {
			break
		}
	}
	cancel() // mengirim sinyal cancel ke context

	time.Sleep(2 * time.Second) // kalo ga ditunggu ada kemungkinan ke cancel duluan sebelum goroutine berhenti jadi total goroutinenya masih 3

	fmt.Println("Total Goroutine", runtime.NumGoroutine()) // 2
}

func CreateCounterWithTimeout(ctx context.Context) chan int {
	destination := make(chan int)

	go func() {
		defer close(destination)
		counter := 1
		for {
			select {
			case <-ctx.Done():
				return
				// kalo pake break bakal stop select, return bakal stop perulangannya
			default:
				destination <- counter
				counter++
				time.Sleep(1 * time.Second) // simulasi kalo prosesnya lemot
			}
		}
	}()

	return destination
}

func TestContextWithTimeout(t *testing.T) {
	fmt.Println("Total Goroutine", runtime.NumGoroutine())

	parent := context.Background()
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel() // tetep pake cancel (tapi defer) biar kalo prosesnya selesai sebelum 5 detik goroutine yang berjalan distop

	destination := CreateCounterWithTimeout(ctx)

	fmt.Println("Total Goroutine", runtime.NumGoroutine())

	for n := range destination {
		fmt.Println("Counter", n)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("Total Goroutine", runtime.NumGoroutine())
}

func CreateCounterWithDeadline(ctx context.Context) chan int {
	destination := make(chan int)

	go func() {
		defer close(destination)
		counter := 1
		for {
			select {
			case <-ctx.Done():
				return
				// kalo pake break bakal stop select, return bakal stop perulangannya
			default:
				destination <- counter
				counter++
				time.Sleep(1 * time.Second) // simulasi kalo prosesnya lemot
			}
		}
	}()

	return destination
}

func TestContextWithDeadline(t *testing.T) {
	fmt.Println("Total Goroutine", runtime.NumGoroutine())

	parent := context.Background()
	ctx, cancel := context.WithDeadline(parent, time.Now().Add(5*time.Second))
	// kalo deadline itu ditentukan by waktu/pukul, misal jam 10 malem, 12 siang, bukan berapa lama setelah proses berjalan
	defer cancel()

	destination := CreateCounterWithTimeout(ctx)

	fmt.Println("Total Goroutine", runtime.NumGoroutine())

	for n := range destination {
		fmt.Println("Counter", n)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("Total Goroutine", runtime.NumGoroutine())
}
