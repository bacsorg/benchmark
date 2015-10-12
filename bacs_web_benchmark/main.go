package main

import (
    "flag"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/bacsorg/benchmark"

    "github.com/howeyc/gopass"
)

var jobs = flag.Int("jobs", 1, "Number of parallel jobs")
var url = flag.String("bacs-url", "", "BACS URL")
var username = flag.String("username", "", "Username")
var password = flag.String("password", "-", "Password, - to read from stdin")
var contestId = flag.Int("contest-id", 0, "Contest ID to use")

func main() {
    flag.Parse()
    if *password == "-" {
        fmt.Printf("password: ")
        *password = string(gopass.GetPasswd())
    }

    var wait sync.WaitGroup
    for i := 0; i < *jobs; i++ {
        wait.Add(1)
        go func(id int) {
            start := time.Now()
            client, err := benchmark.NewWebClient(*url)
            if err != nil {
                log.Fatal(err)
            }

            err = client.Login(*username, *password)
            if err != nil {
                log.Fatal(err)
            }

            err = client.EnterContest(*contestId)
            if err != nil {
                log.Fatal(err)
            }
            _, err = client.AcmMonitor()
            if err != nil {
                log.Fatal(err)
            }
            elapsed := time.Since(start)
            fmt.Printf("id %d: %v\n", id, elapsed)
            wait.Done()
        }(i)
    }
    wait.Wait()
}
