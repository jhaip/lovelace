package main

import (
  "bufio"
	"fmt"
  "time"
  "os"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "net/url"
	"github.com/jung-kurt/gofpdf"
)

const URL = "http://localhost:3000/"

func say(fact string) {
  formData := url.Values{
		"facts": {fact},
	}
  resp, err := http.PostForm(URL + "assert", formData)
	if err != nil {
		fmt.Println(err)
	}
  defer resp.Body.Close()
}

func retract(fact string) {
  formData := url.Values{
		"facts": {fact},
	}
  resp, err := http.PostForm(URL + "retract", formData)
	if err != nil {
		fmt.Println(err)
	}
  defer resp.Body.Close()
}

type SampleId struct {
  Value int
}

type SampleFilename struct {
  Value string
}

type Sample struct {
  Id SampleId
  ShortFilename SampleFilename
}

func selectt(fact string) []Sample {
  formData := url.Values{
		"facts": {fact},
	}

  resp, err := http.PostForm(URL + "select", formData)
  fmt.Println(URL + "select")
  fmt.Println(formData)
	if err != nil {
		fmt.Println(err)
	}

  fmt.Println(resp)
  fmt.Println(resp.Body)

  body, err := ioutil.ReadAll(resp.Body)
  fmt.Println(body)
  bodyString := string(body)
  fmt.Println(bodyString)

  bodyStringBytes := []byte(bodyString)

  samples := make([]Sample,0)
  json.Unmarshal(bodyStringBytes, &samples)

  defer resp.Body.Close()

  return samples
}

func readLines(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }
  return lines, scanner.Err()
}

func setFillColorFromDotCode(pdf *gofpdf.Fpdf, codeChar byte) {
  if (codeChar == '0') {
    pdf.SetFillColor(255, 0, 0)
  } else if (codeChar == '1') {
    pdf.SetFillColor(0, 255, 0)
  } else if (codeChar == '2') {
    pdf.SetFillColor(0, 0, 255)
  } else {
    pdf.SetFillColor(0, 0, 0)
  }
}

