package data

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/rumyantseva/highloadcup/pkg/cache"
	"github.com/rumyantseva/highloadcup/pkg/db"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

type Storage struct {
	DB *db.WithMax

	User     *cache.Storage
	Location *cache.Storage
	Visit    *cache.Storage
}

func NewStorage(
	withdb *db.WithMax, user *cache.Storage, location *cache.Storage, visit *cache.Storage,
) *Storage {
	return &Storage{
		DB:       withdb,
		User:     user,
		Location: location,
		Visit:    visit,
	}
}

// Import prepared data from data files.
func (s *Storage) Import() error {
	archive := "/tmp/data/data.zip"

	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	var userFiles []*zip.File
	var locationFiles []*zip.File
	var visitFiles []*zip.File

	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "users") {
			userFiles = append(userFiles, file)
		} else if strings.HasPrefix(file.Name, "locations") {
			locationFiles = append(locationFiles, file)
		} else if strings.HasPrefix(file.Name, "visits") {
			visitFiles = append(visitFiles, file)
		}
	}

	wg := &sync.WaitGroup{}

	for _, file := range userFiles {
		wg.Add(1)
		go func(file *zip.File) {
			log.Printf("Process file %s...", file.Name)
			err = s.user(file)
			if err != nil {
				log.Printf("Couldn't parse user: %v", err)
			}
			log.Printf("File %s processed.", file.Name)
			wg.Done()
		}(file)
	}

	for _, file := range locationFiles {
		wg.Add(1)
		go func(file *zip.File) {
			log.Printf("Process file %s...", file.Name)
			err = s.location(file)
			if err != nil {
				log.Printf("Couldn't parse location: %v", err)
			}
			log.Printf("File %s processed.", file.Name)
			wg.Done()
		}(file)
	}

	for _, file := range visitFiles {
		wg.Add(1)
		go func(file *zip.File) {
			log.Printf("Process file %s...", file.Name)
			err = s.visit(file)
			if err != nil {
				log.Printf("Couldn't parse visit: %v", err)
			}
			log.Printf("File %s processed.", file.Name)
			wg.Done()
		}(file)
	}

	wg.Wait()

	log.Printf("Processed %d user files", len(userFiles))
	log.Printf("Processed %d location files", len(locationFiles))
	log.Printf("Processed %d visit files", len(visitFiles))

	return nil
}

func LocalTime() (int, error) {
	log.Print("Import options...")
	file, err := os.Open("/tmp/data/options.txt")
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

func (s *Storage) user(file *zip.File) error {
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

	txn := s.DB.DB.Txn(true)
	for _, user := range data.Users {
		if err := txn.Insert("user", user); err != nil {
			return err
		}

		go s.User.SetFrom(fmt.Sprint(user.ID), user)

		/*withdb.MxUser.Lock()
		if user.ID > withdb.MaxUser {
			withdb.MaxUser = user.ID
		}
		withdb.MxUser.Unlock()*/
	}
	txn.Commit()

	return nil
}

func (s *Storage) location(file *zip.File) error {
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

	txn := s.DB.DB.Txn(true)
	for _, loc := range data.Locations {
		if err := txn.Insert("location", loc); err != nil {
			return err
		}

		go s.Location.SetFrom(fmt.Sprint(loc.ID), loc)

		/*withdb.MxLocation.Lock()
		if loc.ID > withdb.MaxLocation {
			withdb.MaxLocation = loc.ID
		}
		withdb.MxLocation.Unlock()*/
	}
	txn.Commit()

	return nil
}

func (s *Storage) visit(file *zip.File) error {
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

	txn := s.DB.DB.Txn(true)
	for _, visit := range data.Visits {
		if err := txn.Insert("visit", visit); err != nil {
			return err
		}

		go s.Visit.SetFrom(fmt.Sprint(visit.ID), visit)

		/*withdb.MxVisit.Lock()
		if visit.ID > withdb.MaxVisit {
			withdb.MaxVisit = visit.ID
		}
		withdb.MxVisit.Unlock()*/
	}
	txn.Commit()

	return nil
}
