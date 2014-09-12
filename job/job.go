package job

import (
  "time"
  "fmt"
  "encoding/json"
  "io/ioutil"
  "io"
  "os"
  "strings"
)

type Check struct {
  Script string
  Interval string
  TTL string
}

type Service struct {
  ID string
  Name string
  Tags []string
  Port int
  Check Check
}

type Job struct {
  ID, Name, Command, Output, OutputErrors string
  Type string // default is "shell"
  NoWait bool
  StartTime int64
  EndTime int64
  StartTimeStr string
  EndTimeStr string
  ExecutionNode string
  Timeout time.Duration //Timeout in seconds
  ExitErrors string
  Services []Service
}

func LoadJobFromFile(file_name string) (job Job, err error) {
  file, e := ioutil.ReadFile(file_name)
  if e != nil {
    fmt.Printf("File error: %v\n", e)
    os.Exit(1)
  }

  dec := json.NewDecoder(strings.NewReader(string(file)))
  //for {
  var j Job
  if err := dec.Decode(&j); err == io.EOF {
      //break
    return j, err
  } else if err != nil {
      return j, err
  }
  //}
  return j, nil
}

func (j *Job) String() string {
  en, _ := json.Marshal(j)
  return string(en)
}

func (j *Job) Encode() []byte {
  en, _ := json.Marshal(j)
  return en
}
