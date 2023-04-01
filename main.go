package main

import (
  "os"
  "fmt"
  "bufio"
  "strings"
  "strconv"
  "time"
  "io/ioutil"
)

const TimeFormat = "01-02-2006"


type DailyGames struct {
  wins int
  losses int
  draws int
  timeControl string
	playedOn time.Time
}

// global state
var playedGames = make([]*DailyGames, 0)
var currentDataFile *os.File
var currentFilePathFull string

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Can't find current user information. " + err.Error())
		os.Exit(-1)
	}

	dataFolder := homeDir + "/cl"
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		if err := os.Mkdir(dataFolder, os.ModePerm); err != nil {
			fmt.Println("Error creating " + dataFolder + " :: " + err.Error())
			os.Exit(-1)
		}
	}

  currentFilePathFull = dataFolder + "/current"
  currentDataFile = getFile(currentFilePathFull)

  loadGamesFile()
  showMenu()

	for {
		cmd := readCmd()
    if strings.HasPrefix(cmd, "q") {
      _ = currentDataFile.Close()
      fmt.Printf("Keep practicing chess daily.\nBye\n") 
      os.Exit(0)
    } else if strings.HasPrefix(cmd, "record ") {
      recParts := strings.Split(cmd, "record ")
      saveGames(recParts[1])
    } else if strings.HasPrefix(cmd, "show ") {
      showParts := strings.Split(cmd, "show ")
      showGameStats(showParts[1])
    } else {
      fmt.Println("I have no idea what you are talking about.")
    }
	}

}

func showMenu() {
  fmt.Println("\nWelcome to keeping track of daily chess games.\nThis program supports only three commands.")
  fmt.Println("1. record <blitz/rapid/classical> W-L-D -> records game results for a time control / type")
  fmt.Println("2. show now-7d -> shows game statistics for a time duration that is expressed as now-Nd, where N is the last N number of days")
  fmt.Println("3. quit -> quits the program")
}

func readCmd() string {
	fmt.Printf("\n> ")
	reader := bufio.NewReader(os.Stdin)
	cmd, e := reader.ReadString('\n')
	if e != nil {
		fmt.Println("Error in reading command " + e.Error())
		os.Exit(-1)
	}
	return strings.Replace(cmd, "\n", "", -1)
}

func saveGames(c string) {
  cmdParts := strings.Split(c, " ");
  gameType := cmdParts[0]
  if gameType == "blitz" || gameType == "classical" || gameType == "rapid" {
    w, l, d := parseScore(cmdParts[1])
    playedGames = append(playedGames, &DailyGames{
      wins : w,
      losses: l,
      draws : d,
      timeControl: gameType,
      playedOn:  time.Now(),
    })
  } else {
    panic("Invalid gameType " + gameType)
  }
  saveGamesFile()
  fmt.Println("Saved.")
}

func parseScore(s string) (int, int, int) {
  scoreParts := strings.Split(s, "-")

  wins, err := strconv.Atoi(scoreParts[0])
  if err != nil {
    fmt.Printf("Invalid Number %s :: %s", scoreParts[0], err.Error())
    return -1, -1, -1
  }

  losses, err := strconv.Atoi(scoreParts[1])
  if err != nil {
    fmt.Printf("Invalid Number %s :: %s", scoreParts[1], err.Error())
    return -1, -1, -1
  }

  draws, err := strconv.Atoi(scoreParts[2])
  if err != nil {
    fmt.Printf("Invalid Number %s :: %s", scoreParts[2], err.Error())
    return -1, -1, -1
  }
  
  return wins, losses, draws
}

func showGameStats(nowStr string) {
  if !strings.HasPrefix(nowStr, "now-") {
    fmt.Println("Invalid time expression " + nowStr)
    return
  }
  durationStr := strings.Replace(nowStr, "now-", "", 1)
  if !strings.HasSuffix(durationStr, "d") {
    fmt.Println("Invalid duration expression " + durationStr)
    return
  }
  daysStr := strings.Replace(durationStr, "d", "", 1)
  daysNum, err := strconv.Atoi(daysStr)
  if err != nil {
    fmt.Println("Invalid number format " + daysStr)
    return
  }
  d := time.Duration(-daysNum) * 24 * time.Hour
  threshold := time.Now().Add(d)
  totalGamesPlayed := 0
  totalWins := 0
  totalLosses := 0
  totalDraws := 0

  fmt.Println("-----------------------------------------------")
	for _, dg := range playedGames {
		if dg.playedOn.After(threshold) {
			fmt.Printf("%s\t(%v) %v-%v-%v\n", dg.playedOn.Format(TimeFormat), dg.timeControl, dg.wins, dg.losses, dg.draws)
      totalGamesPlayed += dg.wins + dg.losses + dg.draws
      totalWins += dg.wins
      totalLosses += dg.losses
      totalDraws += dg.draws
		}
  }

  fmt.Printf("\nTotal games played in %v :: %v\n", nowStr, totalGamesPlayed)
  fmt.Printf("Wins %v \n", totalWins)
  fmt.Printf("Losses %v \n", totalLosses)
  fmt.Printf("Draws %v \n", totalDraws)
  fmt.Println("-----------------------------------------------")
}

func getFile(fileName string) *os.File {
	filePtr, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0664)
	if err != nil {
		fmt.Printf("Unable to open %s :: %s", fileName, err.Error())
		os.Exit(-1)
	}
	return filePtr
}

func saveGamesFile() {
	var lines, recordedGameLine string
	for _, dg := range playedGames {
	  recordedGameLine = fmt.Sprintf("%v %v-%v-%v\t%v\n", dg.timeControl, dg.wins, dg.losses, dg.draws, dg.playedOn.Format(TimeFormat))
		lines = lines + recordedGameLine
	}

	if err := ioutil.WriteFile(currentFilePathFull, []byte(lines), 0644); err != nil {
		fmt.Println("Error writing file system for current tasks file " + err.Error())
		os.Exit(-1)
	}
}

func loadGamesFile() {
	fileScanner := bufio.NewScanner(currentDataFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		lineParts := strings.Split(line, "\t")
		gamesPlayedOn, err := time.Parse(TimeFormat, lineParts[1])
		if err != nil {
			fmt.Println("Oops, error loading done tasks file " + err.Error())
			os.Exit(-1)
		}
    gameParts := strings.Split(lineParts[0], " ")
    w, l, d := parseScore(gameParts[1])

    playedGames = append(playedGames, &DailyGames{
      wins : w,
      losses: l,
      draws : d,
      timeControl: gameParts[0],
      playedOn:  gamesPlayedOn,
    })
	}
	fmt.Printf("Loading recorded games:: %d\n", len(playedGames))
}

