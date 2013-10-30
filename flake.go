package main
import (
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "time"
    "sync"
)

const (
    nano = 1000 * 1000
)

var (
    epoch = uint64(time.Date(2013, 10, 20, 0, 0, 0, 0, time.UTC).UnixNano() / nano)
    errClockBackwards = errors.New("The clock went backwards!")
    errSequenceOverflow = errors.New("Sequence Overflow!")
)

type Flake struct {
    maxTime uint64
    workerId uint64
    sequence uint64
    stats Stats
    lock sync.Mutex
}

type Stats struct {
    generatedIds uint64
    errors uint64
}

func NewFlake(workerId uint64) (*Flake, error) {
    flake := new(Flake)
    flake.maxTime = now()
    flake.workerId = workerId
    flake.sequence = 0
    flake.stats.generatedIds = 0
    flake.stats.errors = 0
    return flake, nil
}

func (flake *Flake) next(writer http.ResponseWriter, request *http.Request) {
    flake.lock.Lock()
    defer flake.lock.Unlock()
    currentTime := now()
    var flakeId uint64 = 0

    if currentTime < flake.maxTime {
        // Our clock is now behind, NTP is shifting the clock
        go func(){ flake.stats.errors += 1 }()
        http.Error(writer, errClockBackwards.Error(), http.StatusInternalServerError)
        return
    }

    if currentTime > flake.maxTime {
      flake.sequence = 0
      flake.maxTime = currentTime
    }

    flake.sequence += 1
    if flake.sequence > 4095 {
        // Sequence overflow
        go func(){ flake.stats.errors += 1 }()
        http.Error(writer, errSequenceOverflow.Error(), http.StatusInternalServerError)
        return
    }

    go func(){ flake.stats.generatedIds += 1 }()
    flakeId = ((currentTime - epoch) << 22) | (flake.workerId << 12) | flake.sequence

    fmt.Fprintf(writer, "%d", flakeId)

}

func (flake *Flake) getStats(writer http.ResponseWriter, request *http.Request){
    type StatsData struct {
        Timestamp uint64
        GeneratedIds uint64
        Errors uint64
        MaxTime uint64
        WorkerId uint64
    }
    stats := StatsData{now(), flake.stats.generatedIds, flake.stats.errors, flake.maxTime, flake.workerId}

    json_stats, err := json.Marshal(stats)
    if err != nil {
        http.Error(writer, err.Error(), http.StatusInternalServerError)
        return
    }

    writer.Header().Set("Content-Type", "application/json")
    writer.Write(json_stats)
}

func now() (uint64) {
    return uint64(time.Now().UnixNano() / nano)
}

func main() {
   var flake, err = NewFlake(0)
   if err != nil {
      fmt.Println("Could not instanciate new Flake generator", err.Error)
      return // exit
   }

    http.HandleFunc("/", flake.next)
    http.HandleFunc("/stats", flake.getStats)

    server := &http.Server{
      Addr: ":8080",
    }
    server.ListenAndServe()

}
