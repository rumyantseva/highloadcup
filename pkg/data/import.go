package data

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

// Import prepared data from data files.
func Import(withdb *db.WithMax) (int, error) {
	archive := "/tmp/data/data.zip"
	//target := "/tmp/data/unzip"

	// try 5 times to open zip
	/*var err error
	for i := 0; i < 5; i++ {
		err = unzip(archive, target)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return 0, err
	}*/

	reader, err := zip.OpenReader(archive)
	if err != nil {
		return 0, err
	}

	var userFiles []*zip.File
	var locationFiles []*zip.File
	var visitFiles []*zip.File
	var optionsFile *zip.File
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "users") {
			userFiles = append(userFiles, file)
		} else if strings.HasPrefix(file.Name, "locations") {
			locationFiles = append(locationFiles, file)
		} else if strings.HasPrefix(file.Name, "visits") {
			visitFiles = append(visitFiles, file)
		} else if strings.HasPrefix(file.Name, "options") {
			optionsFile = file
		}
	}

	//max := 1000000
	//pattern := target + "/%s_%d.json"

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		for _, file := range userFiles {
			err = user(file, withdb)
			if err != nil {
				log.Printf("Couldn't parse user: %v", err)
			}
		}
		log.Printf("Processed %d user files", len(userFiles))
		wg.Done()
	}()

	go func() {
		for _, file := range locationFiles {
			err = location(file, withdb)
			if err != nil {
				log.Printf("Couldn't parse location: %v", err)
			}
		}
		log.Printf("Processed %d location files", len(locationFiles))
		wg.Done()
	}()

	go func() {
		for _, file := range visitFiles {
			err = visit(file, withdb)
			if err != nil {
				log.Printf("Couldn't parse visit: %v", err)
			}
		}
		log.Printf("Processed %d visit files", len(visitFiles))
		wg.Done()
	}()

	/*	go func() {
			var i int
			for i = 1; i < max; i++ {
				file, err := os.Open(fmt.Sprintf(pattern, "locations", i))
				if err != nil {
					break
				}

				err = location(file, withdb)
				if err != nil {
					log.Printf("Couldn't parse location: %v", err)
				}
			}
			log.Printf("Processed %d location files", i-1)
			wg.Done()
		}()

		go func() {
			var i int
			for i = 1; i < max; i++ {
				file, err := os.Open(fmt.Sprintf(pattern, "visits", i))
				if err != nil {
					break
				}

				err = visit(file, withdb)
				if err != nil {
					log.Printf("Couldn't parse user: %v", err)
				}
			}
			log.Printf("Processed %d visits files", i-1)
			wg.Done()
		}()
	*/
	wg.Wait()

	log.Print("Import options...")
	if optionsFile == nil {
		log.Print("There is no options file here!")
		return int(time.Now().Unix()), nil
	}

	file, err := optionsFile.Open()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	line, _, err := bufio.NewReader(file).ReadLine()
	if err != nil {
		return 0, err
	}

	st, err := strconv.Atoi(string(line))
	if err != nil {
		return 0, err
	}

	return st, nil
}

func user(file *zip.File, withdb *db.WithMax) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data := struct {
		Users []models.User `json:"users"`
	}{}
	err = json.NewDecoder(rc).Decode(&data)
	if err != nil {
		return fmt.Errorf("Couldn't parse user file. %v", err)
	}

	txn := withdb.DB.Txn(true)
	for _, user := range data.Users {
		if err := txn.Insert("user", user); err != nil {
			return err
		}

		/*withdb.MxUser.Lock()
		if user.ID > withdb.MaxUser {
			withdb.MaxUser = user.ID
		}
		withdb.MxUser.Unlock()*/
	}
	txn.Commit()

	return nil
}

func location(file *zip.File, withdb *db.WithMax) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data := struct {
		Locations []models.Location `json:"locations"`
	}{}
	err = json.NewDecoder(rc).Decode(&data)
	if err != nil {
		return fmt.Errorf("Couldn't parse locations file. %v", err)
	}

	txn := withdb.DB.Txn(true)
	for _, loc := range data.Locations {
		if err := txn.Insert("location", loc); err != nil {
			return err
		}

		/*withdb.MxLocation.Lock()
		if loc.ID > withdb.MaxLocation {
			withdb.MaxLocation = loc.ID
		}
		withdb.MxLocation.Unlock()*/
	}
	txn.Commit()

	return nil
}

func visit(file *zip.File, withdb *db.WithMax) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	data := struct {
		Visits []models.Visit `json:"visits"`
	}{}
	err = json.NewDecoder(rc).Decode(&data)
	if err != nil {
		return fmt.Errorf("Couldn't parse visits file. %v", err)
	}

	txn := withdb.DB.Txn(true)
	for _, visit := range data.Visits {
		if err := txn.Insert("visit", visit); err != nil {
			return err
		}

		/*withdb.MxVisit.Lock()
		if visit.ID > withdb.MaxVisit {
			withdb.MaxVisit = visit.ID
		}
		withdb.MxVisit.Unlock()*/
	}
	txn.Commit()

	return nil
}
