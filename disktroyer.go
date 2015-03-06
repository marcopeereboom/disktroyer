package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	wg     sync.WaitGroup
	buffer []byte

	prng = rand.Reader
	done = false

	// settings
	maxDirs     *int
	maxFileSize *int
	maxFiles    *int
	root        *string
	verbose     *bool
	debug       *bool
)

func printf(format string, args ...interface{}) {
	if *verbose == false {
		return
	}
	log.Printf(format, args...)
}

func test(dir string) error {
	defer wg.Done()

	var err error
	run := 0
	for !done {
		printf("%v: starting run %v", dir, run)
		err = testFiles(dir)
		if err != nil {
			log.Printf("testFiles(%v): %v", dir, err)
			return err
		}
		printf("%v: finished run %v", dir, run)
		run++
	}

	return nil
}

func testFiles(dir string) error {
	// create target directories
	printf("%v: creating src and dst", dir)
	srcD := filepath.Join(dir, "src")
	dstD := filepath.Join(dir, "dst")
	err := os.MkdirAll(srcD, 0775)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dstD, 0775)
	if err != nil {
		return err
	}

	mrand.Seed(int64(time.Now().Nanosecond()))

	// create source files
	printf("%v: creating %v files", dir, maxFiles)
	for files := 0; files < *maxFiles; files++ {
		filename := filepath.Join(srcD, strconv.Itoa(files))
		f, err := os.OpenFile(filename,
			os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}

		// write crap to it
		max := mrand.Intn(*maxFileSize)
		_, err = f.Write(buffer[:max])
		if err != nil {
			return err
		}
		f.Close()
	}

	// move files, do readdir to stress
	srcFiles, err := ioutil.ReadDir(srcD)
	if err != nil {
		return err
	}
	printf("%v: moving %v files", dir, len(srcFiles))
	for _, v := range srcFiles {
		oldP := filepath.Join(srcD, v.Name())
		newP := filepath.Join(dstD, v.Name())
		err = os.Rename(oldP, newP)
		if err != nil {
			return err
		}
	}

	// delete files
	dstFiles, err := ioutil.ReadDir(dstD)
	if err != nil {
		return err
	}
	printf("%v: deleting %v files", dir, len(dstFiles))
	for _, v := range dstFiles {
		filename := filepath.Join(dstD, v.Name())
		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}

	// delete dirs
	printf("%v: deleting src and dst", dir)
	err = os.Remove(srcD)
	if err != nil {
		return err
	}
	err = os.Remove(dstD)
	if err != nil {
		return err
	}

	return nil
}

func _main() error {
	err := os.MkdirAll(*root, 0775)
	if err != nil {
		return err
	}

	// create random buffer
	buffer = make([]byte, *maxFileSize)
	log.Printf("filling random buffer")
	_, err = prng.Read(buffer)
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	log.Printf("starting test with:")
	log.Printf("\tmaxDirs    :\t%v", *maxDirs)
	log.Printf("\tmaxFiles   :\t%v", *maxFiles)
	log.Printf("\tmaxFileSize:\t%v", *maxFileSize)
	log.Printf("\troot       :\t%v", *root)
	log.Printf("press ctrl-C to exit")

	for i := 0; i < *maxDirs; i++ {
		wg.Add(1)
		go test(filepath.Join(*root, strconv.Itoa(i)))
	}

	wait := make(chan bool)
	go func() {
		<-c
		log.Printf("flushing...")
		done = true
		wg.Wait()
		wait <- true
	}()

	<-wait

	log.Printf("run ended succesfully")

	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	root = flag.String("root", "disktroyer", "root directory for test")
	maxDirs = flag.Int("maxdirs", 16, "number of working directories")
	maxFileSize = flag.Int("maxfilesize", 64*1024, "maximum file size")
	maxFiles = flag.Int("maxfiles", 100, "maximum number of files per directory")
	verbose = flag.Bool("verbose", false, "enable verbosity")
	debug = flag.Bool("debug", false, "enable golang pprof")
	flag.Parse()

	if *debug {
		go http.ListenAndServe("localhost:6060", nil)
	}

	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
