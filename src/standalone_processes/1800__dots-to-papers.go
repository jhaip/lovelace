package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kokardy/listing"
	"github.com/mattn/go-ciede2000"
	zmq "github.com/pebbe/zmq4"
)

const BASE_PATH = "/Users/jhaip/Code/lovelace/src/standalone_processes/"

// const BASE_PATH = "/home/jacob/lovelace/src/standalone_processes/"
const LOG_PATH = BASE_PATH + "logs/1800__dots-to-papers.log"

type Vec struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Dot struct {
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Color     [3]int `json:"color"`
	Neighbors []int  `json:"-"`
}

type Corner struct {
	Corner        Dot      `json:"corner"`
	lines         [][]int  `json:"-"`
	sides         [][]int  `json:"-"`
	PaperId       int      `json:"paperId"`
	CornerId      int      `json:"cornerId"`
	ColorString   string   `json:"colorString"`
	RawColorsList [][3]int `json:"rawColorsList"`
}

type PaperCorner struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	CornerId int
}

type Paper struct {
	Id      string        `json:"id"`
	Corners []PaperCorner `json:"corners"`
}

type P struct {
	G     [][]int
	U     []int
	score float64
}

type BatchMessage struct {
	Type string     `json:"type"`
	Fact [][]string `json:"fact"`
}

const CAM_WIDTH = 1920
const CAM_HEIGHT = 1080
const dotSize = 12

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

	dotCodes8400 := get8400(BASE_PATH + "files/dot-codes.txt")
	if len(dotCodes8400) != 8400 {
		panic("DID NOT GET 8400 DOT CODES")
	}

	MY_ID := 1800
	MY_ID_STR := fmt.Sprintf("%04d", MY_ID)

	log.Println("Connecting to hello world server...")
	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	// publisher.SetSndhwm(1)
	publisher.Connect("tcp://localhost:5556")
	subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
	// subscriber.SetRcvhwm(1)  // subscription queue has only room for 1 message: latest dots
	// Is this ^ actually working?
	subscriber.SetSubscribe(MY_ID_STR)
	subscriber.Connect("tcp://localhost:5555")
	time.Sleep(1.0 * time.Second) // HACK: Wait for subscribers to connect
	count := 0

	dot_sub_id := "f47ac10b-58cc-0372-8567-0e02b2c3d479"
	dot_sub_query := map[string]interface{}{"id": dot_sub_id, "facts": []string{"$source dots $x $y color $r $g $b $t"}}
	dot_sub_query_msg, _ := json.Marshal(dot_sub_query)
	dot_sub_msg := fmt.Sprintf("SUBSCRIBE%s%s", MY_ID_STR, dot_sub_query_msg)
	publisher.Send(dot_sub_msg, 0)

	for {
		start := time.Now()

		points := getDots(subscriber, MY_ID_STR, dot_sub_id, start) // getPoints()

		timeGotDots := time.Since(start)
		// printDots(points)
		step1 := doStep1(points)
		fmt.Println("step1", len(step1))
		// printDots(step1)

		step2 := doStep2(step1)
		fmt.Println("step2", len(step2))
		// printCorners(step2[:5])
		// printJsonDots(step1)
		// claimCorners(step2)
		step3 := doStep3(step1, step2)
		fmt.Println("step3", len(step3))
		// printCorners(step3)
		step4 := doStep4CornersWithIds(step1, step3, dotCodes8400)
		fmt.Println("step4", len(step4))
		// claimCorners(publisher, step4)
		// printCorners(step4)
		papers := getPapersFromCorners(step4)
		// log.Println(papers)
		fmt.Println("papers", len(papers))

		timeProcessing := time.Since(start)
		claimPapers(publisher, MY_ID_STR, papers)

		count += 1
		// claimCounter(publisher, count)

		elapsed := time.Since(start)
		log.Printf("get dots  : %s \n", timeGotDots)
		log.Printf("processing: %s \n", timeProcessing)
		log.Printf("total     : %s \n", elapsed)

		// time.Sleep(10 * time.Millisecond)
	}
}

