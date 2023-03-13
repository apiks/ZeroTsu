package entities

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/sasha-s/go-deadlock"
)

var (
	Mutex         deadlock.RWMutex
	SharedInfo    *sharedInfo
	AnimeSchedule = &AnimeScheduleMap{AnimeSchedule: make(map[int][]*ShowAirTime)}
)

type AnimeScheduleMap struct {
	deadlock.RWMutex
	AnimeSchedule map[int][]*ShowAirTime
}

// LoadSharedDB loads global shared DBs in mem
func LoadSharedDB() {
	// Creates missing "database" and "shared" folder if they don't exist
	if _, err := os.Stat("database"); os.IsNotExist(err) {
		err := os.Mkdir("database", 0777)
		if err != nil {
			log.Println(err)
			return
		}
	}
	if _, err := os.Stat("database/shared"); os.IsNotExist(err) {
		err := os.Mkdir("database/shared", 0777)
		if err != nil {
			log.Println(err)
			return
		}
	}

	files, err := IOReadDir("database/shared")
	if err != nil {
		log.Panicln(err)
	}

	SharedInfo.Lock()
	defer SharedInfo.Unlock()
	SharedInfo = newSharedInfo(make(map[string]*RemindMeSlice), make(map[string][]*ShowSub))
	for _, file := range files {
		LoadSharedDBFile(file)
	}
}

func LoadSharedDBFile(file string) {
	// Reads all the info from the file and puts them in infoByte as bytes
	infoByte, err := ioutil.ReadFile(fmt.Sprintf("database/shared/%s", file))
	if err != nil {
		log.Println(err)
		return
	}

	// Takes the data and puts it into the appropriate field
	switch file {
	case "remindMes.json":
		err = json.Unmarshal(infoByte, &SharedInfo.RemindMes)
		if err != nil {
			log.Println("LoadSharedDBFile remindMes error:", err)
			return
		}
	case "animeSubs.json":
		err = json.Unmarshal(infoByte, &SharedInfo.AnimeSubs)
		if err != nil {
			log.Println("LoadSharedDBFile animeSubs error:", err)
			return
		}
	}
}

// RemindMeWrite writes RemindMes to remindMes.json
func RemindMeWrite(remindMe map[string]*RemindMeSlice) error {
	// Checks if the user has hit the db limit
	for _, remindMeSlice := range remindMe {
		if remindMeSlice == nil {
			continue
		}

		if remindMeSlice.GetPremium() && len(remindMeSlice.GetRemindMeSlice()) > 299 {
			return fmt.Errorf("Error: You have reached the RemindMe limit (300) for this premium account.")
		} else if !remindMeSlice.GetPremium() && len(remindMeSlice.GetRemindMeSlice()) > 49 {
			return fmt.Errorf("Error: You have reached the RemindMe limit (50) for this account. Please remove some or increase it to 300 by upgrading to a premium user at <https://patreon.com/animeschedule>")
		}
	}

	// Turns that slice into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(remindMe, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	SharedInfo.Lock()
	defer SharedInfo.Unlock()
	err = ioutil.WriteFile("database/shared/remindMes.json", marshaledStruct, 0644)
	if err != nil {
		log.Println("RemindMeWrite error:", err)
		return err
	}

	return nil
}

// Writes anime notfication subscription to animeSubs.json
func AnimeSubsWrite(animeSubs map[string][]*ShowSub) error {
	// Turns that slice into bytes to be ready to written to file
	marshaledStruct, err := json.MarshalIndent(animeSubs, "", "    ")
	if err != nil {
		return err
	}

	// Writes to file
	SharedInfo.Lock()
	defer SharedInfo.Unlock()
	err = ioutil.WriteFile("database/shared/animeSubs.json", marshaledStruct, 0644)
	if err != nil {
		log.Println("AnimeSubsWrite error:", err)
		return err
	}

	return nil
}

// Reads and returns the names of every file in that directory
func IOReadDir(root string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}

func SetupGuildSub(guildID string) {
	var (
		shows      []*ShowSub
		now        = time.Now().UTC()
		addedShows = make(map[string]bool)
	)

	// Adds every single non-duplicate show as a guild subscription
	AnimeSchedule.RLock()
	for dayInt, scheduleShows := range AnimeSchedule.AnimeSchedule {
		if scheduleShows == nil {
			continue
		}

		for _, show := range scheduleShows {
			if show == nil {
				continue
			}
			if _, ok := addedShows[show.GetKey()]; ok {
				continue
			}

			// Checks if the show is from today and whether it has already passed (to avoid notifying the user today if it has passed)
			var hasAiredToday bool
			if int(now.Weekday()) == dayInt {

				// Reset bool
				hasAiredToday = false

				// Parse the air hour and minute
				t, err := time.Parse("3:04 PM", show.GetAirTime())
				if err != nil {
					log.Println(err)
					continue
				}

				// Form the air date for today
				scheduleDate := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), now.Second(), now.Nanosecond(), now.Location())
				scheduleDate = scheduleDate.UTC()

				// Calculates whether the show has already aired today
				difference := now.Sub(scheduleDate.UTC())
				if difference >= 0 {
					hasAiredToday = true
				}
			}

			guildSub := NewShowSub(show.GetName(), false, true)
			if hasAiredToday {
				guildSub.SetNotified(true)
			}

			shows = append(shows, guildSub)
			addedShows[show.GetKey()] = true
		}
	}
	AnimeSchedule.RUnlock()

	SharedInfo.Lock()
	SharedInfo.AnimeSubs[guildID] = shows
	SharedInfo.Unlock()
}

func InTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}
