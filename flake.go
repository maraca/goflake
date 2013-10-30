package main
import (
    "errors"
    "fmt"
    "time"
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
}

func NewFlake(workerId uint64) (*Flake, error) {
    flake := new(Flake)
    flake.maxTime = now()
    flake.workerId = workerId
    flake.sequence = 0
    return flake, nil
}

func (flake *Flake) Next() (uint64, error) {
    currentTime := now()
    var flakeId uint64 = 0

    if currentTime < flake.maxTime {
        // Our clock is now behind, NTP is shifting the clock
        return flakeId, errClockBackwards
    }

    if currentTime > flake.maxTime {
      flake.sequence = 0
      flake.maxTime = currentTime
    }

    flake.sequence += 1
    if flake.sequence > 4095 {
        // Sequence overflow
        return flakeId, errSequenceOverflow
    }

    flakeId = ((currentTime - epoch) << 22) | (flake.workerId << 12) | flake.sequence

    return flakeId, nil
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
   for {
     uuid, err := flake.Next()
     if err != nil {
        fmt.Println("Could not get a new flake id")
     }
     time.Sleep(1e9)
     fmt.Println(uuid)
   }
}
