package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	UploadDir = "./upload/"
)

type home struct {
	Title string
}

var mux map[string]func(http.ResponseWriter, *http.Request)

type THandler struct{}

func (*THandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
		return
	}
	if ok, _ := regexp.MatchString("/css/", r.URL.String()); ok {
		http.StripPrefix("/css/", http.FileServer(http.Dir("./static/css/"))).ServeHTTP(w, r)
	} else {
		http.StripPrefix("/", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
	}
}

func main() {
	fmt.Println("-------->")
	server := http.Server{
		Addr:        ":8012",
		Handler:     &THandler{},
		ReadTimeout: 10 * time.Second,
	}
	mux["/"] = index
	mux["/upload"] = Upload
	mux["/download"] = Download
	mux["/file"] = StaticServer

	server.ListenAndServe()
}

func StaticServer(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/file", http.FileServer(http.Dir("./upload/"))).ServeHTTP(w, r)
}

func index(w http.ResponseWriter, r *http.Request) {
	title := home{Title: "首页"}
	t, err := template.ParseFiles("./views/" + "index.tpl")
	if err != nil {
		fmt.Fprintf(w, "%v", err)
		return
	}
	t.Execute(w, title)
}

func Upload(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, err := template.ParseFiles("./views/file.tpl")
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "模版解析失败", err.Error())
			return
		}
		t.Execute(w, "文件上传")
	} else {
		r.ParseMultipartForm(32 << 20) //32M
		file, handler, err := r.FormFile("file")
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}
		fileext := filepath.Ext(handler.Filename)
		if checkType(fileext) == false {
			fmt.Fprintf(w, "%v", "类型不允许")
			return
		}
		filename := strconv.FormatInt(time.Now().Unix(), 10) + fileext
		f, err := os.OpenFile(UploadDir+filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			fmt.Println("open err:", err)
			fmt.Fprintf(w, "%v", "打开失败")
			return
		}
		defer f.Close()
		_, err = io.Copy(f, file)
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}
		filedir, _ := filepath.Abs(UploadDir + filename)
		fmt.Fprintf(w, "%v", filename+"上传完成，服务器地址:"+filedir)
		return
	}

}

func Download(w http.ResponseWriter, r *http.Request) {
	dl := r.FormValue("dl")
	filename := r.FormValue("filename")
	dlint, _ := strconv.Atoi(dl)
	if dlint > 1 {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(filename)))
	}

	http.FileServer(http.Dir(filename)).ServeHTTP(w, r)
}

func checkType(t string) bool {
	ext := []string{".exe"}
	for _, v := range ext {
		if v == t {
			return false
		}
	}
	return true
}

func init() {
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
}
