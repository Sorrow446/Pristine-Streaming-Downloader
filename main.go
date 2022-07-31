package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit" +
		"/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
	siteUrl     = "https://www.pristineclassical.com/"
	checkoutUrl = siteUrl + "pages/player-auth"
	apiKey      = "8a2f7304-55bf-46da-8951-da243f25142c"
	streamBase  = "https://pristinestreaming.com/"
	apiBase     = streamBase + "api/v1/"
	regexStr    = streamBase + `app/browse/albums/(\d+)`
)

var (
	jar, _ = cookiejar.New(nil)
	client = &http.Client{Transport: &Transport{}, Jar: jar}
)

var qualityMap = map[int]Format{
	1: {"320 Kbps MP3", ".mp3", ""},
	2: {"FLAC", ".flac", ""},
}

var titleRegexes = [2]string{
	`^\d+ - (.+)`,
	`^\(\d+\) \[.+] (.+)`,
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", userAgent,
	)
	return http.DefaultTransport.RoundTrip(req)
}

func handleErr(errText string, err error, _panic bool) {
	errString := errText + "\n" + err.Error()
	if _panic {
		panic(errString)
	}
	fmt.Println(errString)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	var speed int64 = 0
	n := len(p)
	wc.Downloaded += int64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	toDivideBy := time.Now().UnixMilli() - wc.StartTime
	if toDivideBy != 0 {
		speed = int64(wc.Downloaded) / toDivideBy * 1000
	}
	fmt.Printf("\r%d%% @ %s/s, %s/%s ", wc.Percentage, humanize.Bytes(uint64(speed)),
		humanize.Bytes(uint64(wc.Downloaded)), wc.TotalStr)
	return n, nil
}