func projectMissingCorner(orderedCorners []PaperCorner, missingCornerId int) PaperCorner {
	cornerA := orderedCorners[(missingCornerId+1)%4]
	cornerB := orderedCorners[(missingCornerId+2)%4]
	cornerC := orderedCorners[(missingCornerId+3)%4]
	return PaperCorner{
		CornerId: missingCornerId,
		X:        cornerA.X + cornerC.X - cornerB.X,
		Y:        cornerA.Y + cornerC.Y - cornerB.Y,
	}
}

func getPapersFromCorners(corners []Corner) []Paper {
	papersMap := make(map[string][]PaperCorner)
	for _, corner := range corners {
		cornerIdStr := strconv.Itoa(corner.PaperId)
		_, idInMap := papersMap[cornerIdStr]
		cornerDotVec := PaperCorner{corner.Corner.X, corner.Corner.Y, corner.CornerId}
		if idInMap {
			papersMap[cornerIdStr] = append(papersMap[cornerIdStr], cornerDotVec)
		} else {
			papersMap[cornerIdStr] = []PaperCorner{cornerDotVec}
		}
	}
	// log.Println(papersMap)
	papers := make([]Paper, 0)
	for id := range papersMap {
		if len(papersMap[id]) < 3 {
			continue
		}
		const TOP_LEFT = 0
		const TOP_RIGHT = 1
		const BOTTOM_RIGHT = 2
		const BOTTOM_LEFT = 3
		orderedCorners := make([]PaperCorner, 4) // [tl, tr, br, bl]
		for _, corner := range papersMap[id] {
			orderedCorners[corner.CornerId] = corner
		}
		if len(papersMap[id]) == 3 {
			// Identify the missing one then use the other three points to guess
			// where the missing corner would be.
			NIL_CORNER := PaperCorner{}
			for i := 0; i < 4; i++ {
				if orderedCorners[i] == NIL_CORNER {
					orderedCorners[i] = projectMissingCorner(orderedCorners, i)
				}
			}
			log.Println("FILLED IN A MISSING CORNER", id)
		}
		papers = append(papers, Paper{id, orderedCorners})
	}
	return papers
}

func printDots(data []Dot) {
	s := make([]string, len(data))
	for i, d := range data {
		s[i] = fmt.Sprintf("%#v", d)
	}
	log.Println(strings.Join(s, "\n"))
}

func printCorners(data []Corner) {
	s := make([]string, len(data))
	for i, d := range data {
		s[i] = fmt.Sprintf("%v \t %v \t %v \t %v \t %v", d.Corner, d.lines, d.sides, d.PaperId, d.CornerId)
	}
	fmt.Println(strings.Join(s, "\n"))

	cornersAlmostStr, err := json.Marshal(data)
	fmt.Println("Err?")
	fmt.Println(err)
	cornersStr := string(cornersAlmostStr)
	fmt.Println(cornersStr)
}

func distanceSquared(p1 Dot, p2 Dot) float64 {
	return math.Pow(float64(p1.X-p2.X), 2) +
		math.Pow(float64(p1.Y-p2.Y), 2)
}

func getWithin(points []Dot, ref Dot, i int, dist int) []int {
	within := make([]int, 0)
	for pi, point := range points {
		if pi != i && distanceSquared(point, ref) < float64(math.Pow(float64(2*dist), 2)) {
			within = append(within, pi)
		}
	}
	return within
}

func doStep1(points []Dot) []Dot {
	// Connect Neighbors
	step1 := make([]Dot, len(points))
	for i, nodeData := range points {
		step1[i] = nodeData
		step1[i].Neighbors = getWithin(points, nodeData, i, dotSize)
	}
	return step1
}

