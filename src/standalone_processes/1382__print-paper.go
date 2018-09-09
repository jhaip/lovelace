package main

import (
  "bufio"
  "time"
  "os"
  "log"
  "fmt"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "net/url"
  "strconv"
	"github.com/jung-kurt/gofpdf"
)

const URL = "http://localhost:3000/"
// const BASE_PATH = "/Users/jhaip/Code/lovelace/src/standalone_processes/"
const BASE_PATH = "/home/jacob/lovelace/src/standalone_processes/"
const DOT_CODES_PATH = BASE_PATH + "files/dot-codes.txt"
const PDF_OUTPUT_FOLDER = BASE_PATH + "files/"
const LOG_PATH = BASE_PATH + "logs/1382__print-paper.log"
const MY_ID = "1382";

func say(fact string) {
  formData := url.Values{
		"facts": {fmt.Sprintf("#%s %s", MY_ID, fact)},
	}
  resp, err := http.PostForm(URL + "assert", formData)
	if err != nil {
		log.Println(err)
	}
  defer resp.Body.Close()
}

func retract(fact string, targetPaper string) {
  formData := url.Values{
		"facts": {fmt.Sprintf("%s %s", targetPaper, fact)},
	}
  resp, err := http.PostForm(URL + "retract", formData)
	if err != nil {
		log.Println(err)
	}
  defer resp.Body.Close()
}

type SampleId struct {
  Value int
}

type SampleValue struct {
  Value string
}

type Sample struct {
  Id SampleId
  ShortFilename SampleValue
}

type SourceCodeSample struct {
  SourceCode SampleValue
}

func selectt(fact string) []byte {
  formData := url.Values{
		"facts": {fmt.Sprintf("$ %s", fact)},
	}

  resp, err := http.PostForm(URL + "select", formData)
  log.Println(URL + "select")
  log.Println(formData)
	if err != nil {
		log.Println(err)
	}

  log.Println(resp)
  log.Println(resp.Body)

  body, err := ioutil.ReadAll(resp.Body)
  log.Println(body)
  bodyString := string(body)
  log.Println("bodyString ---")
  log.Println(bodyString)

  bodyStringBytes := []byte(bodyString)


  defer resp.Body.Close()

  return bodyStringBytes
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
  pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.AddPage()

	pdf.SetAutoPageBreak(false, 0)

	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, topMargin, rightMargin, bottomMargin := pdf.GetMargins()
	circleRadius := 8.0
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
    } else if (i == 3) {
      pdf.TransformTranslate(circleMargin, pageHeight-circleMargin)
    } else if (i == 2) {
      pdf.TransformTranslate(pageWidth-circleMargin, pageHeight-circleMargin)
    } else {
      pdf.TransformTranslate(pageWidth-circleMargin, circleMargin)
    }
  	pdf.TransformRotate(-90.0 * float64(i), 0, 0)
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

	err := pdf.OutputFileAndClose(PDF_OUTPUT_FOLDER + strconv.Itoa(programId) + ".pdf")
	if err != nil {
		log.Println(err)
	}
}

func get_wishes() []Sample {
  bodyStringBytes := selectt("wish paper $id at $shortFilename would be printed")
  samples := make([]Sample,0)
  json.Unmarshal(bodyStringBytes, &samples)
  log.Println("GET WISHES -------------")
  log.Println(samples)
  return samples
}

func get_source_code(shortFilename string) (string, bool) {
  bodyStringBytes := selectt("\"" + shortFilename + "\" has source code $sourceCode")
  log.Println(bodyStringBytes)
  samples := make([]SourceCodeSample,0)
  json.Unmarshal(bodyStringBytes, &samples)
  log.Println("GET SOURCE CODE -----------")
  log.Println(samples)
  if (len(samples) > 0) {
    log.Println(samples[0])
    log.Println(samples[0].SourceCode.Value)
    log.Println("^^^^^^^^^")
    return samples[0].SourceCode.Value, true
  }
  return "", false
}

