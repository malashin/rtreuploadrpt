package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var inputPath = "input.txt"
var outputPath = "output.txt"
var dbPath = "database.db"
var re = regexp.MustCompile(`^((?:sd|hd)_\d{4}(?:_3d)?_[a-zA-Z0-9_]+)__(?:\w+_)*(trailer|film)\.(?:mp4|mpg|mkv)$`)
var reSeries = regexp.MustCompile(`^(.+_s\d+)_\d+$`)
var inputMap = make(map[string]Movie)
var dbMap = make(map[string]Movie)
var buf = new(bytes.Buffer)
var e = gob.NewEncoder(buf)
var d = gob.NewDecoder(buf)

type Movie struct {
	Name    string
	MovFile []string
	TrlFile string
}

type Series struct {
	Name    string
	TrlFile string
}

func main() {
	files, err := readLines(inputPath)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !re.MatchString(file) {
			panic(file + ": WRONG FILENAME")
		}

		var m Movie
		m.Name = re.ReplaceAllString(file, "${1}")
		videoType := re.ReplaceAllString(file, "${2}")

		if reSeries.MatchString(m.Name) {
			m.Name = reSeries.ReplaceAllString(m.Name, "${1}")
		}

		if a, ok := inputMap[m.Name]; ok {
			switch videoType {
			case "film":
				a.MovFile = append(a.MovFile, file)
			case "trailer":
				a.TrlFile = file
			default:

			}
			inputMap[m.Name] = a
		} else {
			switch videoType {
			case "film":
				m.MovFile = append(a.MovFile, file)
			case "trailer":
				m.TrlFile = file
			default:
				panic(file + ": NOT TRAILER OR FILM")
			}
			inputMap[m.Name] = m
		}
	}

	if len(inputMap) == 0 {
		return
	}

	if _, err = os.Stat(dbPath); err == nil {
		err = decodeCacheFile(&dbMap, dbPath)
		if err != nil {
			panic(err)
		}
	}

	var keys []string
	for k := range inputMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var output []string
	for _, k := range keys {
		m := inputMap[k]
		if len(m.MovFile) > 0 && m.TrlFile != "" {
			if _, ok := dbMap[m.Name]; !ok {
				movFiles := m.MovFile[0]
				if len(m.MovFile) > 1 {
					movFiles = "\"" + strings.Join(m.MovFile, "\n") + "\""
				}

				date := time.Now().Format("02.01.2006")
				fmt.Println(movFiles + "\t" + m.TrlFile + "\t" + date)
				output = append(output, movFiles+"\t"+m.TrlFile+"\t"+date+"\n")
				dbMap[k] = m
			}
		}
	}

	if _, err = os.Stat(dbPath); err == nil {
		copyFile(dbPath, dbPath+".backup")
	}
	err = encodeCacheFile(dbMap, dbPath)
	if err != nil {
		panic(err)
	}

	if _, err = os.Stat(outputPath); err == nil {
		copyFile(outputPath, outputPath+".backup")
	}
	writeStringArrayToFile(outputPath, output, 0775)
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

func writeStringArrayToFile(filename string, strArray []string, perm os.FileMode) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	for _, v := range strArray {
		if _, err = f.WriteString(v); err != nil {
			log.Panic(err)
		}
	}
}

func encodeCacheFile(i interface{}, path string) error {
	err := e.Encode(i)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, buf.Bytes(), 0755)
	if err != nil {
		return err
	}
	buf.Reset()
	return nil
}

func decodeCacheFile(i interface{}, path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = buf.Write(b)
	if err != nil {
		return err
	}
	err = d.Decode(i)
	if err != nil {
		return err
	}
	buf.Reset()
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
