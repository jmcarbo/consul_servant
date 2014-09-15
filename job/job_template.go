package job

import (
  "text/template"
  "log"
  "bytes"
  "strings"
)

type JobTemplate struct {
  ID  string
  Name  string
  Template string
}

func NewJobTemplate (name, template string) *JobTemplate {
  return &JobTemplate{ ID: name, Name: name, Template: template }
}

func (jt *JobTemplate) Instantiate() (command string, err error) {

  funcMap := template.FuncMap{
    // The name "title" is what the function will be called in the template text.
    "title": strings.Title,
  }

  t, err := template.New(jt.Name).Funcs(funcMap).Parse(jt.Template)
  if err != nil {
    log.Fatalf("parsing: %s", err)
  }

  var doc bytes.Buffer 
  err = t.Execute(&doc, "the go programming language")
  if err != nil {
    log.Fatalf("execution: %s", err)
  }
  command = doc.String()
  return
}
