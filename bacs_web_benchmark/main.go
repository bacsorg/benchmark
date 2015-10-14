package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "io/ioutil"
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
var scenario = flag.String("scenario", "AcmMonitor", "Name of scenario to run")
var jobsConfiguration = flag.String("jobs-config", "", "Configuration file for jobs")

type JobConfiguration struct {
    Username string
    Password string
}

type JobsConfiguration []JobConfiguration

type JobResult struct {
    fails               map[string]int
    iterDurationSeconds float64
}

type Scenario func(client *benchmark.WebClient) error

func AcmMonitor(client *benchmark.WebClient) error {
    _, err := client.AcmMonitor()
    return err
}

func SchoolFinalMonitor(client *benchmark.WebClient) error {
    _, err := client.SchoolFinalMonitor()
    return err
}

func MySchoolFinalSubmits(client *benchmark.WebClient) error {
    _, err := client.MySchoolFinalSubmits()
    return err
}

func SubmitA(client *benchmark.WebClient) error {
    return client.Submit("A", "C++ 11", `#include <iostream>
                                         int main() {
                                           int a, b;
                                           std::cin >> a >> b;
                                           std::cout << a + b << std::endl;
                                         }`)
}

var scenarios = map[string]Scenario{
    "AcmMonitor":           AcmMonitor,
    "SchoolFinalMonitor":   SchoolFinalMonitor,
    "MySchoolFinalSubmits": MySchoolFinalSubmits,
    "SubmitA":              SubmitA,
}

func main() {
    flag.Parse()

    var config JobsConfiguration

    if *jobsConfiguration == "" {
        if *password == "-" {
            fmt.Printf("password: ")
            *password = string(gopass.GetPasswd())
        }
        config = make(JobsConfiguration, *jobs)
        for i, _ := range config {
            config[i].Username = *username
            config[i].Password = *password
        }
    } else {
        data, err := ioutil.ReadFile(*jobsConfiguration)
        if err != nil {
            log.Fatal(err)
        }
        err = json.Unmarshal(data, &config)
        if err != nil {
            log.Fatal(err)
        }
        config = config[0:*jobs]
    }
    scen := scenarios[*scenario]

    var waitLogin sync.WaitGroup
    waitLogin.Add(*jobs)
    results := make(chan JobResult)
    for i := 0; i < *jobs; i++ {
        go func(id int) {
            start := time.Now()
            client, err := benchmark.NewWebClient(*url)
            if err != nil {
                log.Fatal(err)
            }

            err = client.Login(config[id].Username, config[id].Password)
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
            var jobResult JobResult
            jobResult.fails = make(map[string]int)
            failsTotal := 0

            for i := 0; i < *iterations; i++ {
                err = scen(client)
                if err != nil {
                    jobResult.fails[err.Error()]++
                    failsTotal++
                }
            }
            totalDuration := time.Since(start)
            log.Printf("Scenario %d: %v", id, totalDuration)

            iterDurationNanos := totalDuration.Nanoseconds() / int64(*iterations)
            jobResult.iterDurationSeconds =
                float64(iterDurationNanos) / (1000 * 1000 * 1000)
            results <- jobResult
        }(i)
    }
    var globalResult JobResult
    globalResult.fails = make(map[string]int)
    failsTotal := 0
    for i := 0; i < *jobs; i++ {
        result := <-results
        for key, value := range result.fails {
            globalResult.fails[key] += value
            failsTotal += value
        }
        globalResult.iterDurationSeconds += result.iterDurationSeconds / float64(*jobs)
    }
    log.Printf("Average scenario execution time: %fs", globalResult.iterDurationSeconds)
    log.Printf("Failed %d/%d (%02d%%)", failsTotal, *jobs**iterations,
        100*failsTotal/(*jobs**iterations))
    for key, value := range globalResult.fails {
        log.Printf("Failure %q: %d times", key, value)
    }
}
