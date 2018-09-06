package main

import (
  zmq "github.com/pebbe/zmq4"
  "os"
  "log"
  "time"
  "encoding/json"
  "fmt"
  "math"
  "strings"
  "strconv"
  "image/color"
  "github.com/mattn/go-ciede2000"
  "github.com/kokardy/listing"
)

// const BASE_PATH = "/Users/jhaip/Code/lovelace/src/standalone_processes/"
const BASE_PATH = "/home/jacob/lovelace/src/standalone_processes/"
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
	Corner Dot         `json:"corner"`
  lines  [][]int     `json:"-"`
	sides  [][]int     `json:"-"`
	PaperId     int    `json:"paperId"`
  CornerId     int   `json:"cornerId"`
  ColorString string `json:"colorString"`
  RawColorsList [][3]int `json:"rawColorsList"`
}

type PaperCorner struct {
  X int `json:"x"`
	Y int `json:"y"`
  CornerId     int
}

type Paper struct {
  Id        string        `json:"id"`
  Corners   []PaperCorner `json:"corners"`
}

type P struct {
  G [][]int
  U []int
  score float64
}

const CAM_WIDTH = 1920
const CAM_HEIGHT = 1080
const dotSize = 12

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

  log.Println("Connecting to hello world server...")
  publisher, _ := zmq.NewSocket(zmq.PUB)
  defer publisher.Close()
  // publisher.SetSndhwm(1)
  publisher.Connect("tcp://localhost:5555")
  subscriber, _ := zmq.NewSocket(zmq.SUB)
  defer subscriber.Close()
  // subscriber.SetRcvhwm(1)  // subscription queue has only room for 1 message: latest dots
  // Is this ^ actually working?
  filter := "CLAIM[global/dots]"
  subscriber.SetSubscribe(filter)
  subscriber.Connect("tcp://localhost:5556")
  count := 0

  for {
    start := time.Now()

  	points := getDots(subscriber, start)  // getPoints()

    timeGotDots := time.Since(start)
  	// printDots(points)
  	step1 := doStep1(points)
  	// printDots(step1)

  	step2 := doStep2(step1)
  	// printCorners(step2[:5])
    // printJsonDots(step1)
    // claimCorners(step2)
  	step3 := doStep3(step1, step2)
  	// printCorners(step3)
  	step4 := doStep4CornersWithIds(step1, step3)
    // claimCorners(publisher, step4)
  	// printCorners(step4)
    papers := getPapersFromCorners(step4)
    log.Println(papers)

    timeProcessing := time.Since(start)
    claimPapers(publisher, papers)

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
  cornerA := orderedCorners[(missingCornerId + 1) % 4]
  cornerB := orderedCorners[(missingCornerId + 2) % 4]
  cornerC := orderedCorners[(missingCornerId + 3) % 4]
  return PaperCorner{
    CornerId: missingCornerId,
    X: cornerA.X + cornerC.X - cornerB.X,
    Y: cornerA.Y + cornerC.Y - cornerB.Y,
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
    orderedCorners := make([]PaperCorner, 4)  // [tl, tr, br, bl]
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
	log.Println(strings.Join(s, "\n"))

  cornersAlmostStr, err := json.Marshal(data)
  log.Println("Err?")
  log.Println(err)
  cornersStr := string(cornersAlmostStr)
  log.Println(cornersStr)
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
    listing.IntReplacer([]int{0,1,2,3}), 4, false, 4,
  )
  // log.Println(color_templates)

  calibration := [][3]int{[3]int{255, 0, 0}, [3]int{0, 255, 0}, [3]int{0, 0, 255}, [3]int{0, 0, 0}}
  // calibration := make([][3]int, 4)
  // calibration[0] = [3]int{190, 55, 49}  // red
  // calibration[1] = [3]int{168, 164, 145}  // green
  // calibration[2] = [3]int{148, 151, 190}  // blue
  // calibration[3] = [3]int{113, 72, 96}  // dark

  minScore := -1.0
  var bestMatch []int  // index = color, value = index of group in P that matches color
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

func getGetPaperIdFromColors2(colors [][3]int) (int, int, string) {
  // color_combinations := combinations_as_list(7, 4)
  color_combinations := listing.Combinations(
    listing.IntReplacer([]int{0,1,2,3,4,5,6}), 4, false, 7,
  )
  // log.Println(color_combinations)

  minScore := -1.0
  var bestGroup P
  for rr := range color_combinations {
    r := rr.(listing.IntReplacer)
    p := P{G: [][]int{[]int{r[0]}, []int{r[1]}, []int{r[2]}, []int{r[3]} }}
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
  colors8400Index := indexOf(colorString, get8400())
  if colors8400Index > 0 {
    paperId := colors8400Index % (8400 / 4)
    cornerId := colors8400Index / (8400 / 4)
    return paperId, cornerId, colorString
  }
	return -1, -1, colorString
}

func getGetPaperIdFromColors(colors [][3]int) (int, int, string) {
	var colorString string

  calibrationColors := make([][3]int, 4)
  calibrationColors[0] = [3]int{170, 48, 31}  // red
  calibrationColors[1] = [3]int{138, 131, 94}  // green
  calibrationColors[2] = [3]int{112, 118, 150}  // blue
  calibrationColors[3] = [3]int{52, 23, 21}  // dark

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
  colors8400Index := indexOf(colorString, get8400())
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

func doStep4CornersWithIds(nodes []Dot, corners []Corner) []Corner {
	results := make([]Corner, 0)
	for _, corner := range corners {
		newCorner := corner
    rawColorsList := append(append(lineToColors(nodes, corner.sides[0], true), corner.Corner.Color), lineToColors(nodes, corner.sides[1], false)...)
    paperId, cornerId, colorString := getGetPaperIdFromColors2(rawColorsList)
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

func getDots(subscriber *zmq.Socket, start time.Time) []Dot {
  var reply string
  nLoops := 0
  for reply == "" {
    for {
      nLoops += 1
      tmp_reply, err := subscriber.Recv(zmq.DONTWAIT)
      if err != nil {
        break
      } else {
        reply = tmp_reply
      }
    }
    time.Sleep(1 * time.Millisecond)
  }
  timeGotDotsPre := time.Since(start)
  log.Printf("get dots pre  : %s , %s\n", timeGotDotsPre, nLoops)
	// log.Println("Received ", reply)
  // "CLAIM[global/dots]" = 18 characters to trim off from beginning of JSON
  val := trimLeftChars(reply, 18)
  res := make([]Dot, 0)
	json.Unmarshal([]byte(val), &res)
  return res
}

func claimPapers(publisher *zmq.Socket, papers []Paper) {
  papersAlmostStr, _ := json.Marshal(papers)
  papersStr := string(papersAlmostStr)
  log.Println(papersStr)
  msg := fmt.Sprintf("CLAIM[global/papers]%s", papersStr)
  log.Println("Sending ", msg)
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

func getPoints() []Dot {
	str := `[{"x":481,"y":224,"color":[146,104,58]},{"x":732,"y":270,"color":[176,170,198]},{"x":464,"y":274,"color":[140,84,50]},{"x":726,"y":276,"color":[173,173,198]},{"x":738,"y":276,"color":[179,171,194]},{"x":744,"y":281,"color":[125,91,107]},{"x":720,"y":282,"color":[204,98,107]},{"x":751,"y":287,"color":[200,186,167]},{"x":714,"y":288,"color":[196,189,172]},{"x":696,"y":306,"color":[172,167,200]},{"x":690,"y":312,"color":[191,180,160]},{"x":684,"y":318,"color":[190,179,158]},{"x":791,"y":322,"color":[199,105,110]},{"x":522,"y":323,"color":[129,81,54]},{"x":498,"y":324,"color":[131,85,61]},{"x":678,"y":324,"color":[184,175,164]},{"x":637,"y":328,"color":[137,97,68]},{"x":798,"y":328,"color":[201,107,117]},{"x":595,"y":330,"color":[137,88,79]},{"x":684,"y":330,"color":[201,97,107]},{"x":804,"y":333,"color":[129,94,95]},{"x":691,"y":336,"color":[192,185,169]},{"x":811,"y":339,"color":[185,175,157]},{"x":698,"y":341,"color":[114,86,97]},{"x":806,"y":345,"color":[170,166,182]},{"x":800,"y":351,"color":[192,177,154]},{"x":794,"y":357,"color":[189,174,160]},{"x":630,"y":360,"color":[150,95,69]},{"x":604,"y":361,"color":[131,85,49]},{"x":518,"y":371,"color":[127,80,57]},{"x":776,"y":377,"color":[187,179,148]},{"x":738,"y":378,"color":[205,91,106]},{"x":745,"y":383,"color":[190,177,158]},{"x":770,"y":383,"color":[110,79,91]},{"x":752,"y":389,"color":[200,99,115]},{"x":764,"y":389,"color":[103,70,92]},{"x":627,"y":390,"color":[159,159,197]},{"x":636,"y":390,"color":[165,163,198]},{"x":645,"y":390,"color":[202,91,103]},{"x":653,"y":390,"color":[106,75,82]},{"x":682,"y":390,"color":[195,185,163]},{"x":691,"y":390,"color":[166,160,187]},{"x":700,"y":390,"color":[102,75,76]},{"x":709,"y":390,"color":[165,158,190]},{"x":582,"y":394,"color":[133,86,72]},{"x":758,"y":395,"color":[174,161,189]},{"x":627,"y":399,"color":[164,164,196]},{"x":709,"y":399,"color":[164,166,190]},{"x":709,"y":407,"color":[170,163,184]},{"x":627,"y":408,"color":[159,164,200]},{"x":594,"y":409,"color":[128,86,65]},{"x":627,"y":416,"color":[185,175,156]},{"x":710,"y":416,"color":[196,71,90]},{"x":719,"y":442,"color":[124,78,68]},{"x":596,"y":454,"color":[139,93,69]},{"x":881,"y":460,"color":[186,176,152]},{"x":872,"y":461,"color":[187,175,142]},{"x":854,"y":462,"color":[116,82,97]},{"x":863,"y":462,"color":[111,82,91]},{"x":720,"y":465,"color":[153,97,77]},{"x":799,"y":468,"color":[158,159,191]},{"x":790,"y":469,"color":[155,158,188]},{"x":883,"y":469,"color":[202,90,101]},{"x":781,"y":470,"color":[185,174,146]},{"x":711,"y":471,"color":[187,78,76]},{"x":772,"y":471,"color":[92,53,52]},{"x":627,"y":472,"color":[180,168,145]},{"x":1255,"y":472,"color":[159,127,114]},{"x":884,"y":478,"color":[189,90,84]},{"x":773,"y":480,"color":[185,61,72]},{"x":618,"y":481,"color":[137,94,78]},{"x":627,"y":481,"color":[151,147,177]},{"x":712,"y":481,"color":[98,62,65]},{"x":601,"y":482,"color":[146,98,62]},{"x":886,"y":487,"color":[157,154,178]},{"x":775,"y":489,"color":[82,48,78]},{"x":627,"y":490,"color":[68,43,57]},{"x":712,"y":490,"color":[195,74,85]},{"x":592,"y":494,"color":[133,78,49]},{"x":776,"y":498,"color":[184,65,88]},{"x":628,"y":499,"color":[140,140,180]},{"x":637,"y":499,"color":[141,145,174]},{"x":646,"y":499,"color":[73,37,38]},{"x":655,"y":499,"color":[189,56,65]},{"x":685,"y":499,"color":[180,178,151]},{"x":694,"y":499,"color":[154,158,187]},{"x":703,"y":499,"color":[184,55,71]},{"x":712,"y":499,"color":[86,58,71]},{"x":619,"y":504,"color":[108,59,45]},{"x":586,"y":506,"color":[108,63,34]},{"x":509,"y":510,"color":[96,46,36]},{"x":638,"y":514,"color":[107,55,30]},{"x":892,"y":516,"color":[98,77,82]},{"x":570,"y":518,"color":[102,56,46]},{"x":651,"y":519,"color":[107,56,30]},{"x":709,"y":522,"color":[149,105,64]},{"x":893,"y":525,"color":[161,159,186]},{"x":781,"y":528,"color":[176,178,147]},{"x":1277,"y":531,"color":[73,51,49]},{"x":895,"y":534,"color":[186,90,91]},{"x":782,"y":537,"color":[184,55,73]},{"x":621,"y":539,"color":[101,54,37]},{"x":603,"y":540,"color":[99,52,30]},{"x":897,"y":544,"color":[99,66,78]},{"x":888,"y":545,"color":[176,172,138]},{"x":879,"y":546,"color":[174,171,143]},{"x":784,"y":547,"color":[79,62,71]},{"x":870,"y":547,"color":[191,67,71]},{"x":640,"y":550,"color":[103,59,31]},{"x":813,"y":553,"color":[86,54,74]},{"x":804,"y":554,"color":[157,158,182]},{"x":794,"y":555,"color":[184,176,151]},{"x":785,"y":556,"color":[147,150,187]},{"x":682,"y":559,"color":[138,99,69]},{"x":234,"y":561,"color":[56,45,85]},{"x":228,"y":563,"color":[107,100,140]},{"x":254,"y":563,"color":[138,123,165]},{"x":728,"y":569,"color":[136,88,62]},{"x":793,"y":571,"color":[140,90,48]},{"x":292,"y":589,"color":[77,70,83]},{"x":660,"y":621,"color":[104,59,27]},{"x":74,"y":695,"color":[153,70,47]}]`

	res := make([]Dot, 0)
	json.Unmarshal([]byte(str), &res)

	return res
}

func get8400() []string {
	return []string{"1013211", "0122103", "1021123", "2031100", "2110030", "3122003", "0213303", "2113031", "3303120", "0032331", "2203013", "3223201", "1032313", "2313000", "3210001", "2321310", "2010031", "0220123", "1003323", "3012003", "3132032", "2100301", "1201333", "1303231", "0323133", "0310020", "1103232", "2230122", "2100331", "2012023", "1202223", "0232132", "2233110", "1303312", "3220123", "1032032", "1222330", "3313203", "0313232", "1320232", "2031301", "2311022", "3232120", "3121000", "0320213", "1123020", "2131210", "0111132", "2010033", "2030331", "3211032", "2132013", "1113023", "0313122", "0231111", "0310312", "1202332", "2010312", "3032120", "2001233", "3120003", "0213123", "0122113", "1312210", "1013320", "3021012", "3131020", "1133201", "1200113", "3102330", "2322210", "0311213", "0102131", "0012230", "3013203", "2102231", "3202211", "3012010", "1232110", "0312231", "0201233", "1303032", "0331202", "1330002", "0331002", "2103020", "2331301", "0120103", "0111230", "1223202", "0032131", "0300120", "1001132", "3300312", "1030212", "0300312", "2210332", "1312002", "3312120", "1110322", "0231002", "1020323", "3210302", "0120321", "1003233", "1313200", "0132001", "0313201", "1030210", "2011033", "3003212", "3021001", "0230231", "1103230", "2230101", "3132003", "1312012", "1322021", "3012132", "0013122", "0221230", "1130320", "2310011", "2300212", "1032021", "3002231", "2201113", "2133301", "3120111", "0311132", "1012023", "1133220", "1310022", "2101203", "0203313", "2313010", "0211313", "0203031", "1212003", "1133021", "3010002", "3112202", "2121130", "0122322", "0310212", "0020321", "3301023", "2023100", "1012203", "2331201", "0213222", "1322101", "0301201", "2021331", "2111030", "1230100", "2320201", "2323301", "3201310", "3103200", "3223001", "1312202", "0231031", "2122300", "2110322", "2233021", "3033212", "1223000", "0203212", "0233013", "2013332", "1310223", "1230313", "0021231", "2311230", "3310102", "2031311", "3201231", "1320303", "1203302", "2000123", "2303011", "1220130", "1021033", "3032011", "3010212", "0223021", "0013220", "2033111", "2223210", "3110212", "0031121", "0031302", "0023321", "0130231", "0221131", "1320332", "0231310", "2230311", "0123122", "2300211", "2303312", "0331122", "2103311", "2310200", "1033233", "2230210", "2102311", "0320111", "3220013", "2301233", "1320001", "0331223", "2312100", "1002031", "2013031", "0213300", "2012313", "2233301", "2213000", "0121113", "0001132", "2021033", "0231200", "0121332", "3230122", "3113203", "0311123", "0123002", "2200031", "3310123", "3311023", "0321030", "3200312", "2331020", "1020130", "1012223", "1233300", "0113321", "0323131", "2313201", "0100312", "0123033", "0122213", "2203313", "2113011", "1103231", "2210032", "1320020", "2333031", "2122301", "2320221", "0031112", "2011133", "0233311", "2201003", "2221003", "3211320", "1330232", "0312333", "2220123", "3032102", "0123020", "2322310", "3202321", "3203212", "2203010", "2003331", "1112013", "3311220", "1010231", "0023120", "0123321", "1212130", "3121301", "1023233", "3202133", "0330122", "3023012", "0123332", "2231020", "3022031", "3230031", "2023112", "3210132", "1100123", "2311101", "2103232", "3302110", "2031211", "0103032", "1001321", "0323103", "3213013", "3212330", "1130201", "3021011", "0321120", "3212301", "0231012", "3202031", "1113201", "1120003", "3200231", "3220310", "2030010", "3320331", "1302233", "1123203", "3112012", "1032011", "0203010", "1023013", "3201002", "1320312", "0232201", "1302130", "2032111", "0311022", "2012333", "0332102", "2031101", "3302221", "0310112", "3312031", "0311102", "1131021", "3231301", "3000121", "2102330", "3233013", "3121030", "0023213", "1212300", "0332210", "1130122", "1030220", "1301201", "1000232", "2210033", "2020031", "1330020", "0231110", "0231231", "0201320", "2331013", "3323010", "1223310", "2322013", "2120332", "1310201", "2213030", "3130201", "3301203", "0331302", "1020302", "2110013", "2030231", "1221230", "0121313", "0312201", "2010113", "1201223", "0001231", "3002201", "3231103", "1233320", "1213120", "1101321", "2223001", "2120301", "3023310", "1332303", "3133021", "3100122", "2130312", "2012132", "1121103", "1000123", "0122231", "3100332", "0223301", "0130200", "3012222", "3210013", "1310233", "1322210", "2230123", "0332310", "1311020", "2222013", "1101232", "3223103", "0121321", "3320110", "3210022", "3021212", "2033101", "3120002", "0300123", "1131002", "3033321", "3010233", "1123110", "3112120", "1213130", "1022133", "3111032", "3003201", "2210321", "0121003", "1210231", "1102131", "1123013", "3020110", "0121103", "1223012", "2131202", "2001030", "2330311", "2303100", "2133200", "2301322", "2212031", "1230101", "1213101", "0103231", "3022213", "3130222", "1232300", "2303112", "1021332", "3121103", "2121033", "1110312", "3020212", "3111102", "2133230", "2032221", "0321103", "1103122", "1223011", "3123202", "3120032", "3233010", "0331032", "2130220", "0321133", "2323120", "0231003", "2031300", "3033120", "0021301", "2023312", "3102301", "1121023", "1020310", "0303321", "3203122", "1322031", "0330321", "0231013", "1321210", "1202233", "1230323", "3012211", "2321002", "0310220", "3303012", "2003312", "1301112", "2130001", "1120030", "0332311", "1301332", "2132100", "3012231", "1000213", "3331210", "2131030", "3301222", "1320323", "1320320", "2132103", "1220223", "3020013", "3201323", "3001012", "1032120", "2011230", "2310133", "2133001", "1013231", "3233011", "3021202", "1301122", "1320033", "3321022", "3101220", "2001013", "2213002", "3200120", "1030230", "2101322", "0201032", "1100213", "0232231", "3110323", "0033012", "2303211", "1131200", "2301032", "1220231", "0322122", "1302211", "3101332", "3210211", "0130332", "1232210", "1031211", "2331130", "3213201", "3220110", "1023201", "2102302", "3230113", "0213310", "3000201", "3223301", "0022311", "2032313", "1310312", "2030321", "3001231", "0120123", "3112200", "3303121", "0213030", "0121031", "0231213", "3202110", "1031200", "1022203", "1101213", "0333211", "1322023", "3130203", "3132001", "0123003", "0103121", "3321310", "0231233", "3112203", "3212202", "3102021", "0201332", "2203121", "0312311", "0131112", "1130022", "2030013", "2302121", "1233103", "0311223", "3021232", "1210302", "3131021", "1032223", "0003210", "2320312", "0203321", "0220133", "0313221", "2230113", "2031231", "0320123", "1312001", "0312121", "3000021", "0110213", "0301231", "2320010", "1332201", "0213112", "0232123", "1032323", "3302301", "0320313", "2322001", "2311032", "0131320", "2320031", "3310012", "2102310", "1013212", "3121022", "1113020", "3122021", "1023230", "1132220", "0110302", "1220031", "3301212", "2012231", "1333012", "0022312", "3033231", "1220013", "0102013", "0103312", "3223010", "3132301", "0012030", "0213002", "0312131", "3231302", "2100133", "3013022", "1230212", "2010321", "3003021", "2321003", "0121032", "2313001", "0201133", "0311322", "0321101", "1322002", "0322131", "1230331", "3312201", "0133020", "1202013", "1300222", "0322212", "1312130", "1233011", "2002013", "3010112", "1202322", "1323101", "0321310", "3012113", "0020331", "1201332", "2133032", "1210233", "0011223", "3220121", "1013012", "1121031", "3200310", "2003122", "1113202", "1022213", "0120203", "2031233", "1201131", "3121203", "2300231", "3221032", "0231232", "2303201", "3010200", "3001321", "0103023", "1210013", "1023231", "3032131", "1222103", "2100132", "0020231", "1202333", "1203032", "0132211", "0013022", "1133102", "3322201", "2102003", "3011223", "1301102", "3210330", "0323111", "1021320", "3031213", "3131230", "3103123", "2120033", "3210323", "1223001", "1030022", "3212030", "2013223", "0221223", "0321231", "0011321", "1002033", "3003221", "3212203", "3120022", "1330302", "3210122", "1203033", "0300201", "1220133", "2020312", "1122130", "2113003", "3321303", "0332112", "3002011", "3012100", "1330123", "3210020", "3322021", "1303222", "2221320", "0302123", "3202310", "2032010", "3211033", "3130230", "1102320", "0000213", "2101233", "0023122", "3322010", "0330102", "3202132", "2331303", "1220301", "1132202", "2100223", "0123231", "3230313", "3032201", "1130200", "2021130", "0131032", "0320210", "2110203", "1321302", "1321021", "3331032", "0211233", "0002301", "3302111", "2231220", "2102301", "1332033", "1203223", "2023120", "0323122", "3221101", "3020120", "2131010", "3132200", "0100231", "2130213", "1013122", "3003132", "0120030", "0310233", "1203313", "2131130", "1200223", "2133202", "1002032", "3210233", "2222130", "0123120", "2003310", "1302232", "3211130", "3203211", "1100223", "0132123", "1012301", "0111023", "3011203", "2112330", "2210330", "2031011", "0320131", "0111231", "2011213", "2311102", "1233120", "0300112", "1213033", "2012323", "3301211", "3030231", "1110213", "3210021", "2312320", "2211032", "3230321", "1320311", "2300331", "1203101", "1331201", "0322011", "0302210", "2202123", "2322010", "2103331", "1133032", "1130132", "2102131", "2200133", "0013032", "0330120", "2013321", "0013320", "1023102", "3201321", "3232310", "0123303", "1303332", "1312100", "2110310", "1302123", "0211132", "0021312", "3231000", "3322100", "3102010", "1130210", "3032012", "3021302", "3133200", "3312023", "1301120", "2130302", "3031012", "2031200", "0031120", "3201101", "1302122", "2032212", "3312230", "0122321", "3122001", "0012300", "0322001", "2313302", "1103221", "0100223", "1220113", "0031220", "0121300", "2131102", "0313231", "0231103", "3122023", "2302133", "0302131", "2101023", "1200322", "0130021", "1022300", "2300133", "2313003", "1121301", "3020131", "1032131", "3323012", "1102330", "2130212", "1203231", "0110230", "2311110", "2101333", "1230131", "2132001", "2013133", "0002123", "3222301", "1213220", "2102103", "2123310", "3011302", "0231303", "1332200", "2120123", "1202103", "1120313", "1012320", "0332010", "1032202", "2003012", "2312210", "2232001", "0003231", "3310202", "3021021", "3130032", "2330312", "2301102", "2101313", "2232021", "0333210", "3332012", "0330213", "1302010", "0321222", "1122023", "0122133", "3302103", "1203000", "0203123", "0231023", "2333100", "2103320", "3003231", "2020301", "1310321", "3302310", "0102311", "1231110", "0133132", "0113230", "1120320", "2300021", "3130132", "2202311", "0310211", "0103213", "0012232", "1033201", "0033212", "0032100", "2023211", "1233330", "2010320", "3102133", "2300130", "2113012", "3101221", "0232101", "1221302", "3111023", "1021323", "3321301", "2021030", "1203011", "3110012", "0033210", "0120033", "2001322", "3200301", "3201330", "3122000", "1233130", "1112031", "3112201", "0210320", "0220213", "3023001", "3232201", "3230312", "1201032", "1123302", "0012203", "3102313", "1130221", "2021230", "2022123", "1032130", "0331121", "0130302", "1033122", "3203312", "1302213", "3010023", "0002113", "0202123", "1330320", "1303203", "0233102", "1233021", "3121320", "3023103", "0223110", "0312112", "1213303", "3010230", "3301221", "3313120", "0103022", "3210011", "0331321", "1320000", "1310023", "0332103", "3203231", "2310313", "1110132", "3010202", "1320231", "0201231", "2303131", "3112003", "2100003", "1030232", "0031322", "1203031", "1120031", "2121303", "0200231", "3123033", "2220301", "1120032", "0311220", "2301331", "2332013", "3021310", "0213003", "1303230", "3102202", "3100233", "1210003", "0130212", "2031323", "0213120", "1023220", "3020111", "1323002", "3302131", "0021332", "1023111", "0232102", "0212232", "3132120", "1233000", "2123120", "1301210", "2102230", "3211012", "3202111", "3010211", "1001323", "0302122", "1032331", "2110301", "3302313", "2233001", "2233031", "2311000", "1211320", "3120323", "1123220", "2012301", "1222203", "0320102", "1303112", "0003121", "1310123", "3222310", "1230023", "1122030", "1013230", "0301022", "1323100", "2133201", "2210013", "1211030", "0020113", "1001232", "0300211", "3121200", "3002121", "3110002", "3202101", "0011231", "2201330", "0031200", "0322100", "3132031", "2330111", "0132101", "2131020", "3200331", "3001121", "3001022", "1332103", "1300321", "2000130", "3020312", "1231012", "1103322", "3120212", "1320112", "3110230", "1013323", "3201033", "2032013", "3303132", "0302310", "1123301", "0313222", "3001032", "0023201", "1120233", "3210311", "1020033", "3131320", "1230031", "2312003", "1200233", "1302212", "0100322", "3103102", "1031220", "1302300", "3213003", "1330211", "1031212", "1332203", "3032130", "3301232", "2301002", "2101031", "2210310", "2131100", "1022321", "0003321", "0210310", "2102132", "3101210", "0120300", "3130322", "0233133", "3312220", "1023031", "3303210", "2131110", "1302011", "1103312", "3020122", "0113200", "2100333", "0011230", "3010120", "3120221", "3132202", "1113120", "2223301", "1202321", "1102213", "0313332", "3322102", "1112302", "1203030", "1102321", "2001301", "0221331", "0322132", "3201331", "1010312", "1030320", "3212103", "3121310", "1002322", "3110211", "2330130", "1213001", "1020123", "0333021", "1003122", "2203210", "3022112", "3103213", "0301121", "1220303", "1003022", "2113032", "3013200", "1131210", "1023033", "3102232", "1302102", "0231202", "2110132", "3201320", "3313020", "2130131", "1020030", "0301122", "1230110", "1132000", "2211030", "1120132", "0132131", "3103321", "1022303", "2113120", "1332101", "0332321", "1332001", "3120302", "2213303", "2313320", "2023110", "3010021", "2020131", "0201302", "2132031", "1230030", "1132030", "3322031", "1021231", "3120300", "3312033", "2101301", "2330211", "1012103", "3010022", "0032130", "1112030", "2110230", "2111303", "0122323", "1123320", "3031230", "1023003", "1011203", "2120311", "0110323", "1320322", "3121102", "3220012", "0031222", "2013120", "2301110", "0202132", "0211213", "0233212", "1003120", "0203012", "0002103", "1033222", "2131003", "0010023", "1200300", "2231300", "2320213", "2031122", "0013233", "2010133", "1130231", "2132101", "2031222", "2122302", "1023303", "0112320", "1102322", "1201103", "2001133", "3120312", "2031213", "3222021", "0312313", "1231302", "2302122", "0113102", "1330120", "1013232", "2122030", "1003321", "2210331", "2130031", "3031332", "2033131", "3210103", "3112101", "0110233", "3122130", "2302112", "1132320", "3130121", "1033121", "1230113", "3112300", "1003223", "0102301", "1132003", "1132200", "1300023", "0312102", "2231310", "2322101", "1032230", "3302231", "1332023", "2302102", "2321210", "0301321", "0230102", "3230221", "0211032", "1103200", "1031210", "3212230", "3311102", "1031332", "3100120", "2020113", "1012302", "0303102", "2230310", "2123012", "2231120", "1200301", "2310210", "3200111", "3323130", "3022130", "1032311", "2100032", "1302323", "1212032", "2210311", "0132032", "2301001", "1203332", "1120133", "0112312", "2301121", "1321030", "0113022", "3213102", "0123021", "1122310", "3011122", "3120232", "1003312", "2023123", "1320132", "3013201", "0001203", "3120131", "2023311", "3111320", "0023212", "3321320", "3033221", "2213201", "0123230", "3330120", "3001232", "3333201", "1123330", "3002123", "0202213", "0222103", "2103213", "0333121", "3320102", "0200133", "2011331", "3003112", "2003111", "3130302", "3130211", "2203212", "2303013", "3302311", "0213013", "2102233", "2130122", "1002231", "1031320", "0320011", "0233312", "3211103", "2321020", "2312030", "1203131", "1120033", "2323012", "2033210", "0123032", "2121302", "1320113", "0321311", "2113001", "0330231", "3031122", "3220221", "0220103", "3123003", "0322123", "1312022", "0100132", "1331302", "0212311", "3202100", "1233001", "0200311", "3020102", "0123023", "0203331", "2313022", "2213103", "3021221", "3021312", "3212021", "0131020", "0323311", "2131220", "0321012", "1210023", "0132221", "1022332", "2001203", "3100223", "2232010", "0231221", "1322032", "0302211", "3020301", "1112230", "3132230", "2202213", "3312130", "0130102", "3210133", "1001223", "3113220", "3333102", "0112332", "3011112", "3102213", "2010332", "2032122", "3230012", "2321120", "0313123", "3013323", "3112001", "0032221", "3310233", "3103230", "1233102", "1213021", "2233120", "2103033", "3022122", "1231000", "0312200", "0122330", "2022131", "2113300", "2212023", "1233230", "2123301", "3111230", "3210331", "2011203", "2033012", "1303213", "0313120", "3002331", "0122300", "2101213", "0021303", "3321002", "3113320", "1230012", "2113100", "3000120", "2023321", "1302333", "2133011", "3312102", "1022013", "0231100", "0213133", "1003221", "2030021", "3321230", "0201103", "1031230", "0323120", "0001312", "2311033", "2031001", "0012023", "1211013", "3122320", "2311203", "3302331", "3100230", "0323012", "2201320", "3120123", "1200330", "1203110", "1233310", "3233310", "1213302", "2310103", "2011312", "2113301", "2110031", "2110133", "0031201", "2110333", "0302001", "2131021", "3202301", "2003313", "1310322", "1021313", "3201232", "1031120", "0102330", "3031112", "2223120", "3211013", "0321302", "2001332", "2002113", "2013220", "1032122", "2003132", "0321111", "3313032", "1022230", "2132203", "1303020", "0113012", "1023222", "2322100", "3101132", "3222102", "3131203", "0031212", "0130322", "0313002", "1102203", "3210023", "3203112", "2303210", "0203312", "2030301", "2112023", "3200112", "2011302", "2131310", "1013022", "1033002", "1310021", "3132203", "1022113", "2302210", "3310302", "0223131", "2101311", "2013111", "3013020", "3320321", "0132332", "2302031", "1323020", "1320300", "3320103", "2332101", "1013020", "2211003", "2031322", "1121013", "1231002", "0113202", "1103332", "1021213", "3213033", "1331230", "1330202", "3122010", "0101230", "3013302", "0012312", "1302133", "3110322", "0033123", "0101232", "1210332", "0320001", "3020100", "2031113", "0220331", "3021023", "3033121", "1230203", "3231320", "3100121", "1113203", "0321032", "1212320", "1302310", "0021302", "2200130", "2210123", "0213333", "1023012", "2013322", "0300132", "0331020", "2012330", "0112113", "0213023", "3032110", "0321013", "2121032", "0123112", "3230212", "3320211", "3133002", "2100213", "1103120", "3100323", "2232110", "2213200", "3301021", "2203102", "0301112", "1210303", "2032012", "1002030", "0030123", "0322111", "3230130", "2031031", "0312300", "1301023", "3311302", "0200321", "0320122", "3233100", "1022311", "2102123", "2123103", "0212003", "1001203", "1022003", "0102233", "3120101", "2320001", "0121232", "1302200", "3132100", "0233301", "3120133", "3130220", "1323203", "1210301", "3230133", "0012031", "1323023", "3220213", "0012213", "3013321", "2232013", "0012233", "1331012", "0033221", "3102132", "3003211", "3201213", "0231032", "1032012", "2132200", "3011120", "3021002", "2000132", "3021220", "3021013", "0211231", "0002311", "3121130", "3201132", "1321100", "1321301", "1232003", "2013012", "1113021", "0333012", "0302011", "2000331", "0133210", "3220021", "2120223", "3100221", "3321203", "2131022", "1310332", "0331210", "2223130", "3023311", "0010332", "1213210", "0232321", "3112021", "0201300", "1032000", "1333102", "2310121", "3213023", "2013100", "3100312", "2113013", "1232001", "3201022", "0200310", "3100212", "1001231", "0101213", "1032110", "0132313", "0023310", "3212100", "0203011", "0213232", "0112321", "3101112", "2103230", "0223133", "0032120", "1302332", "0132301", "0100302", "0231000", "1210223", "2123030", "0012133", "3331020", "1302312", "0001223", "3330210", "3022111", "3102210", "1103212", "0310002", "2033010", "2033312", "0130203", "3112011", "1133002", "0112030", "2310201", "1233012", "2300121", "2003013", "1203312", "3310023", "1221003", "0111213", "3321021", "0311002", "0332331", "3213011", "3102120", "3001220", "1201313", "1022132", "0033122", "2120303", "3312010", "2101230", "2312301", "2330113", "3223120", "1110203", "2100232", "0320021", "3021101", "0301203", "1220332", "2301031", "1003020", "0003102", "1003200", "2010213", "3310323", "3120012", "2233210", "0302100", "3310232", "2033212", "1321022", "3301020", "3031023", "2123022", "1131012", "0032113", "0322110", "2221023", "1201320", "2100103", "2230132", "0211013", "0221123", "1120131", "1030023", "2330321", "3002101", "2201132", "1201310", "0023112", "0311121", "0021300", "0311312", "0133120", "2301030", "1300302", "1233033", "0130230", "1130223", "3131202", "1210230", "3231022", "2130203", "2310220", "1123021", "3113210", "1120013", "2320313", "3020211", "0220310", "2003102", "0231121", "2231103", "1302121", "1102003", "3102030", "3223110", "1012233", "0021322", "2130202", "0312011", "1031012", "2023213", "2012311", "0023110", "1320310", "2300111", "3231002", "0231123", "1200003", "3210010", "0212013", "2230231", "0221130", "2200331", "1102033", "0321333", "3311021", "0130321", "2303132", "0311023", "0321031", "1102311", "0031020", "1133020", "1321010", "0133230", "1121320", "0212313", "3110332", "2033331", "3102322", "1001213", "1322033", "0230211", "2011030", "3231203", "0302331", "2021213", "0230001", "1031021", "1021312", "0122013", "1003023", "3013021", "2210223", "2231001", "2132110", "0121130", "3323110", "3111020", "2102232", "3211022", "1100230", "1200131", "0120311", "3110213", "2011320", "2311130", "2222103", "0131210", "2021133", "0102320", "1330112", "1202313", "3030211", "3331120", "3103021", "2100320", "2031110", "3313302", "1033032", "1120230", "0033211", "2030113", "0313211", "1022313", "2022310", "0211323", "0231020", "1232000", "3102012", "0021003", "3112030", "1221033", "3130312", "0032110", "0231010", "1302210", "1030032", "3110220", "1322030", "1100323", "2213010", "3210102", "2321033", "3033211", "1000321", "3000211", "3230210", "3200102", "1210133", "3110221", "0132013", "1130123", "3101212", "3230110", "2322103", "1322220", "1122103", "1301231", "1222230", "1033120", "0132133", "2022031", "0320133", "3102331", "2333110", "0013221", "0223031", "1301203", "1223022", "1312030", "1332003", "1233301", "0113121", "3332102", "0103132", "1011132", "2113110", "2112320", "2011311", "3323310", "3021313", "1220233", "3133032", "1003322", "0233132", "1302031", "0210233", "1232032", "1230020", "3123013", "3332130", "0021033", "1301021", "1003203", "0230213", "0310012", "2001303", "3012111", "1310120", "0331220", "0132330", "1020301", "1030012", "2231033", "0110312", "1220023", "3002122", "0132121", "1323001", "1132013", "3321030", "2302012", "3020103", "2233012", "2021332", "0332101", "1000223", "1031121", "3111201", "2100023", "3221330", "3233102", "1123022", "2300103", "2311120", "2323011", "1130213", "0032301", "3022013", "1103121", "1200231", "0110332", "2003101", "2232101", "1000132", "2030012", "0120333", "3221320", "0133231", "3232110", "3222012", "1103020", "2331101", "2321220", "2112303", "0102302", "3231003", "1233110", "2301023", "0110132", "2231010", "2021301", "1201133", "0023132", "0021320", "1203123", "2301120", "1200033", "0123310", "2132310", "1312303", "0020031", "1233013", "0330212", "2011300", "3030021", "3130212", "0211113", "2131303", "2133303", "1302101", "2123330", "3032231", "3210230", "2310101", "0130023", "0210301", "2200213", "3122200", "3110320", "1211032", "0113210", "2223100", "0123100", "3211203", "1031122", "0032231", "1122032", "0001213", "2132300", "1311302", "1002132", "3003123", "0131202", "1202003", "0323313", "2301302", "3132110", "2130103", "2120013", "2232120", "3321130", "0201303", "3011233", "3120311", "0001123", "3311012", "1230221", "0220013", "0332111", "1300132", "0323121", "3111022", "1300210", "2210031", "3200021", "1011232", "0002331", "1012330", "0201333", "0222132", "2212013", "2302101", "3201012", "3102303", "3312100", "1002302", "2310320", "1303132", "2031223", "2011322", "1310230", "3023111", "3032213", "0132023", "0312023", "0102213", "1330210", "0113221", "0223211", "1300123", "2203112", "2200312", "0120312", "0221302", "1021322", "3002131", "3301210", "2010313", "3001122", "1213320", "2203123", "1020312", "3320132", "0010213", "0303132", "1203311", "1023200", "0332201", "2112230", "0122033", "0013302", "2103101", "1012333", "3010222", "2120213", "2032312", "0313121", "0021333", "1020311", "2010030", "1201232", "3110222", "3121011", "2110311", "3100012", "2010130", "2213210", "2312012", "0233110", "3313012", "1303323", "1230211", "0331332", "3301230", "1302222", "0102030", "1322100", "1033112", "3210000", "0320100", "0321301", "2310202", "0233103", "2311210", "2320103", "2121310", "0103233", "3301320", "2001300", "1210123", "2000312", "1002133", "0012301", "2310032", "0302130", "2130221", "0112023", "1203002", "1300211", "3312310", "3120113", "2030110", "2310310", "3031321", "1222003", "1220313", "2301010", "0032312", "3200113", "1032312", "3332310", "1331202", "3100220", "1022310", "2120003", "3121210", "1010302", "3333120", "1202131", "0213022", "0331320", "0231120", "3233001", "0111223", "1130312", "2232210", "1323302", "0221321", "3031022", "2303103", "2202013", "1223320", "1100232", "0021223", "1310221", "3221302", "0112003", "1133210", "3013212", "0213000", "2030001", "1323300", "1233302", "3210212", "0310221", "1331200", "3322012", "1203300", "3002132", "3201112", "1213023", "3202311", "2120030", "0301222", "1320003", "0333321", "3031221", "0022213", "2103200", "0133232", "1120302", "2023001", "1130321", "2211203", "3013132", "3110231", "3213030", "0001023", "1111320", "1123200", "1032112", "3012213", "0131332", "2310311", "0301002", "1202031", "3100002", "1331203", "3101123", "3312110", "0323010", "0130323", "2212003", "3001123", "0021203", "1001233", "1013223", "1002232", "0332011", "1230220", "2130100", "0213200", "1230231", "2203100", "1321003", "3221202", "2213023", "3313230", "0130211", "1210203", "0130221", "0110123", "2013131", "0133203", "3320221", "0112233", "3221031", "1303221", "2132201", "3021121", "0312301", "2033113", "0321132", "0200123", "3112023", "1203121", "1233020", "3032001", "1032333", "0323211", "1233022", "1201323", "1002311", "0113312", "0312100", "0021321", "2200131", "3301332", "0323212", "2013000", "0301320", "3121201", "2130313", "3012102", "0032211", "3031302", "2201313", "3010203", "1300221", "3023010", "3001323", "1312201", "1313021", "1003213", "0131222", "0213020", "1101322", "0130112", "2311002", "2330301", "1332130", "1020303", "0131022", "2333102", "0322213", "3133022", "3200201", "0200313", "1310323", "0200301", "2230013", "1012231", "0132000", "3133023", "2012321", "0020213", "3120100", "0112310", "3302121", "3321100", "0133222", "0213101", "2133203", "1321013", "3102233", "1020232", "0310200", "0131203", "1032013", "1212033", "2120023", "2210203", "2300013", "0213130", "2200103", "2012332", "1002103", "0230110", "3203213", "0332221", "1020231", "0212230", "0303122", "1010032", "0230021", "1202023", "0230122", "0303112", "2231000", "3330211", "1330221", "0310332", "2310012", "0102331", "1013312", "0011232", "3210322", "0321300", "3303201", "2013002", "1131320", "3212101", "3102032", "0211130", "0301323", "0130213", "1230200", "1132101", "3231023", "3122100", "2103312", "1110231", "1103211", "0103112", "3211220", "0321130", "2320211", "0322231", "0103320", "1300223", "2023021", "2202132", "1023133", "0211330", "3310112", "1321220", "1003212", "3023312", "3012212", "2302312", "0300122", "2300110", "1211330", "2021131", "2001131", "2100302", "0213213", "1020131", "2320301", "3101102", "1302120", "1233220", "2303121", "2100311", "3130102", "0103202", "1030312", "0113212", "2023231", "2010330", "0211103", "3121032", "0023031", "3313102", "3103233", "2000031", "3311203", "0120130", "3122032", "1320130", "1002230", "0311321", "1022333", "3100021", "3011332", "3212010", "3013332", "3210303", "0210332", "3200122", "3100112", "0231001", "1033022", "2231002", "1132021", "1312020", "2013023", "0020313", "1312300", "3002301", "1101230", "0210333", "1210311", "3130232", "3230211", "2032331", "2231102", "2301212", "0001323", "3230011", "2110032", "1202311", "2330101", "0133233", "3122013", "3310321", "0223221", "3313220", "1021333", "2330310", "2123033", "0130312", "0133102", "2322130", "3131201", "0132103", "1223220", "1322103", "1311023", "3201201", "2021113", "2132022", "0312322", "0023013", "3010232", "3021000", "2103120", "2301012", "0211303", "0300121", "1332120", "0021313", "1310202", "3113012", "3203301", "0312230", "1230223", "3210213", "2113230", "2011321", "1230310", "1302012", "2103210", "2210312", "1102133", "3320231", "3121003", "0120323", "1223031", "3001023", "0321331", "3130021", "0012320", "2333210", "1130002", "1223020", "2021232", "0322211", "3022331", "1113102", "1132012", "3121120", "1330200", "2303021", "1202310", "0123012", "0232120", "2030312", "0321323", "0230301", "2223010", "1012030", "2331200", "2302013", "1301022", "2312010", "2311013", "0231300", "0321202", "0203113", "2202312", "2203321", "1131023", "0123200", "0231033", "3112002", "3212032", "0223310", "0010320", "1031221", "2102333", "2012133", "0122310", "0303211", "0121132", "1120332", "0020103", "0333123", "0332301", "1313220", "1312031", "2121031", "1012113", "3021122", "1232010", "3012013", "1023011", "1220213", "0012330", "0020310", "3110233", "2231022", "0201232", "1110232", "2010003", "1223030", "3121012", "0130320", "0333102", "2011113", "3100032", "2331230", "1233002", "0221030", "0132110", "2102323", "3123032", "2103110", "0203110", "3330231", "3202120", "2323101", "2322012", "3210232", "1202301", "0231130", "3121303", "1101320", "3222210", "2230301", "2230331", "2212300", "0023021", "0301210", "3320122", "0310231", "0202131", "1030222", "1322000", "2003100", "0122203", "1012230", "3121002", "2302120", "1330012", "0120213", "2210130", "1203021", "1320223", "1322320", "0301020", "2203111", "0113332", "1232202", "1210313", "0002130", "2300321", "0121311", "1133230", "0331233", "1311210", "3131120", "1033202", "2132301", "3200131", "2033221", "0103210", "2001232", "0223103", "0213111", "3203013", "2310020", "3203313", "3301202", "0223100", "2001113", "0310222", "2130101", "1321031", "3202131", "2103223", "0130002", "0230313", "3210030", "1310122", "1032220", "3213031", "3102112", "2103211", "3110302", "0310021", "1102023", "1033232", "0001321", "3310022", "0321000", "2000133", "2002311", "0221031", "2031330", "3030321", "0331102", "0321201", "0322101", "3303212", "3202213", "3101223", "2010132", "2321031", "2002231", "2301332", "0312310", "0212203", "0311332", "1131220", "3021022", "1321110", "1132032", "1113220", "2220113", "0313312", "0121312", "3021333", "1302013", "3211102", "3031203", "0223321", "0013222", "0232310", "0032101", "0131321", "2101132", "2210132", "2111013", "3010123", "3011323", "3201121", "3103112", "1022123", "1012322", "3300210", "0020123", "1220300", "3123200", "1312000", "2311201", "2121103", "1013032", "3121013", "2123031", "1021321", "0322102", "3210012", "1010321", "0123010", "1233201", "1301312", "3212011", "1323110", "3321020", "2013001", "1333203", "0100321", "3102033", "2230321", "2303310", "0302212", "2130120", "0213312", "1200031", "2001033", "2123021", "3102003", "0223121", "1023300", "1221013", "2132102", "1203333", "0233221", "3001233", "1233032", "0120113", "0100320", "0310132", "0201322", "2113203", "2301202", "0031032", "2012320", "3231031", "2300310", "2130020", "0012302", "2130111", "0013021", "1101203", "1120213", "0320321", "3332110", "0232122", "2000321", "1312220", "1012311", "0132212", "2203331", "1210132", "1333220", "0013132", "3202113", "1112130", "3203311", "1022331", "3021323", "3022212", "2112031", "0331200", "0230100", "3011201", "0012310", "1210232", "2103333", "3302321", "1213203", "3103122", "1130012", "3112303", "1011302", "0021013", "0102023", "3221103", "0031023", "0323102", "3311320", "0010322", "3020210", "3332101", "0213211", "2320112", "1333120", "3110102", "2033130", "3221220", "2332310", "3030102", "2003110", "3220212", "2013122", "1223033", "0312101", "2201302", "3012300", "2101131", "2033121", "1232101", "3230102", "3203100", "0200312", "0313212", "1120333", "0230011", "3130210", "2011003", "0101321", "1022330", "3023132", "2303001", "3112032", "2231202", "2321103", "2332103", "2300313", "1321033", "0132321", "2002313", "2131330", "0123132", "2120233", "1210032", "3320111", "1212030", "2210302", "2301021", "0231320", "3021112", "2300031", "3311200", "0022321", "1312320", "2311023", "0122132", "2113021", "0320110", "1033132", "2013300", "0303231", "1233202", "2213101", "1233010", "3003213", "1010322", "2313102", "2203221", "3231012", "1322012", "1100322", "3301321", "1132302", "1000231", "3202121", "1012131", "0212133", "2033231", "0200331", "1131201", "0131302", "3213303", "3032021", "2133330", "2110312", "1301213", "0033231", "1303210", "0021023", "1202032", "1102013", "3232130", "1031201", "0310302", "2231013", "3110210", "3300012", "2030102", "3100320", "3300122", "3331202", "3103211", "1301211", "3033122", "0112300", "0132303", "0210203", "0120232", "1313022", "3223101", "3120321", "3103023", "3200133", "3001211", "3300112", "0303210", "3210002", "3130221", "3133203", "1220103", "1112301", "3120320", "3001020", "3120112", "3120132", "3012023", "1023002", "2322011", "3023121", "0011123", "3310322", "2202103", "1201030", "1232012", "2312013", "1203112", "0213320", "2233010", "2103000", "0012331", "0212333", "1201322", "1002303", "0333231", "0101302", "3311120", "2023031", "1103123", "0211232", "0113232", "1023103", "3312032", "1200132", "1330321", "1213011", "0223132", "1330023", "3010201", "0201023", "1003202", "0123022", "0302112", "2130113", "1002301", "3021110", "2311103", "2113303", "2102223", "3011322", "0203133", "1213300", "2221303", "1302032", "2031220", "2301210", "2303102", "0213302", "0201301", "0013012", "3020201", "1203201", "1332030", "0122223", "3313201", "0030021", "3102023", "1023032", "2103330", "1002312", "2013022", "3221020", "1321002", "0103230", "2330201", "0032021", "1033020", "0312122", "3101213", "2021023", "2220310", "3031212", "3313002", "3221230", "1032330", "1230010", "2033301", "0022331", "0321210", "0023221", "1330230", "1130211", "2022132", "2103002", "3212200", "0031102", "1320200", "2132303", "1310032", "2210313", "3312020", "3321201", "1310002", "0320212", "0022132", "0010233", "3223210", "3210112", "0211003", "0122232", "1021230", "0203103", "0213220", "0211123", "1232301", "1230011", "0301012", "2030121", "1120303", "3001132", "2110213", "2301220", "1222023", "0123111", "1023232", "3123302", "0231230", "1312103", "1303321", "1013302", "0203211", "0010132", "1200230", "1023113", "3303102", "0031230", "1021003", "1022023", "3232103", "0331112", "1213013", "3302210", "3203121", "2302100", "1320202", "1300202", "0210113", "3221303", "0200103", "2331002", "2112033", "1222302", "2122130", "3231102", "0102333", "2231200", "1310132", "1131203", "1223200", "3123001", "3013213", "3220211", "3221110", "0301032", "1013002", "0120003", "0310123", "0123223", "0233001", "0021232", "0123110", "2112003", "3011202", "3213022", "1113302", "2132220", "1120203", "3101232", "1203022", "2312201", "2110023", "3221130", "0210003", "2013313", "2103332", "1232220", "1031222", "1003032", "2013113", "2311003", "2201131", "3312330", "0030321", "1031213", "3301012", "0331222", "2002131", "3112110", "1203132", "3123230", "0112231", "3000231", "1330102", "0031021", "2023331", "3022120", "2330212", "2312032", "3201000", "2232031", "0332211", "3032331", "3331203", "3020001", "1033221", "2300311", "1122013", "0322312", "2203122", "1030121", "1311202", "3132000", "1002333", "3101233", "2321010", "2031120", "2010301", "2001123", "3212001", "0321220", "2121300", "3211200", "2033311", "0323130", "2022311", "3020123", "3300132", "1023331", "3210101", "2201023", "2302001", "2030111", "2333101", "1231320", "2003211", "3020221", "3120120", "3113202", "0310122", "1200203", "3011321", "3321210", "3211201", "2001330", "0112323", "0230221", "2310003", "0320013", "2313202", "0032311", "2113330", "3023331", "0132021", "1100320", "2231030", "3031220", "3321023", "2213011", "0012333", "0202331", "1030322", "3203131", "0031210", "0133021", "0123233", "0223130", "2301011", "1030213", "2301223", "2301231", "3103332", "1133203", "2001323", "3031320", "0130202", "3211002", "2321130", "2321001", "3120301", "0020311", "1213002", "2313220", "2213130", "1330132", "2213102", "1303233", "2201013", "3013221", "3020311", "2203011", "3123103", "2100332", "3120031", "3322130", "1320102", "0012223", "3020213", "2023013", "2010232", "0103332", "1021300", "1112303", "0122031", "2312020", "3311020", "3103222", "0021130", "2313210", "0301230", "2312002", "1230032", "2120232", "2310222", "0231112", "2003231", "2201300", "0200213", "3013120", "2012103", "0320310", "0231122", "0221301", "3122011", "2133210", "1032111", "3212220", "3313320", "3230213", "0332133", "2203130", "0233123", "2211031", "1101223", "3023031", "3120222", "0230310", "0021113", "2100131", "2211013", "2300120", "2230001", "2331031", "1103213", "2323031", "3100302", "1302320", "2221203", "3321200", "2130130", "2112302", "3210312", "2100303", "1232023", "2213300", "0100123", "2100230", "1312330", "2331302", "1203013", "2130133", "1110223", "3120220", "2031111", "2201033", "3211020", "1310210", "2031010", "0123130", "1222032", "3210032", "3320123", "2021013", "2130023", "0133221", "1123230", "1203113", "2310223", "3110021", "3022133", "0222213", "3231020", "0212123", "2031320", "1223301", "1223101", "1323033", "2311100", "2101303", "3122303", "0012032", "2300201", "1112203", "2012003", "3110022", "2310330", "0312330", "1012313", "0302313", "2230213", "3130200", "1022130", "3021331", "3233012", "2323201", "3132102", "0212233", "0232211", "1003210", "0233201", "1231202", "2001132", "0323312", "0331201", "1231120", "1032210", "3331023", "2030221", "1321320", "1010320", "3302211", "3202212", "1230121", "2303301", "3210220", "2001313", "0110232", "2113220", "2313330", "2013311", "2211302", "2101003", "3113230", "3002031", "2032310", "2302212", "2130301", "0102300", "2020133", "0021030", "0301302", "0321213", "1213032", "1033213", "2311330", "2001003", "3100102", "1021131", "3121330", "3103203", "1301323", "3200210", "0020312", "3121300", "1011321", "3330102", "0131023", "1032200", "3310231", "1223210", "3212120", "3103132", "1033102", "2330102", "2211301", "1310302", "3120102", "3000213", "3123011", "3313021", "3111302", "0010223", "2123203", "2201312", "0201033", "0312203", "1022323", "0022310", "2003123", "1230322", "0131231", "0031223", "0121320", "0233310", "1203213", "0312022", "2132003", "3030123", "2233130", "3001002", "1121300", "1331102", "0321223", "1201023", "0330221", "2123200", "2301221", "2231210", "1100321", "1323230", "3103231", "0003021", "2010230", "3011032", "1122302", "0112033", "2030313", "3132300", "3213300", "1210312", "0202311", "0131201", "3320001", "0013002", "1001322", "3212013", "3201131", "1223300", "2212103", "0320312", "0223120", "0231030", "3102332", "0120023", "1302302", "2130121", "3310032", "1313210", "0112203", "3331021", "1232030", "0232121", "1121033", "1131302", "2002031", "2303113", "3302123", "2123303", "1021203", "0003312", "2332110", "0312002", "3021233", "2122023", "0132033", "0332021", "2233102", "3121023", "1210031", "1103022", "3210320", "2302130", "2300100", "1013203", "0132012", "0123203", "2330103", "1113210", "0121023", "2021032", "1003220", "3102111", "1002313", "1113022", "3231230", "3222013", "0320031", "2113302", "3332031", "3023013", "3220313", "1221300", "0112302", "1231201", "2310212", "0331312", "2112310", "2023010", "3201102", "3010231", "0211300", "3130213", "2133023", "0112330", "2013003", "2210113", "0121013", "0232131", "0323310", "2133021", "1300323", "0021310", "1230330", "2301320", "0123313", "1230001", "3330212", "3023301", "0012332", "0230111", "3012031", "0130222", "3011210", "2020213", "3233201", "0201313", "0230331", "2032132", "2103301", "0323301", "1023323", "1023121", "2310102", "2221103", "2010123", "3303123", "0023121", "1311012", "3201111", "3013320", "2321302", "0103221", "2102030", "3202001", "1202133", "0331212", "3302001", "1311002", "3303211", "0303123", "1101231", "1223120", "2211330", "3122033", "1100032", "2133130", "1030200", "0231102", "1312003", "0111323", "0212321", "2310302", "2003210", "1120023", "3220131", "0213110", "0203213", "3330123", "3322101", "0103322", "3310223", "1001320", "2121203", "2332001", "3323102", "0312233", "1110032", "1320002", "2023221", "1211302", "0123030", "3101312", "2232201", "0131122", "2331210", "3323103", "2311303", "1312021", "3221022", "3211023", "2103231", "1332330", "3203010", "2310111", "1021013", "0120320", "3012330", "1031002", "3211301", "3020031", "1202231", "1213103", "3102333", "2301301", "3032013", "3332120", "1321103", "2330221", "2311001", "1303002", "2333010", "1322300", "2103220", "3101323", "1223102", "1120322", "1330220", "3330021", "2330112", "3321103", "1132002", "0231322", "2230031", "3001102", "3302212", "1332021", "2330011", "2031020", "0112311", "2210103", "2203120", "1030203", "3022010", "3001201", "3230001", "3030210", "3212023", "2221032", "2332301", "2200311", "1202030", "2201332", "0312010", "2112301", "1323103", "1210330", "1302330", "2031112", "0233113", "3003102", "0221310", "1201321", "1310211", "1322120", "3012322", "3002212", "1212330", "2130322", "0123101", "3123330", "0011023", "2302331", "1031321", "0332110", "1123002", "0112322", "1032320", "0232312", "0321332", "0301023", "0311302", "2120131", "1300212", "3012101", "1023123", "3131210", "1120103", "1012123", "0320201", "3201030", "1203321", "2230312", "1013021", "0032313", "3023133", "1300121", "2301313", "2331000", "3213002", "3132023", "3011012", "3312000", "3220312", "2200313", "2101302", "3200010", "3210031", "0122331", "2300122", "0103321", "0032213", "0302132", "3201211", "3332210", "3311230", "0131223", "0202301", "2113210", "0213122", "2133120", "0222131", "2330021", "2001333", "1333023", "1203212", "3033210", "1220312", "0330132", "1020103", "3321102", "0213321", "0232331", "2203132", "3212130", "1220232", "2313310", "3021020", "2003120", "1012213", "0130233", "1113032", "0102033", "1312032", "1132303", "3022102", "3101012", "3112000", "3200123", "1332032", "3000012", "2310031", "0322012", "3033112", "0220231", "1320121", "0202312", "0022031", "1131020", "0220311", "1223302", "0212031", "0313203", "2210320", "3220210", "0200013", "3100321", "2301201", "0103123", "1220030", "3113200", "0112303", "0213210", "1000322", "2013013", "2310331", "1121230", "3032010", "0313322", "2230102", "1022231", "3112220", "2122230", "2012213", "2021103", "1123210", "3320212", "3103020", "3100322", "3210100", "3303221", "2203113", "0132220", "3213310", "3231033", "2031331", "2130300", "1001332", "0232313", "0132122", "0010312", "0032011", "3332013", "3022321", "3031322", "0221333", "0212033", "0213033", "2033201", "0132002", "3120103", "1022103", "3022312", "3201333", "3213320", "3202123", "0121030", "3312003", "1202302", "3102312", "3201011", "1123000", "3001302", "2111300", "2111203", "1102310", "2211310", "3102221", "0223012", "0230210", "2010131", "1000312", "1103021", "2110320", "1023020", "0231311", "3310211", "3011213", "3012202", "3330122", "2131302", "2112203", "3122203", "0131220", "2100123", "3321012", "0302201", "1200320", "2031131", "1002323", "2320101", "3011230", "1230230", "0032112", "3013102", "1010332", "1323030", "3012302", "3231220", "1202113", "3312302", "0233331", "1323010", "0210311", "1230202", "3221033", "0231330", "0313023", "3031123", "2321201", "0231220", "0100230", "0113132", "2200231", "3220112", "3301200", "1013322", "2301330", "2320120", "1110332", "0122131", "3232031", "3010302", "0123001", "3120011", "1002332", "0033102", "3301322", "0011323", "2013230", "3300121", "0320311", "0302101", "1133023", "3020101", "3102103", "2101033", "1032332", "3211230", "2331001", "0112232", "2120300", "1102132", "1231210", "2320212", "0003213", "3232301", "3031222", "3201110", "3101120", "1230033", "0120310", "0330123", "0031213", "3203102", "1303220", "3130223", "1232203", "2013330", "3320120", "1230332", "1130220", "0113023", "0123123", "1203331", "0031012", "2310303", "2010333", "3120202", "3203210", "3020313", "0203101", "3032101", "3332021", "0220301", "3012130", "0030210", "1130120", "0032103", "2301131", "1223003", "1232200", "0203111", "2121301", "0013203", "1232120", "0332100", "3302213", "1000023", "2331310", "1220331", "1320120", "0301322", "2003121", "1203020", "0123333", "1130121", "1320222", "3102223", "1021330", "2013101", "3231210", "2221130", "3322011", "2231130", "2133320", "3102013", "3331200", "3212031", "1033332", "2021333", "2203012", "0222311", "3130122", "1023132", "1203133", "2223103", "3032221", "0312012", "2320310", "1120113", "0210132", "3102123", "0123320", "1333200", "2001023", "1310012", "1201031", "0322130", "1032213", "0132020", "1102223", "2001213", "1301222", "3021123", "2201223", "1000032", "3311202", "0221312", "3001120", "3023313", "3101320", "0102313", "2203133", "2012013", "2301132", "0112213", "2211130", "3123110", "0332121", "1332000", "2232100", "3202210", "3021201", "2110321", "0221233", "3120310", "3213302", "0120332", "0321230", "1331220", "0133211", "1032101", "3121101", "0100332", "2323100", "2120103", "3120313", "3202221", "2113033", "3021320", "0002312", "3121033", "3201103", "3013222", "3101122", "3012323", "2330100", "1211130", "3021010", "2310120", "1023213", "1203003", "1300203", "1022312", "1313320", "2110131", "2012310", "2232012", "1022233", "1132103", "3201212", "3131102", "1110023", "2003010", "3221013", "3211000", "0212131", "3302201", "1323330", "3022301", "0332113", "1333022", "1210103", "3130323", "0201330", "1003332", "0012303", "1023221", "0221313", "1321101", "3132302", "3231011", "3310002", "2220133", "0132010", "3203103", "1312230", "2311031", "3020112", "3003012", "2013333", "1200321", "0233213", "1220131", "2032113", "2030213", "1010213", "1013220", "2112300", "0132223", "2032001", "1123100", "3201210", "2221033", "2303123", "3001021", "1032231", "1320221", "2033123", "3101022", "2003212", "1112003", "0203001", "1231003", "1112023", "1311230", "3110020", "0111320", "2213302", "1203010", "0113320", "3021230", "3331320", "0031312", "1023110", "3122020", "2122330", "1033302", "3132020", "1202303", "1300220", "2303133", "0030211", "3132012", "0131221", "0132031", "2011132", "0313112", "2110303", "1130202", "0203100", "1030120", "0002132", "0122303", "2011310", "1123033", "0131211", "2130132", "0122302", "3203113", "3213103", "0310323", "3011132", "3000112", "3230120", "2012123", "2220130", "0223201", "1212031", "3102110", "1030122", "3023211", "3200211", "1123030", "3033132", "0032123", "3221001", "0312210", "0130032", "3122101", "1300201", "2133302", "3213230", "1121032", "1313012", "2301130", "2230100", "3301120", "0211030", "1013121", "3002012", "3203031", "1023310", "1332022", "3210203", "1011230", "2130303", "0202103", "0023131", "2132000", "0133121", "0122332", "1330032", "0313233", "0113002", "2131320", "0201230", "1013321", "0301233", "1211033", "0331123", "3002111", "0223312", "3231032", "2203101", "2330012", "3312300", "3132002", "2012223", "2033321", "1202323", "1220302", "0013321", "2210233", "2023103", "0203131", "3123002", "0301232", "2013110", "2333301", "2312203", "0223113", "1301032", "1221031", "2331010", "2331300", "3023221", "1300320", "2202321", "2300001", "2311310", "0302012", "2031313", "0102032", "3021303", "3320101", "2110003", "0013211", "0213221", "2312101", "1132203", "1030201", "0232130", "0101231", "3132013", "0123323", "1322110", "2130211", "1210322", "1331002", "1303122", "1321020", "2101032", "2121013", "3113120", "1101302", "0330021", "2331330", "2201031", "2301213", "0123133", "0023102", "3010102", "3223310", "2121320", "1223130", "2110103", "1011320", "0321021", "2303012", "2013130", "2031033", "2313110", "2021233", "0213032", "2202301", "0300213", "0230101", "2310033", "0121330", "0213230", "3020012", "2320130", "0132323", "3112033", "0202321", "2320011", "2013030", "3030212", "3221010", "3122031", "2112130", "3122102", "0213332", "0230031", "2021313", "3103232", "3013322", "1032123", "3000212", "1233030", "3012203", "0132311", "2012113", "1331120", "0313220", "3333021", "3033201", "0102133", "1210213", "0031321", "0232010", "3001213", "2103212", "1102312", "2013202", "2131101", "2210231", "3312103", "1332102", "0222013", "1122033", "2312000", "0301221", "1200312", "0202031", "0320331", "3310332", "0120230", "2010231", "2023122", "2303313", "0131021", "0211310", "1320011", "0113220", "3121021", "2122320", "1002023", "1111203", "0203013", "0021323", "0132322", "3231201", "1033023", "1021311", "1231301", "0122030", "0312223", "0003132", "2313020", "0301312", "1020300", "2330213", "3210221", "3203120", "0103102", "0221322", "1000230", "1320123", "1303212", "3133020", "3331220", "2313303", "2103111", "3012020", "1013123", "2213320", "2032211", "3130112", "0300012", "1201132", "3123101", "2212230", "2100203", "1023203", "1100132", "0121231", "0312001", "1011032", "1120330", "0322113", "0213223", "3310220", "3022313", "0013123", "1320131", "1223103", "1302112", "1212203", "0033121", "3032310", "0220130", "3201003", "1023000", "1103002", "0312302", "3111120", "2332011", "3200313", "0332122", "0313320", "2013323", "3312101", "0310022", "3212002", "0031323", "2131002", "3102001", "3101203", "1020313", "0000312", "0013312", "2100130", "1030211", "3002001", "1333302", "0111123", "1302103", "0231301", "0132111", "0011132", "3312303", "1003021", "3023123", "2031013", "0233100", "1321303", "1010132", "0123202", "2310030", "1033203", "2023313", "0132003", "0103020", "1020322", "3102302", "1023210", "1232033", "1322330", "2030100", "3333210", "3301302", "0230312", "1031022", "1301002", "0131232", "2313031", "1322130", "1120300", "0112031", "2201130", "1032211", "3110201", "1203122", "0123131", "1123130", "1302002", "1330332", "1301132", "3311002", "3001203", "0122301", "3300212", "2320111", "1212103", "1332300", "2120113", "1011231", "1012031", "2013232", "3332010", "3303213", "0211312", "2303321", "3200100", "2313011", "0101223", "2310100", "1310220", "2120133", "3030201", "3012223", "1202132", "2203231", "1131102", "1203100", "3020113", "3103302", "0332001", "0302010", "3123300", "2302313", "0313321", "2130011", "2203110", "0210133", "2032031", "0013210", "1202203", "1023312", "1201013", "0022313", "0300231", "1231022", "1230123", "2133100", "3202313", "1103222", "0030102", "2302211", "1012133", "0132222", "0221023", "2120130", "2312033", "3200012", "0023210", "1333021", "3120332", "0211301", "2103003", "3330121", "2321100", "3122210", "2313130", "3301032", "3213120", "0322210", "1102103", "0321221", "2020103", "0131132", "1231102", "2301232", "2100033", "2011231", "2110300", "2310023", "0121131", "2300012", "1002223", "0311222", "0123300", "1223021", "2323010", "2132202", "1033211", "2101321", "1302311", "0022131", "3310132", "3013230", "1313230", "0212331", "2202331", "1112310", "0002321", "0123213", "3012120", "0010302", "0203132", "1103132", "2310010", "3002313", "1033230", "3202021", "1322013", "2211023", "3102131", "3132101", "1302111", "2330123", "3101231", "1211303", "3220301", "3012220", "3102022", "1332230", "3120010", "0120301", "1203211", "0201130", "2322031", "1103032", "1301233", "1331021", "3110200", "3031233", "1333210", "0103120", "1311220", "0021331", "2120132", "2103321", "3110122", "1320211", "1031323", "1023001", "2303031", "2012331", "1221103", "2311200", "2230111", "0220313", "0002031", "0310202", "2113130", "0331213", "1032132", "0302120", "2330110", "3020130", "2001130", "3031211", "2120330", "2012312", "2201323", "0201131", "0120131", "0311221", "3301213", "1313020", "2320331", "1203230", "1222130", "3220031", "2133101", "3021223", "0212322", "0302121", "1203111", "0211320", "3001223", "3312013", "0312211", "0232012", "0213012", "3123010", "1130020", "1222013", "1133012", "0112133", "3103323", "3110123", "0121223", "1221023", "0231323", "3322013", "2323103", "3101321", "0121322", "1210331", "2001103", "0203112", "2012130", "3001312", "0012033", "1010123", "3011002", "2012131", "3330213", "1203310", "0232311", "3001202", "3133202", "2131013", "2102332", "0030231", "0330121", "2212203", "2123001", "0312320", "0302311", "1103223", "3122330", "2130110", "0210230", "1323022", "0333312", "0033120", "0212023", "2130123", "1312013", "1332301", "3011121", "3032321", "0133213", "2123100", "1212023", "1323303", "0120303", "2032123", "2013112", "1233023", "3120233", "3210111", "3030312", "1330121", "0312003", "3010121", "0130020", "2123002", "2031002", "3013210", "2012233", "3203132", "1231031", "0113123", "1112033", "0102223", "0213331", "2022133", "0233121", "0321001", "0230201", "3312012", "2230121", "2210301", "0213102", "0122333", "1221303", "2321102", "0133320", "3123310", "3001221", "1102233", "2033310", "3212102", "2203131", "2301003", "3213100", "3131220", "1321032", "3021102", "3322120", "2031202", "0210033", "1020132", "0133321", "1301321", "1010223", "3221012", "2023210", "0313102", "0102103", "0232021", "2003213", "3100232", "1322001", "0213132", "2012031", "2000301", "0312130", "2103012", "3221301", "1323003", "1111230", "3303231", "1022223", "1231020", "1323220", "1302322", "0320301", "1332100", "2310333", "2133033", "3320201", "2123032", "0231331", "0030212", "3103220", "2223021", "2002132", "2132210", "0000123", "1031233", "2321101", "0322121", "3201312", "2310131", "2011233", "1123103", "1203200", "0221213", "3103312", "0211203", "0131230", "0213313", "3231110", "1221330", "3321110", "0320101", "2302110", "2011323", "1121303", "1101132", "0121203", "2131033", "2023121", "1203103", "0210223", "1101332", "0132120", "1030233", "3221011", "0212032", "3023113", "0212303", "1320100", "3012221", "3220101", "2310233", "2021132", "0102123", "0033021", "0100232", "1320213", "3110232", "1112330", "0323001", "2201133", "2331120", "3021030", "0311200", "2231031", "3230103", "2100321", "0132230", "3030012", "2122203", "1301012", "0102231", "3330321", "1010230", "2120323", "0223010", "2131031", "3030112", "0223313", "1000233", "2231201", "1132100", "1313202", "2223310", "2000131", "0221033", "3203130", "0230123", "2233310", "1010232", "3133210", "1230213", "1011123", "2021203", "2302113", "0222133", "3200130", "3200031", "2110130", "1002130", "0023211", "3222100", "0112131", "3122103", "0210030", "0312113", "1323102", "0112130", "2110113", "0123212", "3322110", "1330122", "2103313", "2031012", "1203202", "0320121", "1213301", "3100213", "1103220", "1301221", "3102113", "1300332", "1231013", "0213311", "1123101", "2022313", "1223203", "1130222", "0031233", "3130023", "0231113", "2210322", "0301200", "1310222", "3200221", "2121003", "1023212", "3302112", "0322313", "0301120", "1102331", "2330120", "3221021", "3321000", "3312022", "3230021", "0203122", "3000122", "1102130", "1101312", "1030021", "3033213", "2100322", "0312123", "2130033", "1310212", "3023122", "0221203", "1323032", "0212332", "0312331", "2332120", "1310200", "2130323", "0331120", "0113020", "1200130", "0011213", "1310231", "0223212", "3210120", "3130320", "3203101", "2313002", "3201100", "3323120", "2302231", "0312323", "1231033", "2003011", "2212030", "0111322", "1222031", "2013132", "2213013", "1210320", "3002213", "0303120", "2103201", "3320213", "1320212", "2210232", "1102323", "1001312", "1300020", "2032311", "2101231", "2322201", "3120200", "1023311", "3312301", "2312300", "0323132", "2033011", "2221302", "3311032", "3302101", "0131323", "3101121", "2221300", "1031231", "0312232", "0302221", "1230022", "2232011", "2230131", "2033102", "1332202", "1230002", "3100022", "0130220", "1321000", "2331220", "0231332", "1202033", "2103300", "1021130", "0123222", "2020321", "2031032", "1020321", "0023231", "3320121", "3310230", "3021103", "2130222", "3313202", "3320112", "3302012", "3011220", "1303022", "0313012", "2013302", "3031202", "1023333", "0223111", "1020133", "2102331", "1303023", "0312213", "3203111", "2313021", "0132331", "1010203", "2222310", "1100231", "3202112", "2301323", "3220331", "1301223", "2111130", "3320310", "2012322", "3000123", "0013102", "3331201", "0101320", "0213322", "3012232", "1331032", "3310120", "3201133", "1232303", "0102130", "2303221", "2133102", "2001031", "1213020", "0223231", "1023332", "2302311", "3123023", "2031003", "2213120", "2013310", "0213203", "2310001", "1103102", "1310213", "2100300", "3132010", "1320210", "2312031", "2101312", "0103232", "3110132", "2113200", "3023120", "3220130", "0113112", "3013223", "2030201", "1130112", "3102230", "3320113", "0201113", "1312302", "3123303", "2100233", "0313032", "0202313", "1303121", "2310022", "2030310", "2032131", "0200131", "0201203", "0001302", "3021131", "1033322", "2210131", "1130021", "3212020", "0221330", "0202310", "2031030", "1102313", "0011320", "3123210", "1303232", "2203312", "3103022", "0320221", "1002320", "2011232", "3031223", "3230231", "3310213", "2000213", "3310212", "2111302", "1100332", "1012323", "3321300", "2310000", "0032102", "2301300", "2321013", "1200310", "3322301", "0230112", "0132201", "0023101", "1320313", "1201301", "2101103", "3032212", "0221311", "0013212", "1301200", "0210330", "1102303", "2030031", "1132102", "0300221", "2310113", "0201321", "3212210", "0302321", "3023131", "3023213", "1230232", "3012331", "0310321", "2331202", "0321200", "2011333", "2302310", "2331011", "3222011", "2231100", "3320131", "1332302", "3120333", "3332001", "2031123", "3220201", "0130012", "2130223", "1122203", "3110121", "1012332", "2303122", "2101323", "3113302", "2201123", "1302201", "0033312", "2101310", "2320123", "1301123", "3011221", "1312033", "0013213", "0321003", "1021032", "1320111", "3120021", "1311032", "0132102", "1012132", "1220311", "1230021", "2202131", "3220010", "1313120", "1030231", "3203001", "2013200", "0023012", "2101300", "0213113", "2220013", "3321220", "3000321", "1323200", "1010233", "1021132", "3102300", "1321102", "3301122", "3231300", "3120033", "0333120", "0223102", "1033231", "0021133", "3120201", "3203221", "2310301", "2320133", "1312102", "2320132", "0133123", "2102023", "3213110", "1223100", "2003113", "1002330", "3102031", "0003012", "0223001", "1103012", "0333132", "2303111", "0303021", "2301101", "1032133", "0230113", "0311012", "3022100", "0012132", "0023011", "3312002", "2130010", "3011102", "0110231", "3211303", "0120031", "0230013", "0210303", "2102130", "1113230", "0103122", "2032112", "3010210", "1113012", "3221201", "0033213", "1302221", "2012300", "0212320", "2312021", "1000302", "1321330", "0230311", "1011223", "1103320", "3003312", "0120331", "0123221", "2013210", "0211230", "1020223", "2111330", "0333221", "3321101", "0310230", "2133310", "2103122", "2212303", "3203321", "3201021", "2000103", "1203001", "1320010", "0121333", "3330201", "1012331", "2331012", "0101322", "1110320", "3232011", "0332312", "2310232", "0131213", "0102321", "2213012", "1032322", "3320021", "0233130", "1102031", "0021230", "2122303", "0130120", "3212012", "3012233", "0133202", "3330012", "3310122", "2022331", "0232103", "2113310", "3210223", "0321122", "2121023", "1230300", "2231302", "1232031", "1302202", "0321002", "1323011", "0210031", "2320210", "2030311", "0103223", "1130032", "0332130", "1023120", "2023131", "3320313", "0201310", "1300012", "1130323", "2223201", "2211033", "0310223", "0210321", "1323031", "0233101", "1302020", "1231032", "0221132", "1132022", "3021133", "1321202", "2030112", "3101032", "2133030", "3012002", "1030223", "2011301", "0112331", "0022123", "3010213", "3211100", "3132320", "0032201", "0111032", "3201313", "1122003", "3103002", "2102032", "0013020", "0301102", "0012003", "3022221", "0312133", "0111233", "0323101", "1002233", "3231120", "2111003", "2102031", "2210003", "0111203", "1001123", "0302113", "2031212", "2323110", "1320230", "0123201", "2000313", "2231330", "0203021", "2320231", "2031303", "2032100", "3201300", "3101222", "0313202", "2121230", "1230302", "0330211", "3012103", "2333021", "2202313", "1230233", "1320220", "0130232", "1222033", "3010132", "3200101", "3223100", "2132020", "2312130", "0012113", "2220321", "1322022", "3012030", "3212320", "3202130", "2132010", "0233031", "2102113", "2012302", "3332201", "0112223", "0012311", "2100313", "3002100", "3000210", "1233003", "1003211", "0323123", "3100211", "3123021", "0310023", "0003120", "1021310", "2003321", "0112032", "3123201", "0011302", "3322103", "0021213", "1322230", "1032001", "2030211", "0200132", "2031023", "1230122", "3121020", "2103202", "0330201", "2010013", "3231013", "2111310", "3221310", "1020203", "3032121", "1203222", "0132130", "0302103", "2120031", "0233010", "0312303", "1013222", "2313101", "3222110", "1212302", "1120231", "2010303", "3201113", "0322133", "3123031", "3011211", "0321322", "2323021", "3120203", "3312001", "2001320", "0101332", "3020021", "2331320", "1321130", "2102203", "3302021", "2333130", "0012123", "3100202", "0120322", "1100302", "3123203", "1323012", "3121230", "2202130", "1322011", "2203301", "0232133", "1310112", "2201322", "2320021", "3002010", "1030102", "2223031", "2123230", "2220213", "0011332", "3103201", "1220032", "3301102", "0111232", "0311211", "3110223", "2113102", "1232011", "3102011", "3233301", "0313022", "3202103", "3002013", "3201223", "0232100", "1330213", "0212130", "1302030", "2012303", "1203301", "1031312", "3122201", "3120230", "2013123", "3011021", "0003112", "3210310", "0000321", "3220321", "2001302", "3233130", "0131121", "1223201", "0301021", "3020133", "2310112", "2300210", "1200030", "1121030", "0030221", "1032221", "2233103", "1011023", "1023330", "1330322", "1031302", "1320101", "3123100", "0132231", "1203210", "2002130", "1122320", "1213003", "2323210", "2100312", "1232130", "3211300", "0230132", "1002113", "1232020", "2130310", "0313323", "2011130", "3002210", "1300032", "3122110", "0222312", "0322010", "2213020", "1012310", "0322031", "1013221", "3032311", "0221231", "0023113", "0111312", "0023130", "0012313", "1200123", "2133110", "1202330", "3133012", "3111012", "0211321", "0222310", "0030132", "0302031", "0113323", "1020230", "3132011", "0113231", "1213000", "1332011", "3321032", "1330233", "1132020", "3022201", "2210333", "1033210", "3310021", "3301220", "2032133", "1222301", "3010332", "1010323", "3211001", "0110322", "3012032", "2103001", "0331232", "2031102", "3002321", "2212032", "3221102", "2032103", "2301200", "3230101", "3201301", "1211310", "0132202", "0331021", "1300312", "0312032", "0120032", "2321000", "1120232", "2130330", "1132300", "3021301", "1310121", "0232111", "3023110", "1102230", "1333202", "2110232", "2101332", "1203102", "0332031", "0321131", "3332011", "3023231", "0321321", "2101232", "3231100", "2201213", "0121331", "3232012", "0133022", "3102000", "1032303", "1113320", "0021311", "1032310", "1123031", "1023321", "0313021", "1002123", "3100020", "0212302", "0311021", "2100030", "0010032", "0030120", "1100233", "1003132", "0312110", "3102220", "1210130", "2130230", "2302011", "3113102", "1201203", "1233210", "0311122", "3212201", "3211330", "1002013", "1200133", "0101132", "2303010", "2213033", "1332012", "3230311", "2011330", "0312332", "2111032", "2123220", "0030312", "3221100", "1110123", "2300312", "1032010", "3201001", "0302213", "1321230", "1020113", "1230320", "1322302", "2220103", "1033123", "3210130", "1201213", "0122320", "0322013", "2032102", "2302132", "1111302", "0122032", "0023103", "3002120", "0332231", "2011223", "2233201", "3331002", "1213110", "3231200", "3201122", "2133002", "1000203", "3321010", "3312210", "3020310", "3111203", "3011222", "0231313", "3320301", "1303223", "1232230", "3300123", "2013201", "1210033", "3213010", "3121001", "0131212", "1012300", "2230211", "2311012", "0211311", "0213021", "3103012", "3023011", "3022121", "2113022", "3000132", "1302231", "0120313", "2022130", "0320130", "2300123", "2023102", "3312011", "3332100", "2110123", "2010103", "0220031", "1201033", "2220312", "3201202", "0230121", "3011020", "1302303", "2001312", "0120013", "0232031", "3230010", "0102113", "3120331", "2132330", "1310320", "1023112", "2130201", "1123011", "2133020", "0322021", "2023130", "2123013", "1303322", "1021133", "0230212", "0103222", "3013032", "2201310", "3200001", "3031021", "0212223", "1230133", "0120132", "1220333", "2331033", "1221130", "2120231", "2332012", "3221030", "2130012", "2103113", "1223330", "0213202", "3021231", "1200103", "3001200", "0131322", "1030132", "2220311", "3132330", "2331030", "3213001", "1120321", "3213220", "0102303", "3012122", "2330132", "1020013", "2313030", "3213200", "2132130", "3303321", "1120301", "2102322", "2032120", "0203102", "1230311", "2131301", "1120323", "0102323", "1201312", "2013312", "3132303", "2213022", "1303200", "2132011", "0133122", "0212323", "1123032", "3231030", "1320203", "2233101", "3031200", "1232330", "2021303", "3322001", "0113322", "1231103", "2203031", "0223112", "0323031", "2023101", "1201303", "2012230", "0331221", "1002213", "0101233", "1211103", "3022310", "3223102", "0101203", "1300002", "1031203", "0002131", "1201331", "2310132", "3010312", "0303121", "1322301", "3023321", "2310203", "0120231", "0303213", "3013202", "1023202", "0231203", "1032003", "1203320", "0301123", "2212320", "0001320", "1230303", "1230222", "3032210", "0103203", "0321303", "1013132", "2123300", "3012311", "3220122", "3012230", "2230103", "2311030", "2310002", "3313200", "3221200", "3320133", "1231330", "3310203", "1323000", "1023302", "1132120", "3331302", "3203012", "2132012", "3222120", "0020301", "2003133", "2321303", "2212130", "2112030", "3321003", "3320312", "3012320", "2210023", "0031211", "2330313", "3001230", "1212301", "2311010", "0323011", "1023030", "3132130", "3103221", "0231210", "1322102", "2010203", "1220323", "0331012", "3130202", "2123302", "1231101", "0313223", "2030103", "2002103", "1203220", "3123220", "1103023", "3030221", "1201311", "3012301", "2231023", "2333103", "1233303", "1321023", "0002013", "3120122", "3121010", "0122233", "3220011", "0032031", "3033021", "0010232", "0221303", "2003112", "2101320", "0321020", "2013020", "0312033", "3213021", "0221003", "3132033", "3223013", "0332013", "3122022", "2221310", "1320233", "0230120", "3310210", "2201311", "2131120", "2230012", "3101002", "2103112", "2310321", "3302010", "2122031", "1232201", "1302203", "2133220", "2001311", "3033012", "1201233", "2322102", "1232310", "1300213", "2200013", "0212103", "3221002", "0022113", "2323130", "2031132", "1123003", "2122033", "1231010", "2300132", "3102211", "3023210", "2023111", "3101202", "2031203", "0231211", "3133320", "2033112", "1033200", "2122003", "0221332", "3221000", "1032232", "0021233", "2310322", "1123120", "2213301", "2300221", "3021130", "2001310", "0131200", "3211021", "2301100", "2302213", "2211300", "3020121", "2312001", "0231021", "3133120", "0230133", "1232021", "1302230", "2320122", "2221030", "0112103", "2321320", "3012200", "0311323", "1103321", "0033321", "3102231", "2113201", "0113211", "3210113", "3230121", "2030131", "0010203", "3201010", "2013301", "1132010", "1332020", "0201323", "3022101", "3312200", "1120311", "2300113", "2103032", "1210113", "3202013", "0123301", "3312202", "2213310", "0003211", "1230301", "1101023", "2332021", "1320330", "2330013", "2201103", "0312120", "2103132", "0132100", "2231301", "0132210", "1011233", "1120310", "1130023", "3303312", "1032022", "0333122", "1321300", "0032121", "0231022", "2021321", "2131011", "1210321", "3031020", "3100201", "3112302", "3231310", "2212302", "2230021", "3022211", "2133103", "2231021", "1032023", "2130331", "3310200", "3021032", "2130233", "3222201", "2003201", "1002300", "0121230", "2312202", "3122310", "0130132", "1021301", "0213131", "0001032", "2132033", "3210123", "1001023", "2302131", "0012013", "2020231", "1330222", "1032301", "2010302", "0231212", "1323310", "0213031", "1303102", "1002131", "2010311", "3331022", "2022321", "3022021", "3313022", "1003112", "2213330", "0313132", "0033132", "0132213", "0333213", "3013121", "2332010", "0203121", "0213212", "3112100", "1031223", "3123000", "0333212", "0321110", "2201301", "2321021", "0022133", "0310213", "2301123", "2002301", "0020132", "2020132", "0230010", "0133302", "1322310", "3203201", "1012303", "3111220", "1233031", "3030122", "1312310", "2331022", "2032130", "1311201", "3210321", "3231130", "2300010", "1322020", "1231130", "0210323", "2033211", "3220102", "1302100", "0222123", "1112300", "1003201", "3030132", "1023320", "2212310", "0223213", "1302223", "1312011", "1022302", "2331003", "3301123", "2213032", "0233012", "2103031", "1212303", "1310203", "0332131", "2312303", "2230201", "0313020", "2100013", "1032222", "3233021", "0133220", "1102232", "2010223", "0133212", "1213031", "1023131", "3301132", "1322201", "0213001", "2030120", "2330131", "3020321", "0310203", "2100113", "2321030", "2110331", "1230321", "2301020", "0023111", "3233103", "3220001", "3320013", "2111023", "2020013", "1130102", "1321200", "2130002", "0201132", "0032122", "3001212", "1102332", "2021300", "2313120", "0112230", "0312212", "1330201", "0132312", "1203322", "0310210", "1010023", "3303122", "1223013", "3100203", "0013201", "2122013", "0230131", "3132103", "2313023", "0311202", "1123012", "3020132", "2221301", "1003002", "2131023", "3201222", "2213001", "2122032", "2321022", "2003131", "1312010", "3120210", "3021222", "3310222", "0210313", "2021310", "1230013", "0310322", "3221300", "3012011", "0311212", "1301020", "2300101", "1203023", "0321320", "3132220", "0302133", "2113320", "0132112", "3012131", "0123330", "0231333", "0321112", "0230012", "2310013", "1013210", "2103123", "0123232", "2213230", "2331021", "0132011", "1323120", "0103002", "3210110", "1322010", "0113032", "3220111", "2300301", "0221103", "3330312", "0233231", "3202102", "1222300", "0322201", "2030123", "1230132", "2220231", "1302132", "3211010", "0103012", "1331020", "0301213", "2230221", "3300120", "2230010", "1210300", "1323013", "3223021", "0232112", "3301312", "2032321", "0320211", "1302001", "2111033", "3201221", "2130332", "3120223", "2022113", "3223130", "0230103", "2133012", "2120313", "3230123", "3002130", "1110230", "1302331", "2103011", "3012121", "2301113", "3212022", "1320201", "3330112", "3031210", "2210323", "0203221", "0122003", "2102133", "3303021", "3012021", "3302132", "1030321", "3321202", "3021311", "2030130", "2132021", "1330022", "2130311", "0101123", "2010322", "2310130", "0133200", "1132301", "3003121", "1002321", "2220331", "0311232", "1300233", "3012332", "0123322", "2210303", "0013231", "1320333", "3012001", "0000132", "0302301", "1101123", "1223303", "3032301", "0213231", "0102230", "0222113", "1203203", "0011233", "0320132", "2222301", "2332210", "1023211", "1032212", "2033213", "1303302", "2210133", "2110332", "0321010", "0312221", "0103212", "3321031", "2123210", "0321232", "3300102", "0130022", "0323213", "3200132", "3230131", "1020023", "3001332", "3211030", "0323331", "3000312", "2231230", "1221320", "2031332", "3230013", "3132022", "3320031", "3022231", "1313201", "3021033", "0210032", "1210323", "2131000", "1301220", "0031320", "3012123", "1021031", "2002133", "0122023", "1311200", "0201030", "2233100", "2001321", "1303201", "3323011", "2011131", "1032121", "0113201", "0320010", "1200311", "0012323", "1011323", "0222130", "1230201", "3120330", "2013203", "3213203", "3301233", "0321033", "2301333", "3030213", "0212301", "0321211", "2012030", "0232113", "3230331", "3301002", "0013200", "1302321", "1033212", "0123103", "3133102", "2003031", "2311300", "1131202", "1312120", "2312102", "3010322", "1230102", "1032201", "2032011", "2002321", "0030121", "2200113", "3012310", "3300211", "0102031", "3230112", "1200013", "2103303", "2031302", "1220321", "3001112", "3022110", "3310312", "2132002", "1320032", "3201203", "3123130", "2110033", "0223331", "0133332", "0221013", "1231023", "3122030", "0223210", "3122302", "1302220", "2120333", "2112103", "3123030", "2313033", "0110321", "2033132", "0110223", "1031322", "3201322", "1102113", "2130320", "1011312", "1330312", "1303021", "0010231", "3121100", "3333012", "0330210", "0310232", "1102301", "3021132", "3012110", "0122312", "3302113", "1321310", "0231201", "1132230", "0030122", "2033110", "2321011", "2203001", "0211332", "3002102", "0101032", "2201231", "2201232", "3010221", "1012130", "2131203", "2101223", "0312103", "2103310", "0331023", "2302103", "1020333", "2302010", "3232102", "2000231", "3021321", "2230130", "0021330", "2031021", "0332120", "1213201", "1022031", "1320301", "2101133", "1301302", "1022033", "0212030", "3200121", "2231032", "3211101", "2213220", "1133022", "2032231", "2303212", "3022123", "3121202", "1302022", "2031130", "3122220", "1200323", "0211223", "0210013", "2322120", "1220320", "0021131", "2322301", "0213121", "3032312", "2032101", "1032113", "0012321", "3102311", "1032203", "2101123", "3011022", "0302312", "3201123", "2021322", "2301111", "1020003", "0010323", "0212300", "1013112", "1301322", "0312000", "2313203", "3203110", "3113022", "0002213", "2101113", "3211003", "1213102", "2221230", "1230003", "1032002", "3021111", "2333011", "2102321", "2033120", "3202012", "3221003", "3312021", "1201300", "1211023", "0133032", "2311021", "2213031", "2123000", "0133223", "1203330", "2113103", "3211210", "1202320", "3201303", "0231312", "2103021", "0120233", "2021302", "0221032", "2313230", "2320113", "3321302", "2333013", "0110032", "0003122", "0132233", "2032110", "0311120", "2211320", "2110313", "1111032", "1210333", "2123102", "3320012", "0123302", "2120230", "1232320", "3112210", "3012201", "2033100", "0320113", "0120302", "1322003", "0323110", "0132203", "0320012", "0301220", "0301332", "1210310", "1103202", "0301202", "3130321", "1220230", "3112130", "3300231", "0031022", "1331023", "0212310", "0211023", "0213010", "3023102", "3210201", "0123011", "1033012", "1033323", "3021211", "2201233", "0331231", "1013332", "2331102", "3101021", "3030121", "1201113", "3122301", "0131120", "2131032", "2010023", "1013213", "1123023", "2113010", "3131302", "0110023", "1323130", "2013212", "3011320", "0103021", "1032033", "2301303", "2013320", "1030221", "3213210", "2210030", "3331012", "0330012", "3001322", "3033102", "3132201", "2301211", "1101233", "3130120", "3012112", "1300322", "3112330", "0002231", "2320013", "3130123", "2131201", "2301103", "1003012", "0133002", "1222030", "1323210", "0311210", "1032030", "3021300", "1200313", "1303320", "1301320", "1133200", "3022103", "3002113", "0123121", "0012131", "3130012", "2331103", "0031232", "3210301", "2323102", "1023322", "3123020", "0131312", "0102312", "2011303", "2230110", "3213130", "0332313", "0200031", "0233321", "0201213", "1311021", "0013323", "1201123", "1232100", "2103322", "3220311", "3132310", "0123312", "2021231", "2312302", "1123202", "0331022", "3020010", "2312110", "1023010", "2301022", "3031201", "3100222", "3022131", "0022103", "0221300", "0011312", "1031032", "3102310", "3322310", "0331322", "3302130", "3231021", "2001223", "2013102", "1311203", "1012232", "0103200", "1302023", "2211230", "0210103", "0300210", "2120302", "1312101", "2120320", "3300321", "3231303", "0113120", "3101023", "2122310", "0322310", "0023133", "3133220", "0023001", "0122311", "2200310", "3003321", "3220103", "0331323", "3000221", "3131002", "1120123", "2321300", "3232010", "2321023", "1200332", "2310110", "2330010", "2201230", "0233210", "2313301", "1232102", "2011013", "0313230", "2332102", "0303221", "0323210", "1213012", "1220203", "3302312", "1212013", "3032113", "1223023", "0312312", "0032133", "0031122", "1030323", "2310123", "0023301", "3201332", "1300231", "1231310", "3200212", "1002003", "0130223", "1321201", "3022132", "3100123", "1320022", "0121323", "2102313", "1102333", "1131120", "0302013", "0223011", "0133112", "1332320", "2011332", "1110233", "1023023", "0023312", "2033103", "1301230", "0221232", "1032302", "2320121", "1221032", "2201321", "2102013", "3220100", "3320011", "2311011", "0321123", "3133302", "1110323", "2321032", "3202201", "2000113", "0121033", "1100312", "2033021", "2101330", "0132232", "3310221", "3201032", "3013231", "3021330", "0032212", "2030122", "3332301", "0210232", "0323113", "0312132", "2223110", "3120110", "0222331", "1013201", "2302111", "1332110", "0130210", "0122313", "2301133", "0321022", "1331210", "3120211", "1103112", "2213203", "2303231", "3232210", "3203310", "2330122", "3130231", "2201333", "3112102", "0332213", "3323031", "1303120", "2310230", "0311112", "1200331", "3000102", "2103221", "0203130", "3010323", "1130203", "0321121", "1022301", "2301230", "3113002", "0212113", "2320321", "0320231", "0131002", "2231303", "1211301", "1331022", "2302321", "0101323", "2112032", "0112123", "3002021", "2120321", "3312030", "0231132", "2130003", "3230100", "3102212", "2301310", "0022130", "2310231", "0201123", "0321100", "1213100", "2232103", "0232110", "0123210", "3210332", "3201302", "1231011", "2311320", "3100231", "1301232", "1102300", "3212303", "3220120", "0330112", "3012012", "3220133", "1323021", "0013232", "2313300", "1332031", "1232022", "0032010", "3110112", "1130322", "1232013", "3032031", "0110320", "1132011", "1021331", "0021031", "1220322", "1022032", "2331203", "3102101", "3213020", "3002110", "3323001", "0323221", "0222301", "1003121", "2133300", "2300112", "0333112", "0112313", "2302221", "0032310", "0221323", "1232103", "3331102", "1232002", "3323021", "0032132", "1202300", "3012210", "2312103", "0313213", "0031002", "1122303", "1123303", "1231030", "1220123", "3011212", "1031123", "0113233", "3123301", "0112333", "2031201", "1213310", "3302102", "1310102", "0202231", "3201230", "1320023", "2222031", "3112022", "0330312", "0232013", "1022320", "2320100", "3031102", "1223002", "0100032", "0103201", "3233031", "0323021", "3321120", "3211310", "2312330", "3102102", "2312310", "1101323", "0301212", "3201200", "2210300", "2332100", "3231330", "0112301", "3022210", "3211120", "3023100", "2123023", "1020031", "0021103", "2122103", "3130020", "1132201", "3021031", "1321001", "2200301", "3010032", "1033312", "3231001", "1213030", "2022231", "1133202", "1320302", "2203021", "0113021", "0223123", "1301121", "3120023", "3302122", "3031121", "3022113", "0031123", "2001331", "3203021", "3211110", "2212330", "2130022", "2132120", "3102323", "3112013", "2223013", "3122202", "2301203", "1231001", "1201003", "0013120", "2002310", "1000323", "3112230", "2111301", "3323201", "3012333", "1032321", "2302201", "1032020", "0132300", "0030213", "1132130", "0132310", "2303130", "0021132", "3330132", "1123001", "2032021", "0310320", "2310021", "2011313", "2123010", "0123311", "2013221", "1320110", "0323100", "2111103", "2020311", "3201220", "1303123", "1320021", "3101201", "0331203", "0323013", "3023201", "2322021", "3022012", "3031120", "1030302", "2013213", "0020130", "0023331", "3023130", "0002310", "0100023", "2011032", "3132021", "0020013", "1131230", "0213011", "2022103", "2310332", "2130333", "0232301", "0303212", "1301212", "2101130", "0332212", "1130332", "1031023", "0100323", "2313032", "3012312", "3102121", "2103203", "2333012", "0220113", "1302113", "2130102", "1131022", "1302131", "1202230", "3210313", "0030201", "2203311", "0201331", "2100031", "0232001", "2021003", "1302033", "0121301", "0122130", "2212301", "3121302", "2012032", "2130021", "0322331", "0321312", "3001210", "2113202", "0121303", "1332210", "3130022", "2023132", "0133312", "2023201", "3223012", "3301121", "2232301", "2113000", "1112320", "2031221", "3203123", "2120203", "3210003", "0013121", "3120213", "1002331", "3231010", "3230201", "2301321", "1330212", "2103102", "2320012", "0210123", "3111200", "1231300", "3213012", "2310213", "2330331", "2310122", "0322103", "1020332", "3130233", "2013021", "2330210", "3013023", "0321011", "3201023", "1003230", "2003311", "3103322", "3021210", "0123331", "0312031", "3211031", "3122230", "1200213", "2103233", "0310120", "0001233", "3122002", "0313210", "1230112", "2313100", "3103121", "3102100", "2321330", "1301202", "1310232", "0003212", "1023100", "0102332", "2330121", "1203130", "2123110", "3032111", "2321200", "3310201", "2211303", "2030132", "1012033", "3230301", "1233203", "1313203", "0103323", "2110223", "0312321", "1313032", "1113200", "1302000", "0210302", "3010220", "0033201", "3003120", "2321202", "1000332", "2013231", "1313023", "0013230", "0222321", "2331032", "0031221", "3222103", "1330021", "2213021", "3320311", "2203310", "3311201", "2021223", "1021232", "2202133", "1030020", "2130200", "2310312", "0113213", "3221023", "2300102", "0232213", "3232100", "0322221", "0010230", "2031321", "1200032", "2320131", "0213100", "3112320", "3200103", "3021203", "0312111", "2120032", "1231220", "1230120", "1011213", "3231101", "3103202", "1123310", "2121330", "2103323", "2102033", "3202122", "3301223", "3221120", "1120331", "0011322", "3013233", "1121330", "0300021", "2032210", "2031230", "3011232", "0013332", "0003123", "2230120", "0301211", "2022013", "2103010", "1311102", "2202113", "3013122", "2203211", "0032012", "0111321", "2123202", "3303112", "0212312", "1100203", "3222010", "2110330", "3230111", "0310201", "0331132", "2032201", "1200333", "3133230", "2311202", "3210231", "1001230", "1023122", "3101200", "2031210", "3300213", "1023101", "2033313", "0232212", "3200311", "3330221", "0322321", "2320311", "3320010", "1312203", "2010331", "2223102", "1221030", "3320130", "0032001", "1300232", "0121302", "1021223", "2231203", "0201003", "2020331", "0312202", "3101322", "2032213", "3112301", "0221133", "1231021", "1333201", "0011032", "3102321", "3213301", "1003231", "1021113", "2302301", "3020331", "2022301", "2003221", "0202113", "3211202", "1221310", "2010032", "1033320", "0121233", "0010123", "1122330", "1001302", "2021330", "3102130", "2132023", "2003001", "1330203", "2333201", "1210030", "0322301", "2021123", "3031032", "1312023", "3021213", "0130122", "1013200", "1023021", "0312030", "2033122", "2312220", "3232101", "2231101", "1300021", "2110233", "1003320", "3211011", "2132230", "0311320", "2301222", "2113023", "3221210", "0131102", "0231302", "2021320", "0312020", "1003123", "0300102", "1332013", "1203221", "0202130", "0102322", "0203120", "3230132", "3012303", "3203133", "1203303", "3310121", "1331320", "2133000", "2121030", "2312022", "3002312", "1030332", "3010321", "3012000", "3203331", "2011123", "2223101", "2132030", "1322200", "0233131", "1021103", "0120223", "3233120", "1002203", "2101013", "3003122", "3020231", "3302120", "0332012", "1313302", "0203201", "1103233", "0301223", "3321330", "2330001", "0121133", "2130210", "0212330", "3032132", "3102020", "2013103", "2120331", "0133012", "3301231", "2131230", "1213022", "2100323", "0031203", "0113203", "3021120", "1003232", "2013303", "2230212", "1200023", "3203011", "0311201", "3232013", "2003130", "3123012", "0202133", "3102222", "3102200", "2001230", "2212033", "1132110", "1220003", "1330231", "1020032", "2312011", "1230103", "2231011", "1231200", "2102300", "2001032", "2032301", "2120312", "0132333", "0102132", "1300122", "0312222", "2133031", "0322112", "3013002", "3002112", "1102302", "0130121", "2231110", "3213032", "1031020", "3321013", "3021003", "3110312", "2321110", "3021100", "1313102", "1031102", "3201311", "0220321", "3220113", "1211300", "3101211", "2202231", "0132030", "2012203", "0010321", "0233011", "0233111", "1211203", "3022011", "3323301", "3120030", "1030123", "3010122", "0211133", "3003210", "2323013", "3233210", "3220132", "1222303", "2123011", "1310020", "3311210", "1030112", "0331211", "2013233", "2202310", "0122123", "2310300", "0231222", "1300022", "0323112", "3230310", "2132032", "1220330", "2311020", "3233101", "3201130", "0022301", "1021233", "3302133", "2232102", "2213100", "0311233", "1132210", "2323310", "3031002", "1013233", "2012033", "2123101", "3032123", "2200123", "3021113", "0013202", "2003021", "2030011", "3323101", "3300021", "3122012", "0123000", "0001232", "3120001", "1300112", "1320031", "0221113", "1021303", "3100210", "1323320", "2031022", "2203103", "1201231", "2000311", "3200013", "0313200", "2010310", "2031232", "1220132", "3130002", "3302100", "1102032", "3001320", "0311203", "1013102", "3212302", "0210213", "2200132", "0110203", "3110023", "1032100", "2220031", "1103203", "2313012", "2231003", "0230130", "1012013", "0320120", "2010233", "1033220", "1022030", "2033013", "1220033", "3200011", "0102232", "2310221", "0113223", "2132320", "0123031", "0031231", "0102310", "1223010", "1322202", "0002313", "1000320", "1133120", "3023101", "3110203", "1130302", "3113021", "1020331", "3132210", "3112310", "1233200", "2202031", "2210213", "3012033", "3211302", "0222313", "3123320", "0131123", "0223311", "3212110", "0321233", "3120013", "2100310", "3031312", "3320100", "2030101", "1032103", "3102203", "0312220", "2023212", "3113032", "1213202", "2333001", "1200303", "2102213", "3312203", "2221031", "3320210", "0120330", "2300131", "1122301", "3212033", "2003103", "1312301", "1120223", "2133022", "2002123", "1011332", "1322303", "1300120", "1023223", "0320112", "1333020", "0103302", "2303110", "1103201", "0212213", "3232021", "3103320", "0101023", "2321012", "1312200", "2313103", "1202213", "0113302", "2110302", "3111202", "3131012", "0213330", "3210202", "2022213", "1100023", "1023301", "1112103", "3110032", "2010323", "2333310", "2201203", "2113002", "0023100", "3301112", "2120322", "3221203", "0022231", "3212003", "2131001", "3200213", "1203232", "0231131", "2133013", "2312120", "1130233", "2010300", "2031121", "2310211", "0322311", "3311022", "0133023", "3232001", "0233313", "0221320", "2020123", "3301323", "2331110", "1203233", "2013033", "1013120", "1021302", "3102122", "2301112", "0302021", "2313200", "0212132", "1122230", "1130230", "1302003", "1021023", "0203210", "2303101", "1123102", "2013032", "0232210", "2130030", "1202312", "0312021", "1032031", "2313013", "2003301", "0210322", "0012103", "0201223", "0321212", "3313023", "3023112", "3212000", "3102002", "0102203", "0210312", "3102201", "2230112", "1231230", "2112013", "1201130", "1213200", "3121220", "3103210", "2110323", "1030202", "3013211", "0132302", "3002133", "3030120", "2012232", "2213003", "0012322", "0133323", "1300230", "3011231", "3321011", "0322120", "0300321", "2130232", "1120312", "0323321", "1032102", "2302021", "0031132", "2023011", "2231012", "3011312", "1232302", "3013220", "1012321", "2103100", "1203012", "0100203", "3021332", "3110321", "3323100", "0201031", "0131012", "3013123", "1223230", "0020133", "1233101", "2331100", "3021322", "2230313", "3113020", "1103323", "1333230", "1023313", "3101020", "0332123", "3121110", "1333320", "0210131", "1223032", "2302123", "3323013", "0231133", "0321023", "3032100", "2320102", "0032321", "2103130", "0132200", "0310121", "1332220", "1113002", "3021200", "2123130", "3012133", "1230000", "3212310", "1032300", "2331023", "0210023", "2033031", "3213000", "3011200", "2100330", "2321301", "1313002", "3321033", "3302011", "3122300", "2021311", "1022322", "2031000", "2130013", "0121213", "2330133", "3213202", "0023010", "1230312", "2133003", "3223011", "1123010", "0111332", "2130231", "0313302", "2301311", "3113023", "0123211", "0321203", "2131300", "2020310", "0233122", "2130321", "2123201", "1020233", "1121302", "3222001", "2323001", "2110231", "1320133", "1123300", "3103212", "1302021", "0013112", "1122031", "3022311", "2232310", "1200232", "1003102", "2221013", "2300213", "0310102", "3201031", "3100023", "3210333", "3103223", "0011203", "0100233", "1131032", "1201302", "3312320", "3032112", "1323301", "0222231", "0023123", "0003221", "0310032", "0332132", "1212310", "0232221", "0123220", "0211333", "1312110", "3200110", "0302111", "3201013", "1321203", "0303012", "2230133", "1011322", "2203213", "3100200", "0100213", "3202011", "0301132", "3011123", "3031232", "3023212", "0231321", "0113122", "0210231", "2301000", "3121031", "1223110", "1321011", "0020131", "0211322", "2021312", "1211230", "2023310", "3322210", "3131022", "2213110", "1121003", "1323202", "1001032", "2001231", "0012130", "2320110", "0112013", "3321001", "1121310", "3212300", "1012312", "1320103", "0130201", "3323210", "1303012", "2031312", "2032121", "1003222", "3031323", "2033001", "1311120", "3301022", "3120000", "3103032", "3222031", "1111023", "1330323", "0231101", "1211031", "2303311", "3120231", "1213010", "0131233", "0103220", "3032103", "1302110", "2111320", "0233211", "0233021", "1333032", "3103120", "3210033", "3002311", "3220231", "3132030", "2103013", "1320030", "0203311", "1103302", "2103133", "0123013", "1202130", "1122300", "0201311", "2021323", "0102003", "3013012", "2312200", "0210331", "2201303", "0213301", "0222031", "3201020", "3123120", "1332010", "0133201", "2103022", "3112103", "1233100", "3032313", "2000310", "3223031", "0220312", "1031232", "0030112", "0311231", "2031133", "1033321", "3033123", "1213330", "2201030", "3120121", "0032013", "3112010", "3110120", "3013232", "2301312", "1033021", "2130032", "3111210", "2013121", "3012321", "0230321", "1332310", "0003201", "1231203", "1101032", "0321113", "0013322", "0132320", "3111002", "1110302", "3022001", "0133322", "2030212", "2201331", "0223013", "2023113", "0002133", "2321203", "2130112", "2130000", "1303211", "0121310", "3102320", "1311320", "2101331", "2000013", "1230210", "2023012", "1130212", "0321313", "3120130", "0001332", "0013223", "2220131", "2020130", "3131032", "2311301", "0031332", "0021032", "3032211", "2232130", "1230111", "2213202", "2200321", "2002312", "1201330", "0130123", "1021030", "0121123", "2013222", "1120130", "2220132", "3010223", "1132310", "1320331", "3031231", "0113222", "1102123", "2113020", "0123113", "2131200", "3202331", "1031112", "0132113", "3011023", "0101312", "0211031", "3013112", "3113201", "3310020", "0032111", "0132132", "0111302", "2131012", "0311020", "1210131", "1121203", "1300200", "3331230", "0321102", "0220131", "0232011", "1230130", "2231320", "0231223", "2301122", "1132031", "1303202", "1020330", "0212231", "1211003", "0203231", "3202010", "1320013", "1321012", "2201032", "0030012", "3010020", "2131103", "0311032", "1302301", "0132022", "1022131", "3300201", "2123020", "1202331", "1130232", "2103222", "1222320", "1203120", "1102030", "0032210", "1222310", "0001322", "2103131", "3302031", "3010012", "2120310", "3110202", "2020313", "2332031", "1221301", "2103103", "1132330", "2312023", "1031132", "2303213", "3210131", "1103210", "3131023", "2330031", "2021031", "1031202", "3001222", "2031310", "3222130", "1132023", "0233112", "3120020", "2030210", "1112032", "2311220", "3023021", "2303331", "2312230", "2011031", "1023130", "3100132", "1302313", "1012003", "0201013", "1033223", "1320122", "2111031", "3201233", "0323201", "1321120", "3032133", "1023022", "1121130", "1330223", "0303201", "0023313", "0022013", "2132302", "0023311", "0321330", "0211131", "2223012", "2011023", "2300011", "2133010", "3112020", "1022232", "2031333", "1213230", "2123003", "3133201", "0233120", "0021123", "3002221", "0210300", "0331230", "2030133", "0123102", "2221330", "1300102", "2103023", "2103121", "2233011", "2332201", "2023301", "1320321", "0223122", "2301033", "2113030", "1332002", "0320103", "3101230", "3111021", "1201230", "3310320", "2013331", "3222101", "3210300", "3332103", "2210230", "0013023", "0213201", "3122120", "3210200", "0103211", "3213101", "0120133", "3012022", "0203301", "1030002", "1311022", "2332130", "0213233", "0213103", "1132033", "0211033", "0303312", "1102231", "3300221", "2102303", "1202232", "2303120", "2310323", "3120322", "1110321", "1203323", "0203310", "1231303", "1020320", "2330231", "1123201", "1133302", "3231202", "3033312", "0201312", "1323201", "0231011", "0323231", "0312013", "3210121", "2111230", "3120303", "0031202", "3202231", "2101030", "0213323", "1002310", "3213330", "1200302", "2230011", "3130332", "0200113", "1012032", "2002213", "2103030", "3200321", "3201120", "0302110", "0210130", "3020011", "0112132", "2123320", "0333201", "1333002", "2321230", "2013211", "2102312", "0000231", "0200130", "2311302", "2103302", "2233013", "0311230", "2322110", "2031103", "3210210", "3123022", "0220132", "2113101", "1003302", "1020213", "0033112", "1230333", "3010320", "0122230", "3012313", "2223011", "2220313", "0302102", "3031132", "3233110", "3301201", "0012231", "1032233", "2002331", "3032122", "1133320", "2100231", "0202013", "3112031", "3131200", "0211302", "3302013", "2022312", "3013312", "2011103", "1013202", "2023133", "3002310", "1212230", "2013010", "3002103", "3002211", "0302231", "2033133", "3123102", "2102320", "1220310", "0001230", "2301013", "1013023", "2211103", "1322203", "3202312", "3313210", "3210222", "3101302", "2013011", "1221203", "1231100", "0211331", "1132001", "2203201", "0223101", "1320012", "1202123", "0300212", "2333120"}
}