func main() {
  /*** Set up logging ***/
  f, err := os.OpenFile(LOG_PATH, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
      log.Fatalf("error opening file: %v", err)
  }
  defer f.Close()

  log.SetOutput(f)
  // log.Println("This is a test log entry")
  /*** /end logging setup ***/

  code8400, err := readLines(DOT_CODES_PATH)
  if err != nil {
		log.Println(err)
    return
	}
  for {
    samples := get_wishes()
    for _, sample := range samples {

      log.Printf("%#v\n", sample)
      log.Println("PROGRAM ID:::")
      log.Printf("%#v\n", sample.Id.Value)
      log.Printf("%#v\n", sample.ShortFilename.Value)
      programId := sample.Id.Value
      retract("wish paper " + strconv.Itoa(programId) + " at $ would be printed", "$")
      shortFilename := sample.ShortFilename.Value
      sourceCode, foundSourceCode := get_source_code(shortFilename)
      if (foundSourceCode == false) {
        log.Println("SOURCE CODE NOT FOUND!")
        return
      }
      /*
    	sourceCode := "const Room = require('@living-room/client-js')\nconst execFile = require('child_process').execFile;\n\nconst room = new Room()\n\nroom.subscribe(\n  `wish $name would be running`,\n  ({assertions, retractions}) => {\n    retractions.forEach(async ({ name }) => {\n      const existing_pid = await room.select(`☻${name}☻ has process id $pid`)\n      console.error(`making ${name} NOT be running`)\n      console.error(existing_pid)\n      existing_pid.forEach(({ pid }) => {\n        pid = pid.value;\n        console.log(☻STOPPING PID☻, pid)\n        process.kill(pid, 'SIGTERM')\n        room.retract(`☻${name}☻ has process id $`);\n        room.retract(`☻${name}☻ is active`);\n      })\n    })\n    assertions.forEach(async ({ name }) => {\n      const existing_pid = await room.select(`☻${name}☻ has process id $pid`)\n      if (existing_pid.length === 0) {\n        console.error(`making ${name} be running!`)\n        let languageProcess = 'node'\n        let programSource = `src/standalone_processes/${name}`\n        if (name.includes('.py')) {\n          console.error(☻running as Python!☻)\n          languageProcess = 'python3'\n        }\n        const child = execFile(\n          languageProcess,\n          [programSource],\n          (error, stdout, stderr) => {\n            // TODO: check if program should still be running\n            // and start it again if so.\n            room.retract(`☻${name}☻ has process id $`);\n            room.retract(`☻${name}☻ is active`);\n            console.log(`${name} callback`)\n            if (error) {\n                console.error('stderr', stderr);\n            }\n            console.log('stdout', stdout);\n        });\n        const pid = child.pid;\n        room.assert(`☻${name}☻ has process id ${pid}`);\n        console.error(pid);\n      }\n    })\n  }\n)\n\nroom.assert('wish ☻390__initialProgramCode.js☻ would be running')\nroom.assert('wish ☻498__printingManager.py☻ would be running')\nroom.assert('wish ☻577__programEditor.js☻ would be running')\nroom.assert('wish ☻826__runSeenPapers.js☻ would be running')\nroom.assert('wish ☻277__pointingAt.py☻ would be running')\nroom.assert('wish ☻620__paperDetails.js☻ would be running')\n\nroom.assert(`camera 1 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)\n\n// room.assert(`camera 1 sees paper 1924 at TL (100, 100) TR (1200, 100) BR (1200, 800) BL (100, 800) @ 1`)\n// room.assert(`paper 1924 is pointing at paper 472`)  // comment out if pointingAt.py is running\n"
      */
      generatePrintFile(sourceCode, programId, shortFilename, code8400)
      say("wish file \"" + PDF_OUTPUT_FOLDER + strconv.Itoa(programId) + ".pdf" + "\" would be printed")
    }
    time.Sleep(100 * time.Millisecond)
    // break
  }

}
