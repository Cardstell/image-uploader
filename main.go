package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
	"strconv"
	"github.com/gorilla/mux"
	"strings"
	"math/rand"
	"os/exec"
	"io/ioutil"
	"time"
)

var (
	db_file *os.File
	db_content []string
	countUploads map[string]int
	lastUpdate string
)

func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	ipport := r.RemoteAddr
	if forwarded != "" {
		ipport = forwarded
	}
	ipports := strings.Split(ipport, ":")
	ipports = ipports[:len(ipports)-1]
	return strings.Join(ipports, ":")
}

func isUnauthorized(r *http.Request) bool {
	return false
	// cookie, err := r.Cookie("cid")
	// if err != nil {
	// 	return true
	// }
	// return cookie.Value != cid
}

func addItemToDB(item string) {
	db_content = append(db_content, item)
	_, err := db_file.WriteString(item + "\n")
	if err != nil {
		panic(err)
	}
}

func getRandomFileName() string {
	const N = 6
	s := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, N)
	for i := range result {
		result[i] = s[rand.Intn(len(s))]
	}
	return string(result) + ".png"
}

func generatePreview(filename string, previewFilename string) error {
	geometry := strconv.Itoa(previewWidth) + "x" + strconv.Itoa(previewHeight)
	var args = []string{
		filename, "-resize", geometry, previewFilename,
	}
	path, err := exec.LookPath("convert")
	if err != nil {
		return err
	}
	cmd := exec.Command(path, args...)
	return cmd.Run()
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	day := time.Now().Format("02-01-2006")
	if day != lastUpdate {
		lastUpdate = day
		countUploads = make(map[string]int)
	}
	if countUploads[ip] >= maxUploadsPerIP {
		fmt.Fprintf(w, "{\"ok\":\"false\", \"result\":\"Uploading failed!\"}")
		return
	}
	countUploads[ip] += 1
	if isUnauthorized(r) {
		http.Error(w, "404 file not found", 404)
		return
	}

	r.ParseMultipartForm(1 << 25)
	reqfile, _, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "{\"ok\":\"false\", \"result\":\"Uploading failed!\"}")
		return
	}
	defer reqfile.Close()
	
	filename := getRandomFileName()
	previewFilename := getRandomFileName()

	file, err := os.Create("img/" + filename)
	if err != nil {
		fmt.Fprintf(w, "{\"ok\":\"false\", \"result\":\"Uploading failed!\"}")
		return
	}
	io.Copy(file, reqfile)
	file.Close()
	err = generatePreview("img/" + filename, "img/" + previewFilename)
	if err != nil {
		os.Remove("img/" + filename)
		fmt.Fprintf(w, "{\"ok\":\"false\", \"result\":\"Uploading failed!\"}")
		return
	}
	addItemToDB(filename + "," + previewFilename + "," + time.Now().Format("02-01-2006 15:04") + 
		"," + ip)

	url := "https://" + host + prefix + "/img/" + filename
	previewURL := "https://" + host + prefix + "/img/" + previewFilename
	fmt.Fprintf(w, "{\"ok\":\"true\", \"result\":\"Success!<br/><a href=\\\"" + url + "\\\">" +
		url + "</a>\", \"url\":\"" + previewURL + "\"}")
}

func uploadPageHandler(w http.ResponseWriter, r *http.Request) {
	if isUnauthorized(r) {
		http.Error(w, "404 file not found", 404)
		return
	}
	http.ServeFile(w, r, "src/upload.html")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if isUnauthorized(r) {
		http.Error(w, "404 file not found", 404)
		return
	}
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "{\"ok\":\"false\"}")
		return;
	}
	var start, end int
	var err error
	if start, err = strconv.Atoi(r.FormValue("start")); err != nil {
		fmt.Fprintf(w, "{\"ok\":\"false\"}")
		return;
	}
	if end, err = strconv.Atoi(r.FormValue("end")); err != nil {
		fmt.Fprintf(w, "{\"ok\":\"false\"}")
		return;
	}
	if start < end || end < 0 || start >= len(db_content) || start - end > 100 {
		fmt.Fprintf(w, "{\"ok\":\"false\"}")
		return;
	}
	result := "{\"ok\":\"true\", \"result\":[";
	for i := start;i>=end;i-- {
		current := strings.Split(db_content[i], ",");
		result += "{\"url\":\"" + "https://" + host + prefix + "/img/" + current[0] + "\",\"previewURL\":\"" +
			"https://" + host + prefix + "/img/" + current[1] + "\",\"time\":\"" + current[2] + "\"}";
		if i != end {
			result += ",";
		}
	}
	result += "]}";
	fmt.Fprintf(w, result)
}

type Template struct {
	Index string
}

func downloadPageHandler(w http.ResponseWriter, r *http.Request) {
	if isUnauthorized(r) {
		http.Error(w, "404 file not found", 404)
		return
	}
	html, _ := ioutil.ReadFile("src/all.html")
	fmt.Fprintf(w, string(html), len(db_content)-1)
}

func staticHandlers(dir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := mux.Vars(r)["filename"]
		file, err := os.Open("./" + dir + "/" + filename)
		defer file.Close()
		if err != nil {
			http.Error(w, "404 file not found", 404)
			return
		}
		header := make([]byte, 512)
		file.Read(header)
		FileContentType := http.DetectContentType(header)
		if strings.HasSuffix(filename, ".css") {
			FileContentType = "text/css"
		} else if strings.HasSuffix(filename, ".js") {
			FileContentType = "text/javascript"
		} else if strings.HasSuffix(filename, ".png") {
			FileContentType = "image/png"
		}
		FileStat, _ := file.Stat()
		FileSize := strconv.FormatInt(FileStat.Size(), 10)
		// w.Header().Set("Content-Disposition", "attachment;filename="+filename)
		w.Header().Set("Content-Type", FileContentType)
		w.Header().Set("Content-Length", FileSize)
		file.Seek(0, 0)
		io.Copy(w, file) 
	}
}

// func setCookieHandler() func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		id := mux.Vars(r)["id"]
// 		if id != cid {
// 			http.Error(w, "404 page not found", 404)
// 			return
// 		}

// 		http.SetCookie(w, &http.Cookie{
// 			Name: "cid", 
// 			Value: cid, 
// 			MaxAge: 0, 
// 			Path: "/",
// 		})

// 		fmt.Fprintf(w, "ok")	
// 	}
// }

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	file, err := os.Open(db_filename)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(file)
	db_content = strings.Split(string(b), "\n")
	if len(db_content) > 0 && len(db_content[len(db_content)-1]) == 0 {
		db_content = db_content[:len(db_content)-1]
	}
	file.Close()
	db_file, err = os.OpenFile(db_filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer db_file.Close()

	r := mux.NewRouter()
	r.HandleFunc(prefix + download_prefix, downloadPageHandler).Methods("GET")
	r.HandleFunc(prefix + download_prefix, downloadHandler).Methods("POST")
	r.HandleFunc(prefix + "/", uploadPageHandler).Methods("GET")
	r.HandleFunc(prefix + "/", uploadHandler).Methods("POST")
	r.HandleFunc(prefix + "/static/{filename}", staticHandlers("static")).Methods("GET", "POST")
	r.HandleFunc(prefix + "/img/{filename}", staticHandlers("img")).Methods("GET", "POST")
	// r.HandleFunc(prefix + "/cookie/{id}", setCookieHandler()).Methods("GET")
	http.ListenAndServe(":" + port, r)
}