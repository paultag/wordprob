package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli"
)

const SCALE_FACTOR = 14.3

func main() {
	app := cli.NewApp()
	app.Name = "wordprob"
	app.Usage = ""
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "worddb",
			Value: "",
			Usage: "",
		},
	}

	app.Commands = []cli.Command{
		CompileCommand,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

type DB struct {
	db *leveldb.DB
}

func (d DB) Save(word string, value float64) error {
	return d.db.Put([]byte(word), []byte(fmt.Sprintf("%f", value)), nil)
}

func (d DB) Get(word string) float64 {
	data, err := d.db.Get([]byte(word), nil)
	if err != nil {
		return -SCALE_FACTOR
	}
	number, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return -SCALE_FACTOR
	}
	return number
}

func LoadDB(c *cli.Context) (*DB, error) {
	path := c.GlobalString("worddb")

	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func CalculateFreqWeight(value float64) float64 {
	return math.Log(value) - SCALE_FACTOR
}

func Compile(c *cli.Context) error {
	db, err := LoadDB(c)
	if err != nil {
		return err
	}

	fd, err := os.Open(c.String("wordfreq"))
	if err != nil {
		return err
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		entry := strings.SplitN(strings.TrimSpace(scanner.Text()), "\t", 2)

		word := strings.ToUpper(entry[0])
		count, err := strconv.Atoi(entry[1])
		if err != nil {
			return err
		}

		prob := CalculateFreqWeight(float64(count))
		if err := db.Save(word, prob); err != nil {
			return err
		}
	}

	return nil
}

var CompileCommand = cli.Command{
	Name:   "compile",
	Action: Compile,
	Usage:  "",

	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "wordfreq",
			Value: "",
			Usage: "",
		},
	},
}
