package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type counter map[uint8]int

func main() {
	now := time.Now()
	defer func() {
		fmt.Println(time.Since(now))
	}()

	max := runtime.GOMAXPROCS(0)
	var ch chan []byte = make(chan []byte)
	var wg sync.WaitGroup
	wg.Add(max)
	var counters = make([]counter,max)
	for i := 0; i <= max-1; i++ {
		c := counter{}
		counters[i] = c;
		go func() {
			defer wg.Done()
			for chunk := range ch {
				process(c, chunk)
			}
		}()
	}
	err := read("./files", 1024, ch)
	if err != nil{
		log.Fatal(err)
	}
	wg.Wait()

	counter := merge(counters)
	minimax := norm(counter)
	plot := plot(counter, minimax)
	fmt.Println(plot)
}

func read(folder string, chunkSize int, ch chan []byte) (error) {
	defer close(ch)
	chunk := []byte{}
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		r := bufio.NewReader(f)
		for {
			buf := make([]byte, 0, chunkSize)
			n, err := r.Read(buf[:cap(buf)])
			buf = buf[:n]
			if n == 0 {
				if err == nil {
					continue
				}
				if err == io.EOF {
					break
				}
			}
			chunk = append(chunk, buf...)
			if len(chunk) >= chunkSize {
				ch <- chunk
				chunk = []byte{}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(chunk) != 0 {
		ch <- chunk
	}
	return nil
}

func process (counter counter, chunk []byte){
	for _,v := range chunk{
			counter[v] += 1
	}
}

func merge(counters []counter) map[uint8]int {
	result := make(counter)
	for _, x := range counters {
		for k, v := range x {
			result[k] += v
		}
	}
	return result
}

func norm(counter counter) counter {
	res := make(map[uint8]int,len(counter))
	var max =-(math.MaxInt32 << 1)+1;
	var min =+(math.MaxInt32);
	for _,v := range counter {
		if (v > max) {max = v};
		if (v < min) {min = v};
	}
	var dx  = (max - min)
	for k,v := range counter {
		res[k] = (v-min) * 100 / dx
	}
	return res;
}

func plot(c counter, norm counter) string {
	nice := func(a uint8) (string){
		switch a {
		case 10:
			return "LF"
		case 13:
			return "CR"
		case 32:
			return "SB"
		default:
			return "" + string(rune(a)) + " "
		}
	}
	srt := func(c counter) ([]uint8) {
		result := []uint8{}
		for k := range c {
			result = append(result, k)
		}
		sort.Slice(result, func(i, j int) bool {
			return c[result[uint8(i)]] > c[result[uint8(j)]]
		})
		return result
	}
	keys := srt(c)
	builder := strings.Builder{}
	for _,k := range keys {
		v := c[k]
		n := norm[k]
		gauge := strings.Repeat(".", n / 2);
		x := fmt.Sprintf("%v| %v [%v]\n",nice(k), gauge , v)
		builder.WriteString(x)
	}
	return builder.String()
}