func generatePrintFile(sourceCode string, programId int, name string, code8400 []string) {
  pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetAutoPageBreak(false, 0)

	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, topMargin, rightMargin, bottomMargin := pdf.GetMargins()
	circleRadius := 7.0
	circleSpacing := circleRadius * 2.0 * 1.3
	circleMargin := 10 + circleRadius

	pdf.SetFont("Courier", "", 8)
	pdf.ClipRect(circleMargin+leftMargin, circleMargin+topMargin, pageWidth-circleMargin*2-leftMargin-rightMargin, pageHeight-topMargin*2-bottomMargin-circleMargin*2, true)
	pdf.TransformBegin()
	pdf.TransformTranslate(circleMargin, circleMargin)
	pdf.MultiCell(pageWidth-circleMargin*2-leftMargin-rightMargin, 4, sourceCode, "", "", false)
	pdf.TransformEnd()
	pdf.ClipEnd()

  for i := 0; i < 4; i++ {
  	pdf.TransformBegin()
  	pdf.SetFillColor(0, 0, 0)
    if (i == 0) {
	    pdf.TransformTranslate(circleMargin+0, circleMargin+0)
    } else if (i == 1) {
      pdf.TransformTranslate(circleMargin, pageHeight-circleMargin)
    } else if (i == 2) {
      pdf.TransformTranslate(pageWidth-circleMargin, pageHeight-circleMargin)
    } else {
      pdf.TransformTranslate(pageWidth-circleMargin, circleMargin)
    }
  	pdf.TransformRotate(90.0 * float64(i), 0, 0)
    code := code8400[i*(8400/4) + programId]
    setFillColorFromDotCode(pdf, code[0])
  	pdf.Circle(circleSpacing*0.0, circleSpacing*3.0, circleRadius, "F")
  	setFillColorFromDotCode(pdf, code[1])
  	pdf.Circle(circleSpacing*0.0, circleSpacing*2.0, circleRadius, "F")
  	setFillColorFromDotCode(pdf, code[2])
  	pdf.Circle(circleSpacing*0.0, circleSpacing*1.0, circleRadius, "F")
  	setFillColorFromDotCode(pdf, code[3])
  	pdf.Circle(circleSpacing*0.0, circleSpacing*0.0, circleRadius, "F")
  	setFillColorFromDotCode(pdf, code[4])
  	pdf.Circle(circleSpacing*1.0, circleSpacing*0.0, circleRadius, "F")
  	setFillColorFromDotCode(pdf, code[5])
  	pdf.Circle(circleSpacing*2.0, circleSpacing*0.0, circleRadius, "F")
    setFillColorFromDotCode(pdf, code[6])
  	pdf.Circle(circleSpacing*3.0, circleSpacing*0.0, circleRadius, "F")
  	pdf.TransformEnd()
  }

  pdf.SetFont("Courier", "B", 10)
  pdf.SetXY(0, pageHeight - topMargin * 2 - bottomMargin - circleRadius)
  pdf.WriteAligned(0, 20, name, "C")

	err := pdf.OutputFileAndClose("hello.pdf")
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
  code8400, err := readLines("/Users/jhaip/Code/lovelace/src/standalone_processes/files/dot-codes.txt")
  if err != nil {
		fmt.Println(err)
    return
	}
  fmt.Println(code8400)
  for {
    samples := selectt("wish paper $id at $shortFilename would be printed")
    for _, sample := range samples {
      fmt.Printf("%#v\n", sample)
      fmt.Printf("%#v\n", sample.Id.Value)
      fmt.Printf("%#v\n", sample.ShortFilename.Value)
      programId := sample.Id.Value
      shortFilename := sample.ShortFilename.Value
      // get sourceCode

    	sourceCode := "const Room = require('@living-room/client-js')\nconst execFile = require('child_process').execFile;\n\nconst room = new Room()\n\nroom.subscribe(\n  `wish $name would be running`,\n  ({assertions, retractions}) => {\n    retractions.forEach(async ({ name }) => {\n      const existing_pid = await room.select(`☻${name}☻ has process id $pid`)\n      console.error(`making ${name} NOT be running`)\n      console.error(existing_pid)\n      existing_pid.forEach(({ pid }) => {\n        pid = pid.value;\n        console.log(☻STOPPING PID☻, pid)\n        process.kill(pid, 'SIGTERM')\n        room.retract(`☻${name}☻ has process id $`);\n        room.retract(`☻${name}☻ is active`);\n      })\n    })\n    assertions.forEach(async ({ name }) => {\n      const existing_pid = await room.select(`☻${name}☻ has process id $pid`)\n      if (existing_pid.length === 0) {\n        console.error(`making ${name} be running!`)\n        let languageProcess = 'node'\n        let programSource = `src/standalone_processes/${name}`\n        if (name.includes('.py')) {\n          console.error(☻running as Python!☻)\n          languageProcess = 'python3'\n        }\n        const child = execFile(\n          languageProcess,\n          [programSource],\n          (error, stdout, stderr) => {\n            // TODO: check if program should still be running\n            // and start it again if so.\n            room.retract(`☻${name}☻ has process id $`);\n            room.retract(`☻${name}☻ is active`);\n            console.log(`${name} callback`)\n            if (error) {\n                console.error('stderr', stderr);\n            }\n            console.log('stdout', stdout);\n        });\n        const pid = child.pid;\n        room.assert(`☻${name}☻ has process id ${pid}`);\n        console.error(pid);\n      }\n    })\n  }\n)\n\nroom.assert('wish ☻390__initialProgramCode.js☻ would be running')\nroom.assert('wish ☻498__printingManager.py☻ would be running')\nroom.assert('wish ☻577__programEditor.js☻ would be running')\nroom.assert('wish ☻826__runSeenPapers.js☻ would be running')\nroom.assert('wish ☻277__pointingAt.py☻ would be running')\nroom.assert('wish ☻620__paperDetails.js☻ would be running')\n\nroom.assert(`camera 1 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)\n\n// room.assert(`camera 1 sees paper 1924 at TL (100, 100) TR (1200, 100) BR (1200, 800) BL (100, 800) @ 1`)\n// room.assert(`paper 1924 is pointing at paper 472`)  // comment out if pointingAt.py is running\n"

      generatePrintFile(sourceCode, programId, shortFilename, code8400)
    }
    // select
    // wish paper $id at $shortFilename would be printed
    // every second or so
    // later:
    // room.assert(`wish file Y would be printed`)
    time.Sleep(10 * time.Millisecond)
    break
  }

}
