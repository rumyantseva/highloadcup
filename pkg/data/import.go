package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	memdb "github.com/hashicorp/go-memdb"
	"github.com/rumyantseva/highloadcup/pkg/models"
)

// Import prepared data from data files.
func Import(db *memdb.MemDB) (int, error) {
	archive := "/tmp/data/data.zip"
	target := "/tmp/data/unzip"

	// try 5 times to open zip
	var err error
	for i := 0; i < 5; i++ {
		err = unzip(archive, target)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return 0, err
	}

	max := 1000000
	pattern := target + "/%s_%d.json"

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		var i int
		for i = 1; i < max; i++ {
			file, err := os.Open(fmt.Sprintf(pattern, "users", i))
			if err != nil {
				break
			}

			err = user(file, db)
			if err != nil {
				log.Printf("Couldn't parse user: %v", err)
			}
		}
		log.Printf("Processed %d user files", i-1)
		wg.Done()
	}()

	go func() {
		var i int
		for i = 1; i < max; i++ {
			file, err := os.Open(fmt.Sprintf(pattern, "locations", i))
			if err != nil {
				break
			}

			err = location(file, db)
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

			err = visit(file, db)
			if err != nil {
				log.Printf("Couldn't parse user: %v", err)
			}
		}
		log.Printf("Processed %d visits files", i-1)
		wg.Done()
	}()

	wg.Wait()

	log.Print("Import options...")
	file, err := os.Open(target + "/options.txt")
	if err != nil {
		return 0, err
	}

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

func user(file *os.File, db *memdb.MemDB) error {
	defer file.Close()

	data := struct {
		Users []models.User `json:"users"`
	}{}
	err := json.NewDecoder(file).Decode(&data)
	if err != nil {
		return fmt.Errorf("Couldn't parse user file. %v", err)
	}

	txn := db.Txn(true)
	for _, user := range data.Users {
		if err := txn.Insert("user", user); err != nil {
			return err
		}
	}
	txn.Commit()

	return nil
}

func location(file *os.File, db *memdb.MemDB) error {
	defer file.Close()

	data := struct {
		Locations []models.Location `json:"locations"`
	}{}
	err := json.NewDecoder(file).Decode(&data)
	if err != nil {
		return fmt.Errorf("Couldn't parse locations file. %v", err)
	}

	txn := db.Txn(true)
	for _, loc := range data.Locations {
		if err := txn.Insert("location", loc); err != nil {
			return err
		}
	}
	txn.Commit()

	return nil
}

func visit(file *os.File, db *memdb.MemDB) error {
	defer file.Close()

	data := struct {
		Visits []models.Visit `json:"visits"`
	}{}
	err := json.NewDecoder(file).Decode(&data)
	if err != nil {
		return fmt.Errorf("Couldn't parse visits file. %v", err)
	}

	txn := db.Txn(true)
	for _, visit := range data.Visits {
		if err := txn.Insert("visit", visit); err != nil {
			return err
		}
	}
	txn.Commit()

	return nil
}
