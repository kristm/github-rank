package main

import (
  "fmt"
  "net/http"
  "log"

  "goji.io"
  "goji.io/pat"
  "github.com/zenazn/goji/graceful"

  "encoding/csv"
  "strings"
  "os"

  "html/template"
)

type RepoCount struct {
  NumForked int
  Total int
}

type Profile struct {
  UserName string
  Name string
  Email string
  Location string
  Url string
  Company string
  Organizations []string
  AvatarUrl string
  Rank string
  Score string
  Languages []string
}

type PageData struct {
  ModifiedDate string
  ActiveUsers []Profile
}

func parseCsv(fileName string) PageData {
  users := []Profile{}

  file, err := os.Open(fileName)
  if err != nil {
    log.Fatal(err)
  }

  stat, err := os.Stat(fileName)
  if err != nil {
    log.Fatal(err)
  }

  defer file.Close()

  result, _ := csv.NewReader(file).ReadAll()
  //TODO: Handle error

  for row := range result {
    if row > 0 { // skip title
      languages := strings.Split(result[row][11], ",")
      users = append(users, Profile{UserName: result[row][2], Name: result[row][1], Email: result[row][3], Location: result[row][4], Languages: languages, Rank: result[row][0], Url: result[row][12], Score: result[row][6], Company: result[row][9], Organizations: strings.Split(result[row][10], ","), AvatarUrl: result[row][5]})
    }
  }

  return PageData{ActiveUsers: users, ModifiedDate: stat.ModTime().Format("Jan 2, 2006 15:04:00 GMT")}
}

func index(w http.ResponseWriter, r *http.Request) {
  content := parseCsv("data/most-active-github-users-default.csv")
  log.Printf("file date: %s", content.ModifiedDate)

  t, _ := template.ParseFiles("views/index.html")
  t.Execute(w, content)
}

func countryIndex(w http.ResponseWriter, r *http.Request) {
  countryCode := pat.Param(r, "country")
  fileName := fmt.Sprintf("data/most-active-github-users-%s.csv", countryCode)
  content := parseCsv(fileName)
  log.Printf("file date: %s", content.ModifiedDate)

  t, _ := template.ParseFiles("views/index.html")
  t.Execute(w, content)
}

func main() {
  mux := goji.NewMux()
  mux.HandleFunc(pat.Get("/"), index)
  mux.HandleFunc(pat.Get("/:country"), countryIndex)

  fs := http.FileServer(http.Dir("public"))
  mux.Handle(pat.Get("/assets/*"), http.StripPrefix("/assets/", fs))

  mux.Use(logRequest)

  log.Printf("listening to localhost:8090")
  graceful.ListenAndServe("localhost:8090", mux)
}

//middleware
// thanks to http://ndersson.me/post/capturing_status_code_in_net_http/
type loggingResponseWriter struct {
  http.ResponseWriter
  statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
  return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
  lrw.statusCode = code
  lrw.ResponseWriter.WriteHeader(code)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
    lrw := NewLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)

    statusCode := lrw.statusCode
    log.Printf("%d %s\n", statusCode, http.StatusText(statusCode))
	})
}
