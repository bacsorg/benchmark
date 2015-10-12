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
var iterations = flag.Int("iterations", 1, "Number of iterations")
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

    var waitLogin, waitFinish sync.WaitGroup
    waitLogin.Add(*jobs)
    waitFinish.Add(*jobs)
    for i := 0; i < *jobs; i++ {
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
            log.Printf("Logged in %d: %v", id, time.Since(start))
            start = time.Now()
            err = client.EnterContest(*contestId)
            if err != nil {
                log.Fatal(err)
            }
            log.Printf("Entered contest %d: %v", id, time.Since(start))

            waitLogin.Done()
            waitLogin.Wait()
            start = time.Now()

            for i := 0; i < *iterations; i++ {
                _, err = client.AcmMonitor()
                if err != nil {
                    log.Fatal(err)
                }
            }
            fmt.Printf("id %d: %v\n", id, time.Since(start))
            waitFinish.Done()
        }(i)
    }
    waitFinish.Wait()
}