func cosineSimilarity(vectA Vec, vectB Vec) float64 {
	var dotProduct float64 = float64(vectA.X*vectB.X + vectA.Y*vectB.Y)
	var normA float64 = float64(vectA.X*vectA.X + vectA.Y*vectA.Y)
	var normB float64 = float64(vectB.X*vectB.X + vectB.Y*vectB.Y)
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func sub(a Dot, b Dot) Vec {
	return Vec{a.X - b.X, a.Y - b.Y}
}

func crossProduct(a Vec, b Vec) float64 {
	return float64(a.X*b.Y - a.Y*b.X)
}

func getNeighborsInDirection(nodeData []Dot, node Dot, ref Dot) []int {
	direction := sub(node, ref)
	results := make([]int, 0)
	for _, neighbor := range node.Neighbors {
		neighborNode := nodeData[neighbor]
		if cosineSimilarity(sub(neighborNode, node), direction) > 0.95 {
			results = append(results, neighbor)
		}
	}
	return results
}

func searchInner(nodes []Dot, start Dot, depth int, results [][]int) [][]int {
	if depth == 0 {
		return results
	}
	newResults := make([][]int, 0)
	for _, path := range results {
		neighbors := getNeighborsInDirection(nodes, nodes[path[len(path)-1]], start)
		for _, neighbor := range neighbors {
			newResults = append(newResults, append(path, neighbor))
		}
	}
	return searchInner(nodes, start, depth-1, newResults)
}

func search(nodes []Dot, start Dot, depth int) [][]int {
	results := make([][]int, len(start.Neighbors))
	for i, neighbor := range start.Neighbors {
		results[i] = []int{neighbor}
	}
	return searchInner(nodes, start, depth-1, results)
}

func doStep2(points []Dot) []Corner {
	step2 := make([]Corner, len(points))
	for i, point := range points {
		step2[i] = Corner{Corner: point, lines: search(points, point, 3)}
	}
	return step2
}

func doStep3(nodes []Dot, corners []Corner) []Corner {
	results := make([]Corner, 0)
	for cornerIndex, corner := range corners {
		if len(corner.lines) < 2 {
			continue
		}
		for i := 0; i < len(corner.lines); i += 1 {
			for j := i + 1; j < len(corner.lines); j += 1 {
				side1 := corner.lines[i]
				side2 := corner.lines[j]
				line1 := sub(nodes[side1[len(side1)-1]], nodes[cornerIndex])
				line2 := sub(nodes[side2[len(side2)-1]], nodes[cornerIndex])
				similarity := cosineSimilarity(line1, line2)
				if math.Abs(similarity) < 0.5 {
					newCorner := corner
					if crossProduct(line1, line2) > 0 {
						newCorner.sides = [][]int{side2, side1}
						results = append(results)
					} else {
						newCorner.sides = [][]int{side1, side2}
					}
					results = append(results, newCorner)
				}
			}
		}
	}
	return results
}

func indexOf(word string, data []string) int {
	for k, v := range data {
		if word == v {
			return k
		}
	}
	return -1
}

func getColorDistance(a, b [3]int) float64 {
	return math.Abs(float64(a[0]-b[0])) + math.Abs(float64(a[1]-b[1])) + math.Abs(float64(a[2]-b[2]))
	// using CIEDE2000 color diff is 5x slower than RGB diff (almost 2ms for one corner)
	// c1 := &color.RGBA{
	//   uint8(a[0]),
	//   uint8(a[1]),
	//   uint8(a[2]),
	//   255,
	// }
	// c2 := &color.RGBA{
	//   uint8(b[0]),
	//   uint8(b[1]),
	//   uint8(b[2]),
	//   255,
	// }
	// return ciede2000.Diff(c1, c2)
}

func identifyColorGroups(colors [][3]int, group P) string {
	color_templates := listing.Permutations(
		listing.IntReplacer([]int{0, 1, 2, 3}), 4, false, 4,
	)
	// log.Println(color_templates)

	calibration := [][3]int{[3]int{255, 0, 0}, [3]int{0, 255, 0}, [3]int{0, 0, 255}, [3]int{0, 0, 0}}
	// calibration := make([][3]int, 4)
	// calibration[0] = [3]int{190, 55, 49}  // red
	// calibration[1] = [3]int{168, 164, 145}  // green
	// calibration[2] = [3]int{148, 151, 190}  // blue
	// calibration[3] = [3]int{113, 72, 96}  // dark

	minScore := -1.0
	var bestMatch []int // index = color, value = index of group in P that matches color
	for rr := range color_templates {
		r := rr.(listing.IntReplacer)
		score := 0.0
		score += getColorDistance(calibration[0], colors[group.G[r[0]][0]])
		score += getColorDistance(calibration[1], colors[group.G[r[1]][0]])
		score += getColorDistance(calibration[2], colors[group.G[r[2]][0]])
		score += getColorDistance(calibration[3], colors[group.G[r[3]][0]])
		// log.Println(r, score)
		if minScore == -1 || score < minScore {
			minScore = score
			bestMatch = r
		}
	}

	// log.Println("best match", bestMatch)

	result := make([]string, 7)
	for i, g := range bestMatch {
		for _, k := range group.G[g] {
			result[k] = strconv.Itoa(i)
		}
	}
	// log.Println("Result", result)  // Something like "1222203"

	return strings.Join(result, "")
}

func getGetPaperIdFromColors2(colors [][3]int, dotCodes8400 []string) (int, int, string) {
	// color_combinations := combinations_as_list(7, 4)
	color_combinations := listing.Combinations(
		listing.IntReplacer([]int{0, 1, 2, 3, 4, 5, 6}), 4, false, 7,
	)
	// log.Println(color_combinations)

	minScore := -1.0
	var bestGroup P
	for rr := range color_combinations {
		r := rr.(listing.IntReplacer)
		p := P{G: [][]int{[]int{r[0]}, []int{r[1]}, []int{r[2]}, []int{r[3]}}}
		// Fill p.U with unused #s
		for i := 0; i < 7; i += 1 {
			if r[0] != i && r[1] != i && r[2] != i && r[3] != i {
				p.U = append(p.U, i)
			}
		}
		// pop of each element in p.U and add to closet colored group
		for i := 0; i < 3; i += 1 {
			u_color := colors[p.U[0]]
			// add element to group closest in color
			min_i := 0
			min := getColorDistance(u_color, colors[r[0]])
			for j := 1; j < 4; j += 1 {
				d := getColorDistance(u_color, colors[r[j]])
				if d < min {
					min = d
					min_i = j
				}
			}
			p.G[min_i] = append(p.G[min_i], p.U[0])
			p.U = p.U[1:]
			p.score += min
		}

		// log.Println(p)

		// Keep track of the grouping with the lowest score
		if minScore == -1 || p.score < minScore {
			minScore = p.score
			bestGroup = p
		}
	}

	// log.Println("Best group", bestGroup)
	colorString := identifyColorGroups(colors, bestGroup)

	log.Printf("%v \n", colorString)
	colors8400Index := indexOf(colorString, dotCodes8400)
	if colors8400Index > 0 {
		paperId := colors8400Index % (8400 / 4)
		cornerId := colors8400Index / (8400 / 4)
		return paperId, cornerId, colorString
	}
	return -1, -1, colorString
}

func getGetPaperIdFromColors(colors [][3]int, dotCodes8400 []string) (int, int, string) {
	var colorString string

	calibrationColors := make([][3]int, 4)
	calibrationColors[0] = [3]int{170, 48, 31}   // red
	calibrationColors[1] = [3]int{138, 131, 94}  // green
	calibrationColors[2] = [3]int{112, 118, 150} // blue
	calibrationColors[3] = [3]int{52, 23, 21}    // dark

	// calibrationColors[0] = [3]int{202, 61, 79}  // red
	// calibrationColors[1] = [3]int{162, 156, 118}  // green
	// calibrationColors[2] = [3]int{126, 148, 191}  // blue
	// calibrationColors[3] = [3]int{85, 58, 94}  // dark

	// calibrationColors[0] = [3]int{204, 98, 107}  // red
	// calibrationColors[1] = [3]int{200, 186, 167}  // green
	// calibrationColors[2] = [3]int{176, 170, 198}  // blue
	// calibrationColors[3] = [3]int{125, 91, 107}  // dark

	for _, colorData := range colors {
		minIndex := 0
		minValue := 99999.0
		for i, calibrationColorData := range calibrationColors {
			c1 := &color.RGBA{
				uint8(colorData[0]),
				uint8(colorData[1]),
				uint8(colorData[2]),
				255,
			}
			c2 := &color.RGBA{
				uint8(calibrationColorData[0]),
				uint8(calibrationColorData[1]),
				uint8(calibrationColorData[2]),
				255,
			}
			value := ciede2000.Diff(c1, c2)
			if i == 0 || value < minValue {
				minIndex = i
				minValue = value
			}
		}
		colorString += strconv.Itoa(minIndex)
	}
	log.Printf("%v \n", colorString)
	colors8400Index := indexOf(colorString, dotCodes8400)
	if colors8400Index > 0 {
		paperId := colors8400Index % (8400 / 4)
		cornerId := colors8400Index / (8400 / 4)
		return paperId, cornerId, colorString
	}
	return -1, -1, colorString
}

func lineToColors(nodes []Dot, line []int, shouldReverse bool) [][3]int {
	results := make([][3]int, len(line))
	for i, nodeIndex := range line {
		if shouldReverse {
			results[len(line)-1-i] = nodes[nodeIndex].Color
		} else {
			results[i] = nodes[nodeIndex].Color
		}
	}
	return results
}

func doStep4CornersWithIds(nodes []Dot, corners []Corner, dotCodes8400 []string) []Corner {
	results := make([]Corner, 0)
	for _, corner := range corners {
		newCorner := corner
		rawColorsList := append(append(lineToColors(nodes, corner.sides[0], true), corner.Corner.Color), lineToColors(nodes, corner.sides[1], false)...)
		paperId, cornerId, colorString := getGetPaperIdFromColors2(rawColorsList, dotCodes8400)
		newCorner.PaperId = paperId
		newCorner.CornerId = cornerId
		newCorner.ColorString = colorString
		newCorner.RawColorsList = rawColorsList
		results = append(results, newCorner)
	}
	return results
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

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func getDots(subscriber *zmq.Socket, MY_ID_STR string, dot_sub_id string, start time.Time) []Dot {
	var reply string
	nLoops := 0
	dot_prefix := fmt.Sprintf("%s%s", MY_ID_STR, dot_sub_id)
	for reply == "" || reply[len(reply)-4:] == "[{}]" {
		for {
			nLoops += 1
			tmp_reply, err := subscriber.Recv(zmq.DONTWAIT)
			if err != nil {
				break
			} else {
				reply = tmp_reply
				// fmt.Println("GOT REPLY:")
				// fmt.Println(reply)
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
	timeGotDotsPre := time.Since(start)
	log.Printf("get dots pre  : %s , %s\n", timeGotDotsPre, nLoops)
	// log.Println("Received ", reply)
	timeVal, err := strconv.ParseInt(reply[len(dot_prefix):len(dot_prefix)+13], 10, 64)
	if err != nil {
		panic(err)
	}
	// fmt.Println("time val")
	// fmt.Println(timeVal)
	timeDiff := makeTimestamp() - timeVal
	fmt.Printf("time diff: %v ms\n", timeDiff)
	val := trimLeftChars(reply, len(dot_prefix)+13)
	// fmt.Println("GOT RESULT:")
	// fmt.Println(val)
	json_val := make([]map[string][]string, 0)
	/*
		  type Dot struct {
			X         int    `json:"x"`
			Y         int    `json:"y"`
			Color     [3]int `json:"color"`
			Neighbors []int  `json:"-"`
		}
	*/
	// TODO: parse val
	json.Unmarshal([]byte(val), &json_val)
	fmt.Println("GET JSON RESULT:")
	// fmt.Println(json_val)
	// fmt.Println(json_val[0])
	claimTime, _ := strconv.ParseFloat(json_val[0]["t"][1], 64)
	claimTimeDiff := makeTimestamp() - int64(claimTime)
	fmt.Printf("claim time diff: %v ms\n", claimTimeDiff)
	res := make([]Dot, 0)
	for _, json_result := range json_val {
		x, _ := strconv.Atoi(json_result["x"][1])
		y, _ := strconv.Atoi(json_result["y"][1])
		r, _ := strconv.Atoi(json_result["r"][1])
		g, _ := strconv.Atoi(json_result["g"][1])
		b, _ := strconv.Atoi(json_result["b"][1])
		res = append(res, Dot{x, y, [3]int{r, g, b}, make([]int, 0)})
	}
	return res
}

func claimPapers(publisher *zmq.Socket, MY_ID_STR string, papers []Paper) {
	fmt.Println("CLAIM PAPERS -----")
	fmt.Println(papers)
	// papersAlmostStr, _ := json.Marshal(papers)
	// papersStr := string(papersAlmostStr)
	// log.Println(papersStr)
	/*
		  type PaperCorner struct {
			X        int `json:"x"`
			Y        int `json:"y"`
			CornerId int
		}

		type Paper struct {
			Id      string        `json:"id"`
			Corners []PaperCorner `json:"corners"`
		}
	*/

	// for _, paper := range papers {
	// 	papersStr := fmt.Sprintf("camera 1 sees paper %s at TL %v %v TR %v %v BR %v %v BL %v %v at %v", paper.Id, paper.Corners[0].X, paper.Corners[0].Y, paper.Corners[1].X, paper.Corners[1].Y, paper.Corners[2].X, paper.Corners[2].Y, paper.Corners[3].X, paper.Corners[3].Y, 99)
	// 	msg := fmt.Sprintf("....CLAIM%s%s", MY_ID_STR, papersStr)
	// 	log.Println("Sending ", msg)
	// 	publisher.Send(msg, 0)
	// }

	batch_claims := make([]BatchMessage, 0)
	batch_claims = append(batch_claims, BatchMessage{"retract", [][]string{
		[]string{"id", MY_ID_STR},
		[]string{"postfix", ""},
	}})
	// $ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time
	for _, paper := range papers {
		batch_claims = append(batch_claims, BatchMessage{"claim", [][]string{
			[]string{"id", MY_ID_STR},
			[]string{"text", "camera"},
			[]string{"integer", "1"},
			[]string{"text", "sees"},
			[]string{"text", "paper"},
			[]string{"integer", paper.Id},
			[]string{"text", "at"},
			[]string{"text", "TL"},
			[]string{"text", "("},
			[]string{"integer", strconv.Itoa(paper.Corners[0].X)},
			[]string{"text", ","},
			[]string{"integer", strconv.Itoa(paper.Corners[0].Y)},
			[]string{"text", ")"},
			[]string{"text", "TR"},
			[]string{"text", "("},
			[]string{"integer", strconv.Itoa(paper.Corners[1].X)},
			[]string{"text", ","},
			[]string{"integer", strconv.Itoa(paper.Corners[1].Y)},
			[]string{"text", ")"},
			[]string{"text", "BR"},
			[]string{"text", "("},
			[]string{"integer", strconv.Itoa(paper.Corners[2].X)},
			[]string{"text", ","},
			[]string{"integer", strconv.Itoa(paper.Corners[2].Y)},
			[]string{"text", ")"},
			[]string{"text", "BL"},
			[]string{"text", "("},
			[]string{"integer", strconv.Itoa(paper.Corners[3].X)},
			[]string{"text", ","},
			[]string{"integer", strconv.Itoa(paper.Corners[3].Y)},
			[]string{"text", ")"},
			[]string{"text", "@"},
			[]string{"integer", "99"},
		}})
	}
	batch_claim_str, _ := json.Marshal(batch_claims)
	msg := fmt.Sprintf("....BATCH%s%s", MY_ID_STR, batch_claim_str)
	log.Println("Sending ", msg)
	fmt.Println("Sending ", msg)
	publisher.Send(msg, 0)
}

func printJsonDots(dots []Dot) {
	cornersAlmostStr, err := json.Marshal(dots)
	log.Println("Err?")
	log.Println(err)
	cornersStr := string(cornersAlmostStr)
	log.Println(cornersStr)
	log.Println("---")
}

func claimCorners(publisher *zmq.Socket, corners []Corner) {
	cornersAlmostStr, err := json.Marshal(corners)
	log.Println("Err?")
	log.Println(err)
	cornersStr := string(cornersAlmostStr)
	log.Println(cornersStr)
	msg := fmt.Sprintf("CLAIM[global/corners]%s", cornersStr)
	log.Println("Sending ", msg)
	publisher.Send(msg, 0)
}

func claimCounter(publisher *zmq.Socket, count int) {
	msg := fmt.Sprintf("CLAIM[global/dtpcount]%v", count)
	log.Println("Sending ", msg)
	publisher.Send(msg, 0)
}

func get8400(fileName string) []string {
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	// Create new Scanner.
	scanner := bufio.NewScanner(f)
	result := []string{}
	// Use Scan.
	for scanner.Scan() {
		line := scanner.Text()
		// Append line to result.
		result = append(result, line)
	}
	return result
}