func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(fname), nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string) bool {
	for _, line := range lines {
		if strings.EqualFold(line, value) {
			return true
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine) {
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url) {
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

func parseCfg() (*Config, error) {
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if args.Format != -1 {
		cfg.Format = args.Format
	}
	if !(cfg.Format == 1 || cfg.Format == 2) {
		return nil, errors.New("Format must be 1 or 2.")
	}
	if args.OutPath != "" {
		cfg.OutPath = args.OutPath
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "Pristine Streaming downloads"
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		errString := "Failed to process URLs.\n" + err.Error()
		return nil, errors.New(errString)
	}
	return cfg, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func makeDirs(path string) error {
	return os.MkdirAll(path, 0755)
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkUrl(_url string) string {
	regex := regexp.MustCompile(regexStr)
	match := regex.FindStringSubmatch(_url)
	if match != nil {
		return match[1]
	}
	return ""
}

func auth(email, password string) error {
	data := url.Values{}
	data.Set("form_type", "customer_login")
	data.Set("utf8", "✓")
	data.Set("checkout_url", checkoutUrl)
	data.Set("customer[email]", email)
	data.Set("customer[password]", password)
	req, err := http.NewRequest(
		http.MethodPost, siteUrl+"account/login", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Referer", siteUrl)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return errors.New(do.Status)
	}
	if do.Request.URL.String() != checkoutUrl {
		return errors.New("Bad credentials?")
	}
	return nil
}

func getAlbumMeta(albumId, ref string) (*AlbumMeta, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"albums/"+albumId, nil)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("with[0]", "tracks")
	req.URL.RawQuery = query.Encode()
	req.Header.Add("Referer", ref)
	req.Header.Add("X-Pristine-Player-API-Key", apiKey)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj AlbumMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	if !obj.Success {
		return nil, errors.New("Bad response.")
	}
	return &obj, nil
}

func sanitise(filename string) string {
	regex := regexp.MustCompile(`[\/:*?"><|]`)
	return regex.ReplaceAllString(filename, "_")
}

func getStreamMeta(trackId int, ref string) (*StreamMeta, error) {
	req, err := http.NewRequest(
		http.MethodGet, apiBase+"listen/"+strconv.Itoa(trackId), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Referer", ref)
	req.Header.Add("X-Pristine-Player-API-Key", apiKey)
	do, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK {
		return nil, errors.New(do.Status)
	}
	var obj StreamMeta
	err = json.NewDecoder(do.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	if !obj.Success {
		return nil, errors.New("Bad response.")
	}
	return &obj, nil
}

func chooseFormat(wantFmt int, meta *StreamMeta) *Format {
	var chosen Format
	mp3Url := meta.Sources.Mp3
	flacUrl := meta.Sources.Flac
	if wantFmt == 1 {
		chosen = qualityMap[1]
		chosen.URL = mp3Url
	} else if wantFmt == 2 {
		if flacUrl == "" {
			fmt.Println("Unavailable in FLAC.")
			chosen = qualityMap[1]
			chosen.URL = mp3Url
		} else {
			chosen = qualityMap[2]
			chosen.URL = flacUrl
		}
	}
	return &chosen
}

func fixTrackTitle(title string) string {
	for _, regexStr := range titleRegexes {
		regex := regexp.MustCompile(regexStr)
		match := regex.FindStringSubmatch(title)
		if match != nil {
			return match[1]
		}
	}
	return title
}

func downloadTrack(trackPath, url string) error {
	f, err := os.OpenFile(trackPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Range", "bytes=0-")
	do, err := client.Do(req)
	if err != nil {
		return err
	}
	defer do.Body.Close()
	if do.StatusCode != http.StatusOK && do.StatusCode != http.StatusPartialContent {
		return errors.New(do.Status)
	}
	totalBytes := do.ContentLength
	counter := &WriteCounter{
		Total:     totalBytes,
		TotalStr:  humanize.Bytes(uint64(totalBytes)),
		StartTime: time.Now().UnixMilli(),
	}
	_, err = io.Copy(f, io.TeeReader(do.Body, counter))
	fmt.Println("")
	return err
}

func init() {
	fmt.Println(`
 _____ _____    ____                _           _         
|  _  |   __|  |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
|   __|__   |  |  |  | . | | | |   | | . | .'| . | -_|  _|
|__|  |_____|  |____/|___|_____|_|_|_|___|__,|___|___|_|
	`)
}

func main() {
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	cfg, err := parseCfg()
	if err != nil {
		handleErr("Failed to parse config/args.", err, true)
	}
	err = makeDirs(cfg.OutPath)
	if err != nil {
		handleErr("Failed to make output folder.", err, true)
	}
	err = auth(cfg.Email, cfg.Password)
	if err != nil {
		handleErr("Failed to auth.", err, true)
	}
	fmt.Println("Signed in successfully.\n")
	albumTotal := len(cfg.Urls)
	for albumNum, _url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		albumId := checkUrl(_url)
		if albumId == "" {
			fmt.Println("Invalid URL:", _url)
			continue
		}
		_meta, err := getAlbumMeta(albumId, _url)
		if err != nil {
			handleErr("Failed to get album metadata.", err, false)
			continue
		}
		meta := _meta.Album
		albumTitle := meta.Title
		fmt.Println(albumTitle)
		sanAlbumFolder := strings.TrimSuffix(sanitise(albumTitle), ".")
		if len(sanAlbumFolder) > 120 {
			fmt.Println("Album folder was chopped as it exceeds 120 characters.")
			sanAlbumFolder = sanAlbumFolder[:120]
		}
		albumPath := filepath.Join(cfg.OutPath, sanAlbumFolder)
		err = makeDirs(albumPath)
		if err != nil {
			handleErr("Failed to make album folder.", err, false)
			continue
		}
		trackTotal := len(meta.Tracks)
		for trackNum, track := range meta.Tracks {
			trackNum++
			trackTitle := fixTrackTitle(track.Title)
			streamMeta, err := getStreamMeta(track.ID, _url)
			if err != nil {
				handleErr("Failed to get track stream metadata.", err, false)
				continue
			}
			format := chooseFormat(cfg.Format, streamMeta)
			trackFname := fmt.Sprintf("%02d. %s%s", trackNum, sanitise(trackTitle), format.Extension)
			trackPath := filepath.Join(albumPath, trackFname)
			exists, err := fileExists(trackPath)
			if err != nil {
				handleErr("Failed to check if track already exists locally.", err, false)
				continue
			}
			if exists {
				fmt.Println("Track already exists locally.")
				continue
			}
			fmt.Printf(
				"Downloading track %d of %d: %s - %s\n", trackNum, trackTotal, trackTitle,
				format.Specs,
			)
			err = downloadTrack(trackPath, format.URL)
			if err != nil {
				handleErr("Failed to download track.", err, false)
				continue
			}
		}
	}
}
