package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	zmq "github.com/pebbe/zmq4"
)

const URL = "http://localhost:3000/"

const BASE_PATH = "/Users/jhaip/Code/lovelace/src/standalone_processes/"

// const BASE_PATH = "/home/jacob/lovelace/src/standalone_processes/"
const DOT_CODES_PATH = BASE_PATH + "files/dot-codes.txt"
const PDF_OUTPUT_FOLDER = BASE_PATH + "files/"
const LOG_PATH = BASE_PATH + "logs/1382__print-paper.log"
const MY_ID = 1382

type PrintWishResult struct {
	paperId       int
	shortFilename string
	sourceCode    string
}

type BatchMessage struct {
	Type string     `json:"type"`
	Fact [][]string `json:"fact"`
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
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
	if codeChar == '0' {
		pdf.SetFillColor(255, 0, 0)
	} else if codeChar == '1' {
		pdf.SetFillColor(0, 255, 0)
	} else if codeChar == '2' {
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
	circleRadius := 7.5
	circleSpacing := circleRadius * 2.0 * 1.4
	circleMargin := 10 + circleRadius

	pdf.SetFont("Courier", "", 7)
	useOutline := false
	pdf.ClipRect(circleMargin+leftMargin, circleMargin+topMargin, pageWidth-circleMargin*2-leftMargin-rightMargin, pageHeight-topMargin*2-bottomMargin-circleMargin*2, useOutline)
	pdf.TransformBegin()
	pdf.TransformTranslate(circleMargin, circleMargin)
	prevX, prevY := pdf.GetXY()
	NlinesOnPaper := 74
	lineNumbers := make([]string, NlinesOnPaper)
	for i := 0; i < NlinesOnPaper; i++ {
		lineNumbers[i] = strconv.Itoa(i)
	}
	lineNumbersString := strings.Join(lineNumbers, "\n")
	lineNumbersWidth := 6.0
	pdf.SetTextColor(150, 150, 150)
	pdf.MultiCell(lineNumbersWidth, 3, lineNumbersString, "", "R", false)
	pdf.SetXY(prevX, prevY)
	pdf.TransformTranslate(lineNumbersWidth, 0)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(pageWidth-circleMargin*2-leftMargin-rightMargin-lineNumbersWidth, 3, strings.Replace(sourceCode, string(9787), "\"", -1), "", "L", false)
	pdf.TransformEnd()
	pdf.ClipEnd()

	for i := 0; i < 4; i++ {
		pdf.TransformBegin()
		pdf.SetFillColor(0, 0, 0)
		if i == 0 {
			pdf.TransformTranslate(circleMargin+0, circleMargin+0)
		} else if i == 3 {
			pdf.TransformTranslate(circleMargin, pageHeight-circleMargin)
		} else if i == 2 {
			pdf.TransformTranslate(pageWidth-circleMargin, pageHeight-circleMargin)
		} else {
			pdf.TransformTranslate(pageWidth-circleMargin, circleMargin)
		}
		pdf.TransformRotate(-90.0*float64(i), 0, 0)
		code := code8400[i*(8400/4)+programId]
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
	pdf.SetXY(0, pageHeight-topMargin*2-bottomMargin-circleRadius-2)
	pdf.WriteAligned(0, 20, strconv.Itoa(programId), "C")
	pdf.SetXY(0, pageHeight-topMargin*2-bottomMargin-circleRadius-2+6)
	pdf.SetFont("Courier", "", 8)
	pdf.WriteAligned(0, 20, name, "C")

	err := pdf.OutputFileAndClose(PDF_OUTPUT_FOLDER + strconv.Itoa(programId) + ".pdf")
	if err != nil {
		log.Println(err)
	}
}

// https://stackoverflow.com/questions/48798588/how-do-you-remove-the-first-character-of-a-string
func trimLeftChars(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}

func get_wishes(subscriber *zmq.Socket, MY_ID_STR string, subscription_id string) []PrintWishResult {
	reply, err := subscriber.Recv(0)
	if err != nil {
		log.Println("get wishes error:")
		log.Println(err)
		panic(err)
	} else {
		log.Println("reply:")
		log.Println(reply)
	}
	msg_prefix := fmt.Sprintf("%s%s", MY_ID_STR, subscription_id)
	val := trimLeftChars(reply, len(msg_prefix)+13)
	json_val := make([]map[string][]string, 0)
	jsonValErr := json.Unmarshal([]byte(val), &json_val)
	if jsonValErr != nil {
		panic(jsonValErr)
	}
	printWishResults := make([]PrintWishResult, len(json_val))
	for i, json_result := range json_val {
		paperId, paperIdParseErr := strconv.Atoi(json_result["id"][1])
		checkErr(paperIdParseErr)
		printWishResults[i] = PrintWishResult{paperId, json_result["shortFilename"][1], json_result["sourceCode"][1]}
	}
	return printWishResults
}

func cleanupWishes(publisher *zmq.Socket, MY_ID_STR string) {
	batch_claims := make([]BatchMessage, 0)
	batch_claims = append(batch_claims, BatchMessage{"retract", [][]string{
		[]string{"variable", ""},
		[]string{"text", "wish"},
		[]string{"text", "paper"},
		[]string{"variable", ""},
		[]string{"text", "at"},
		[]string{"variable", ""},
		[]string{"text", "would"},
		[]string{"text", "be"},
		[]string{"text", "printed"},
	}})
	batch_claim_str, jsonMarshallErr := json.Marshal(batch_claims)
	checkErr(jsonMarshallErr)
	msg := fmt.Sprintf("....BATCH%s%s", MY_ID_STR, batch_claim_str)
	log.Println("Sending ", msg)
	s, err := publisher.Send(msg, 0)
	checkErr(err)
	log.Println("post send message!")
	log.Println(s)
}

func wishOutputFileWouldBePrinted(publisher *zmq.Socket, MY_ID_STR string, outputFilename string) {
	batch_claims := make([]BatchMessage, 0)
	batch_claims = append(batch_claims, BatchMessage{"claim", [][]string{
		[]string{"id", MY_ID_STR},
		[]string{"text", "wish"},
		[]string{"text", "file"},
		[]string{"text", outputFilename},
		[]string{"text", "would"},
		[]string{"text", "be"},
		[]string{"text", "printed"},
	}})
	batch_claim_str, jsonMarshallErr := json.Marshal(batch_claims)
	checkErr(jsonMarshallErr)
	msg := fmt.Sprintf("....BATCH%s%s", MY_ID_STR, batch_claim_str)
	log.Println("Sending ", msg)
	s, err := publisher.Send(msg, 0)
	checkErr(err)
	log.Println("post send message!")
	log.Println(s)
}

func initZeroMQ(MY_ID_STR string) (*zmq.Socket, *zmq.Socket) {
	log.Println("Connecting to hello world server...")
	publisher, newPubErr := zmq.NewSocket(zmq.PUB)
	checkErr(newPubErr)
	pubSubErr := publisher.Connect("tcp://localhost:5556")
	checkErr(pubSubErr)
	subscriber, newSubErr := zmq.NewSocket(zmq.SUB)
	checkErr(newSubErr)
	setSubErr := subscriber.SetSubscribe(MY_ID_STR)
	checkErr(setSubErr)
	connectErr := subscriber.Connect("tcp://localhost:5555")
	checkErr(connectErr)
	time.Sleep(1.0 * time.Second) // HACK: Wait for subscribers to connect
	return publisher, subscriber
}

func initWishSubscription(publisher *zmq.Socket, MY_ID_STR string) string {
	subscription_id := "bf272176-2df5-4664-b2a1-f9c5628e1d9f"
	sub_query := map[string]interface{}{
		"id": subscription_id,
		"facts": []string{
			"$ wish paper $id at $shortFilename would be printed",
			"$ $shortFilename has source code $sourceCode",
		},
	}
	sub_query_msg, jsonMarshallErr := json.Marshal(sub_query)
	checkErr(jsonMarshallErr)
	sub_msg := fmt.Sprintf("SUBSCRIBE%s%s", MY_ID_STR, sub_query_msg)
	_, sendErr := publisher.Send(sub_msg, 0)
	checkErr(sendErr)
	return subscription_id
}

func main() {
	/*** Set up logging ***/
	f, err := os.OpenFile(LOG_PATH, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	// log.Println("This is a test log entry")
	/*** /end logging setup ***/

	code8400, err := readLines(DOT_CODES_PATH)
	checkErr(err)

	MY_ID_STR := fmt.Sprintf("%04d", MY_ID)

	publisher, subscriber := initZeroMQ(MY_ID_STR)
	defer publisher.Close()
	defer subscriber.Close()
	subscription_id := initWishSubscription(publisher, MY_ID_STR)

	log.Println("done with init")

	for {
		printWishResults := get_wishes(subscriber, MY_ID_STR, subscription_id)
		cleanupWishes(publisher, MY_ID_STR)
		for _, result := range printWishResults {
			log.Printf("%#v\n", result)
			log.Println("PROGRAM ID:::")
			log.Printf("%#v\n", result.paperId)
			log.Printf("%#v\n", result.shortFilename)
			generatePrintFile(result.sourceCode, result.paperId, result.shortFilename, code8400)
			outputFilename := PDF_OUTPUT_FOLDER + strconv.Itoa(result.paperId) + ".pdf"
			wishOutputFileWouldBePrinted(publisher, MY_ID_STR, outputFilename)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
