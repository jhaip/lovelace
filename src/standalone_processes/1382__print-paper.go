package main

import (
	"fmt"
  "encoding/json"
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

func selectt(fact string) {
  formData := url.Values{
		"facts": {fact},
	}
  // http://polyglot.ninja/golang-making-http-requests/
  // https://postman-echo.com/post
  resp, err := http.PostForm(URL + "select", formData)
  fmt.Println(URL + "select")
  fmt.Println(formData)
	if err != nil {
		fmt.Println(err)
	}

  fmt.Println(resp)

  var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println(result)
  defer resp.Body.Close()
  // return result
  // Example JSON:
  /*
  [
      {
          "id": {
              "value": 561
          },
          "shortFilename": {
              "value": "561.py"
          }
      }
  ]
  */
}

func main() {
  selectt("wish paper $id at $shortFilename would be printed")
  // select
  // wish paper $id at $shortFilename would be printed
  // every second or so
  // later:
  // room.assert(`wish file Y would be printed`)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetAutoPageBreak(false, 0)

	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, topMargin, rightMargin, bottomMargin := pdf.GetMargins()
	circleRadius := 7.0
	circleSpacing := circleRadius * 2.0 * 1.3
	circleMargin := 10 + circleRadius

	pdf.SetFont("Courier", "", 8)
	t := "const Room = require('@living-room/client-js')\nconst execFile = require('child_process').execFile;\n\nconst room = new Room()\n\nroom.subscribe(\n  `wish $name would be running`,\n  ({assertions, retractions}) => {\n    retractions.forEach(async ({ name }) => {\n      const existing_pid = await room.select(`☻${name}☻ has process id $pid`)\n      console.error(`making ${name} NOT be running`)\n      console.error(existing_pid)\n      existing_pid.forEach(({ pid }) => {\n        pid = pid.value;\n        console.log(☻STOPPING PID☻, pid)\n        process.kill(pid, 'SIGTERM')\n        room.retract(`☻${name}☻ has process id $`);\n        room.retract(`☻${name}☻ is active`);\n      })\n    })\n    assertions.forEach(async ({ name }) => {\n      const existing_pid = await room.select(`☻${name}☻ has process id $pid`)\n      if (existing_pid.length === 0) {\n        console.error(`making ${name} be running!`)\n        let languageProcess = 'node'\n        let programSource = `src/standalone_processes/${name}`\n        if (name.includes('.py')) {\n          console.error(☻running as Python!☻)\n          languageProcess = 'python3'\n        }\n        const child = execFile(\n          languageProcess,\n          [programSource],\n          (error, stdout, stderr) => {\n            // TODO: check if program should still be running\n            // and start it again if so.\n            room.retract(`☻${name}☻ has process id $`);\n            room.retract(`☻${name}☻ is active`);\n            console.log(`${name} callback`)\n            if (error) {\n                console.error('stderr', stderr);\n            }\n            console.log('stdout', stdout);\n        });\n        const pid = child.pid;\n        room.assert(`☻${name}☻ has process id ${pid}`);\n        console.error(pid);\n      }\n    })\n  }\n)\n\nroom.assert('wish ☻390__initialProgramCode.js☻ would be running')\nroom.assert('wish ☻498__printingManager.py☻ would be running')\nroom.assert('wish ☻577__programEditor.js☻ would be running')\nroom.assert('wish ☻826__runSeenPapers.js☻ would be running')\nroom.assert('wish ☻277__pointingAt.py☻ would be running')\nroom.assert('wish ☻620__paperDetails.js☻ would be running')\n\nroom.assert(`camera 1 has projector calibration TL (0, 0) TR (1920, 0) BR (1920, 1080) BL (0, 1080) @ 1`)\n\n// room.assert(`camera 1 sees paper 1924 at TL (100, 100) TR (1200, 100) BR (1200, 800) BL (100, 800) @ 1`)\n// room.assert(`paper 1924 is pointing at paper 472`)  // comment out if pointingAt.py is running\n"
	pdf.ClipRect(circleMargin+leftMargin, circleMargin+topMargin, pageWidth-circleMargin*2-leftMargin-rightMargin, pageHeight-topMargin*2-bottomMargin-circleMargin*2, true)
	pdf.TransformBegin()
	pdf.TransformTranslate(circleMargin, circleMargin)
	pdf.MultiCell(pageWidth-circleMargin*2-leftMargin-rightMargin, 4, t, "", "", false)
	pdf.TransformEnd()
	pdf.ClipEnd()

	pdf.TransformBegin()
	pdf.SetFillColor(0, 0, 0)
	pdf.TransformTranslate(circleMargin+0, circleMargin+0)
	pdf.TransformRotate(0, 0, 0)
  pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*3.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*2.0, circleRadius, "F")
	pdf.SetFillColor(0, 255, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*1.0, circleRadius, "F")
	pdf.SetFillColor(255, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*1.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*2.0, circleSpacing*0.0, circleRadius, "F")
  pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*3.0, circleSpacing*0.0, circleRadius, "F")
	pdf.TransformEnd()

	pdf.TransformBegin()
	pdf.SetFillColor(0, 0, 0)
	pdf.TransformTranslate(circleMargin, pageHeight-circleMargin)
	pdf.TransformRotate(90, 0, 0)
  pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*3.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*2.0, circleRadius, "F")
	pdf.SetFillColor(0, 255, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*1.0, circleRadius, "F")
	pdf.SetFillColor(255, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*1.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*2.0, circleSpacing*0.0, circleRadius, "F")
  pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*3.0, circleSpacing*0.0, circleRadius, "F")
	pdf.TransformEnd()

	pdf.TransformBegin()
	pdf.SetFillColor(0, 0, 0)
	pdf.TransformTranslate(pageWidth-circleMargin, pageHeight-circleMargin)
	pdf.TransformRotate(180, 0, 0)
  pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*3.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*2.0, circleRadius, "F")
	pdf.SetFillColor(0, 255, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*1.0, circleRadius, "F")
	pdf.SetFillColor(255, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*1.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*2.0, circleSpacing*0.0, circleRadius, "F")
  pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*3.0, circleSpacing*0.0, circleRadius, "F")
	pdf.TransformEnd()

	pdf.TransformBegin()
	pdf.SetFillColor(0, 0, 0)
	pdf.TransformTranslate(pageWidth-circleMargin, circleMargin)
	pdf.TransformRotate(270, 0, 0)
  pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*3.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*2.0, circleRadius, "F")
	pdf.SetFillColor(0, 255, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*1.0, circleRadius, "F")
	pdf.SetFillColor(255, 0, 0)
	pdf.Circle(circleSpacing*0.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*1.0, circleSpacing*0.0, circleRadius, "F")
	pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*2.0, circleSpacing*0.0, circleRadius, "F")
  pdf.SetFillColor(0, 0, 255)
	pdf.Circle(circleSpacing*3.0, circleSpacing*0.0, circleRadius, "F")
	pdf.TransformEnd()

  pdf.SetFont("Courier", "B", 10)
  pdf.SetXY(0, pageHeight - topMargin * 2 - bottomMargin - circleRadius)
  pdf.WriteAligned(0, 20, "show-dots (395)", "C")

	err := pdf.OutputFileAndClose("hello.pdf")
	if err != nil {
		fmt.Println(err)
	}
}
