package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var config Config
var linksCount int
var ssConfigs []SsServerConf
var rawSsConf []byte

type Link struct {
	Url           string
	Mask          []string
	ConfigCount   int
	ParseTopToBot bool
}

type Config struct {
	SsConfigFile        string
	SsPath              string
	SsRestartCommand    []string
	SsConfigSectionPath []string
	SsServersEditPos    int
	SsModeDefault       string
	SsTimeOutDefault    int32
	OutputFile          string
	Links               []Link
}

type SsConfigs struct {
	Configs []SsServerConf `json:"ss"`
}

type SsServerConf struct {
	Server   string `json:"server"`
	Port     int    `json:"server_port"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Timeout  int32  `json:"timeout,omitempty"`
	//Mode     string `json:"mode,omitempty"`
}

func decodeSsServerConfig(str string) {
	var datastr string
	index := strings.IndexByte(str, '@')
	if index == -1 { // fully encoded string
		data, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		datastr = string(data[:])
	} else { // encoded only method:password
		shortstr := str[:index]
		data, err := base64.StdEncoding.DecodeString(shortstr)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		datastr = string(data[:]) + str[index:]
	}
	errstr := createSsServerConfig(datastr)
	if errstr != "" {
		fmt.Println(errstr)
	} else {
		
	}
}

func createSsServerConfig(str string) (errstr string) { //, errcode int
	var index int
	ind := strings.IndexByte(str, '@')
	if ind == -1 {
		errString := "Invalid format of string " + str
		return errString //, 1
	} else {
		mpstr := str[:ind]
		conf := new(SsServerConf)
		index = strings.IndexByte(mpstr, ':')
		if index == -1 {
			errString := "Invalid format of string " + mpstr
			return errString //, 2
		} else {
			conf.Method = mpstr[:index]
			conf.Password = mpstr[index+1:]
		}
		spstr := str[ind+1:]
		// find '?'
		indx := strings.IndexByte(spstr, '/')
		if indx != -1 {
			spstr = spstr[:indx]
			//paramstr := spstr[indx+1:]

		} else {
			indx := strings.IndexByte(spstr, '?')
			if indx != -1 {
				spstr = spstr[:indx]
				//paramstr := spstr[indx+1:]

			}
		}
		index = strings.IndexByte(spstr, ':')
		if index == -1 {
			errString := "Invalid format of string " + spstr
			return errString //, 3
		} else {
			conf.Server = spstr[:index]
			i, err := strconv.Atoi(spstr[index+1:])
			if err != nil {
				errString := "Invalid format of port " + spstr
				return errString //, 4
			}
			conf.Port = i
		}
		//if conf.Mode == "" && config.SsModeDefault != "" {
		//	conf.Mode = config.SsModeDefault
		//}
		if config.SsTimeOutDefault != 0 {
			conf.Timeout = config.SsTimeOutDefault
		}
		ssConfigs = append(ssConfigs, *conf)
	}
	return "" //, 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func readConfig(path string) {
	file, err := os.Open(path)
	if err != nil { // если возникла ошибка
		fmt.Println("Unable to create file:", err)
		os.Exit(1) // выходим из программы
	}
	defer file.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Unable to read file:", err)
		os.Exit(1)
	}
	jsonErr := json.Unmarshal(data, &config)
	if jsonErr != nil {
		fmt.Println("Unable to parse json:", jsonErr)
		os.Exit(1)
	}

}

func parseUp(link Link, body string) {
	lastPos := len(body)
	count := link.ConfigCount
	maskLen := len(link.Mask)
	i := lastPos - 10
	for i >= 0 {
		for j := 0; j < maskLen; j++ {
			mask := link.Mask[0]
			if body[i] == mask[0] {
				lm := len(mask)
				_mask := body[i : i+lm]
				if mask == _mask {
					c := i + lm
					for c <= lastPos {
						if body[c] == '#' { // || body[c] == '?'
							str := body[i+lm : c]
							if mask == "ss://" {
								decodeSsServerConfig(str)
								break
							}
							
						}
						c++
					}
					count = count - 1
					i = i - 10
					lastPos = i
				}
			}
		}
		if count == 0 {
			break
		}
		i = i - 1
	}
}

func parseDown(link Link, body string) {
	lastPos := len(body)
	count := link.ConfigCount
	maskLen := len(link.Mask)
	for i := 0; i < lastPos; i++ {
		for j := 0; j < maskLen; j++ {
			mask := link.Mask[0]
			if body[i] == mask[0] {
				lm := len(mask)
				_mask := body[i : i+lm]
				if mask == _mask {
					c := i + lm
					for c <= lastPos {
						if body[c] == '#' { // || body[c] == '?'
							str := body[i+lm : c]
							if mask == "ss://" {
								decodeSsServerConfig(str)
								break
							}
						}
						c++
					}
					count = count - 1
					i = c
					//lastPos = i
				}
			}
		}
		if count == 0 {
			break
		}
	}
}

func parse(link Link, body string) {
	if link.ParseTopToBot {
		parseDown(link, body)
	} else {
		parseUp(link, body)
	}
}

func getHtml(link Link, wg *sync.WaitGroup) {
	defer wg.Done()
	response, err := http.Get(link.Url)
	if err != nil {
		fmt.Println("Unable to connect to server:", err)
	} else if response.StatusCode == 200 {
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Unable to read html body:", err)
		} else {
			parse(link, string(body))
		}
	} else {
		fmt.Println("Unable to get html:", err)
	}
}

func saveSsConfigs(resFile os.File) bool {
	SsConfigs := new(SsConfigs)
	SsConfigs.Configs = ssConfigs
	jsondata, err := json.MarshalIndent(SsConfigs, "", "	") //ssConfigs
	if err != nil {
		fmt.Println("json encoding error", err)
		return false
	} else {
		_, err := resFile.Write(jsondata)
		if err != nil {
			fmt.Println("json writning err", err)
			return false
		} else {
			return true
		}
	}
}

func main() {
	args := os.Args
	args_count := len(args)
	if args_count > 1 {
		if args[1] == "help" {
			fmt.Println("help not ready")
			os.Exit(1)
		} else if args[1] == "version" {
			fmt.Println("version 1.0")
			os.Exit(1)
		} else {
			path := args[1]
			if fileExists(path) {
				readConfig(path)
				//fmt.Println(config)
			} else {
				fmt.Println("config file not exists")
				os.Exit(1)
			}
		}
	}
	linksCount = len(config.Links)
	if linksCount == 0 {
		fmt.Println("Links for parsing is not defined")
		os.Exit(1)
	}
	var waitgroup sync.WaitGroup
	resultFile, err := os.Create(config.OutputFile)
	if err != nil { // 
		fmt.Println("Unable to create file:", err)
	}
	defer resultFile.Close()
	for i := 0; i < linksCount; i++ {
		waitgroup.Add(1)
		go getHtml(config.Links[i], &waitgroup)
	}
	waitgroup.Wait()
	ssConfToSave := len(ssConfigs)
	if ssConfToSave > 0 {
		if saveSsConfigs(*resultFile) {
			rawSsConf, err = os.ReadFile(config.OutputFile) //
			if err != nil {
				fmt.Println("Unable to read parsingresult file:", err)
			}
			middle := ReadSection("ss", rawSsConf) // rawSsConf[1 : len(rawSsConf)-1]
			if middle != nil && setSsServiceConfig(config.SsConfigFile, middle) {
				RestartSs() // restart ss
			} else {
				ssConfToSave = 0
			}
		}
	}
	fmt.Println("parser finish")
}

func RestartSs() {
	ssbin, lookerr := exec.LookPath(config.SsPath)
	if lookerr != nil {
		fmt.Println("Unable to find ss bin:", lookerr)
	} else {
		//env := os.Environ()
		cmd := exec.Command(ssbin, config.SsRestartCommand...)
		cmd.Stdout = os.Stdout
		err := cmd.Start() //syscall.Exec(ssbin,config.SsRestartCommand,env)
		if err != nil {
			fmt.Println("Unable to restart ss:", err)
		}
	}
}

func setSsServiceConfig(path string, middle []byte) bool {
	if fileExists(path) {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil { // 
			fmt.Println("Unable to open file:", err)
			return false
		}
		defer file.Close()
		data, err := os.ReadFile(path)
		//fileInfo, err := os.Stat(path)
		//perm := fileInfo.Mode().Perm()
		if err != nil {
			fmt.Println("Unable to read SS config file:", err)
			return false
		}
		secPos := findSection(data, config.SsConfigSectionPath)
		res, startPosToEdit, endPosToEdit := findPosToEdit(data, secPos)
		if res {
			first := data[:startPosToEdit+1]
			if config.SsServersEditPos > 0 {
				first = append(first, ',')
			}
			last := data[endPosToEdit:]
			newdata := bytes.Join([][]byte{first, middle, last}, nil) //make([]byte, 0, len(first)+len(ssConfigs)+len(last))
			_, writeerr := file.Write(newdata)                        //os.WriteFile(path, newdata, perm)
			if writeerr != nil {
				fmt.Println("Unable to write ss config file:", writeerr)
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func ReadSection(name string, data []byte) (res []byte) {
	datalen := len(data)
	namelen := len(name)
	for i := 0; i < datalen; i++ {
		if data[i] == name[0] {
			_name := string(data[i : i+namelen])
			if _name == name {
				for j := i + namelen; j < datalen; j++ {
					if data[j] == '[' {
						{ // start array
							endpos := findTokenEnd(data, j+1, datalen, '[', ']')
							if endpos > 0 {
								res = data[j+1 : endpos]
								return res
							} else {
								return nil
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func findNextSection(data []byte, section string, pos int, datalen int) (nextpos int) {
	res := -1
	seclen := len(section)
	for i := pos; i < datalen; i++ {
		if data[i] == section[0] {
			name := string(data[i : i+seclen])
			if name == section { // section found
				res = i + seclen
				return res
			}
		}
	}
	return res
}

func findSection(data []byte, sectionPart []string) (startPos int) {
	res := -1
	datalen := len(data)
	slen := len(sectionPart)
	pos := 0
	for s := 0; s < slen; s++ {
		section := sectionPart[s]
		pos = findNextSection(data, section, pos, datalen)
		if pos < 0 {
			return pos
		}
	}
	if pos > 0 {
		res = pos
	}
	return res
}

func findPosToEdit(data []byte, startpos int) (res bool, start int, end int) {
	res = false
	datalen := len(data)
	count := 0
	for i := startpos; i < datalen; i++ {
		if data[i] == '[' {
			end = findTokenEnd(data, i+1, datalen, '[', ']')
			if config.SsServersEditPos == 0 {
				start = i
				res = true
				return res, start, end
			}
			if end > 0 {
				for j := i; j < end; j++ {
					if data[j] == '{' { //
						count++
						c := findTokenEnd(data, j+1, end, '{', '}')
						if config.SsServersEditPos == count {
							start = c
							res = true
							return res, start, end
						}
					}
				}
			}
		}
	}
	return res, 0, 0
}

func findTokenEnd(data []byte, startpos int, len int, token byte, closeToken byte) (endpos int) {
	count := 0
	for i := startpos; i < len; i++ {
		switch data[i] {
		case token:
			{
				count++
			}
		case closeToken:
			{
				if count == 0 {
					return i
				} else {
					count--
				}
			}
		}
	}
	return -1 // error - token not found
}
