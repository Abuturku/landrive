//Authors: Andreas Schick (2792119), Linda Latreider (7743782), Niklas Nikisch (9364290)
package server

import (
	//"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var AuthenticatorVar Authenticator = (Authenticator)(&AuthenticatorStruct{})

type AuthenticatorStruct struct {
	AuthenticatorFunc
}

/*
StartFileserver setzte alle Adressen für die Händler und Startet den Server für die zuvor angegebenen Parameter
*/
func StartFileserver() {
	log.Println("Server Startet")
	http.HandleFunc("/", index)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", newUserHandler)
	http.HandleFunc("/landrive", landrive)
	http.HandleFunc("/uploadFile", uploadFileHandler)
	http.HandleFunc("/getFolderStruct", folderStructHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/newFolder", createFolderHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/wget", wgetHandler)
	http.HandleFunc("/changePw", changePasswordHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("website"))))
	err := http.ListenAndServeTLS(":"+flag.Lookup("P").Value.String(), flag.Lookup("C").Value.String(), flag.Lookup("K").Value.String(), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

/*
Wird aufgerufen wenn ein User die Webseite (Startseite) betritt. Sollte er einen gültigen cookie besitzen wird er sofort auf die Hauptseite weitergeleitet
*/
func index(w http.ResponseWriter, req *http.Request) {
	cookiecheck, _, _ := checkCookie(w, req)
	log.Println(cookiecheck)
	if cookiecheck {
		http.Redirect(w, req, "/landrive", http.StatusMovedPermanently)
	} else {
		t, _ := template.ParseFiles("website/index.html")
		t.Execute(w, nil)
	}
}

/*
Wird beim Betreten der Hauptseite aufgerufen. Sollte jedoch kein gültiger Cookie vorhanden sein wird man auf die Startseite weitergeleitet.
*/
func landrive(w http.ResponseWriter, req *http.Request) {
	cookiecheck, _, _ := checkCookie(w, req)
	if cookiecheck {
		t, err := template.ParseFiles("website/landrive.html")
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println(string(b))

		t.Execute(w, nil)
	} else {
		http.Redirect(w, req, "/", http.StatusMovedPermanently)
	}
}

/*
Wird genutzt um das Ordnerverzeichnes eines Users zurück zu geben. Zuvor wird geprüft ob ein gültiger Cookie vorhanden ist.
*/
func folderStructHandler(w http.ResponseWriter, req *http.Request) {
	cookiecheck, user, _ := checkCookie(w, req)
	if cookiecheck {
		log.Println("Loading FolderStruct for user " + user.name)
		folders := getFolderStruct(user.name)
		js, err := json.Marshal(folders)
		if err != nil {
			fmt.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		http.Redirect(w, req, "/", http.StatusMovedPermanently)
	}
}

/*
Wird genutzt um einen User aus zu loggen und somit den Cookie zu löschen.
*/
func logoutHandler(w http.ResponseWriter, req *http.Request) {
	cookiecheck, user, cookie := checkCookie(w, req)
	if cookiecheck {
		log.Println("Logout: " + user.name)
		cookie.Expires = time.Now().Add(-1)
		cookie.Value = ""
		http.SetCookie(w, &cookie)
		http.Redirect(w, req, "/", http.StatusMovedPermanently)
	} else {
		http.Redirect(w, req, "/", http.StatusNotModified)
	}
}

/*
Methode wird aufgerufen wenn ein neuer Ordner erstellt werden soll.
*/
func createFolderHandler(w http.ResponseWriter, req *http.Request) {
	cookiecheck, user, _ := checkCookie(w, req)
	if cookiecheck {

		path := req.FormValue("path")
		newFolderName := req.FormValue("newFolderName")
		log.Println("Create Folder: " + path + "/" + newFolderName)
		createFolder(user.name + "/" + path + "/" + newFolderName)
		http.Redirect(w, req, "/", http.StatusFound)
	} else {
		http.Redirect(w, req, "/", http.StatusMovedPermanently)
	}
}

/*
Löscht einen Ordner oder eine Datei der anhand des übergebenen Pfades identifiziert wird.
*/
func deleteHandler(w http.ResponseWriter, req *http.Request) {
	cookiecheck, user, _ := checkCookie(w, req)
	if cookiecheck {

		path := req.FormValue("path")
		log.Println("delete: " + path)
		if path != "" {
			os.RemoveAll(flag.Lookup("F").Value.String() + user.name + "/" + path)
		} else {
			log.Println("Path is empty")
		}

		http.Redirect(w, req, "/", http.StatusFound)
	} else {
		http.Redirect(w, req, "/", http.StatusMovedPermanently)
	}
}

/*
Startet den Download der Angeforderten Datei
*/
func downloadHandler(w http.ResponseWriter, req *http.Request) {
	cookiecheck, user, _ := checkCookie(w, req)
	if cookiecheck {
		path := req.FormValue("path")
		stringarray := strings.Split(path, "/")
		log.Println("Download File: " + path)

		w.Header().Set("Content-Disposition", "attachment; filename=\""+stringarray[len(stringarray)-1]+"\"")

		http.ServeFile(w, req, flag.Lookup("F").Value.String()+user.name+"/"+path)

	} else {
		http.Redirect(w, req, "/", http.StatusMovedPermanently)
	}
}

/*
Startet den Download der Angeforderten Datei welchen über eine Konsole getätigt wurde
*/
//wget muss mit den Parameter --no-check-certificate und --auth-no-challenge augerufen werden (am besten auch noch mit --content-disposition
//z.B. wget --user=[username] --password=[password] --no-check-certificate --auth-no-challenge --content-disposition https://[host]:[port]/wget?path=[filepath]
func wgetHandler(w http.ResponseWriter, req *http.Request) {

	username, password, _ := req.BasicAuth()

	user := loadUser(username)

	log.Println("testing wget")

	if AuthenticatorVar.Authenticate(user, password) {
		path := req.URL.Query().Get("path")
		stringarray := strings.Split(path, "/")
		log.Println("Download File from wget: " + path)

		w.Header().Set("Content-Disposition", "attachment; filename=\""+stringarray[len(stringarray)-1]+"\"")

		http.ServeFile(w, req, flag.Lookup("F").Value.String()+user.name+"/"+path)
	} else {
		w.WriteHeader(401)
	}

}

/*
Prüft einen Cookie, ob er valide ist.
*/
func checkCookie(w http.ResponseWriter, req *http.Request) (bool, user, http.Cookie) {
	cookies := req.Cookies()

	for _, cookie := range cookies {

		cookieName := cookie.Name
		cookiePw := cookie.Value
		user := loadUser(cookieName)
		if cookiePw == hash([]string{user.name, user.password}) {
			//following 3 lines are responsible for setting the cookie expiration date 15 minutes to the future
			expiration := time.Now().Add(15 * time.Minute)
			cookie.Expires = expiration
			http.SetCookie(w, cookie)
			return true, *user, *cookie
		}

	}
	return false, user{}, http.Cookie{}
}

/*
Prüft Name und Passwort beim Login und leitet bei erfolgreicher Überprüfung den Login weiter
*/
func loginHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User tried to log in")
	username := req.FormValue("username")
	password := req.FormValue("password")
	log.Println("User:", username, "Password:", password)
	user := loadUser(username)
	log.Println("Found user: ", user)
	authenticationSuccessful := AuthenticatorVar.Authenticate(user, password)

	if authenticationSuccessful {
		loginUser(user, w, req)
	} else {
		http.Redirect(w, req, "?login=false", http.StatusMovedPermanently)
	}

}

/*
Loged den User ein, indem ein Cookie für diesen erstellt und gesetzt wird.
*/
func loginUser(user *user, w http.ResponseWriter, req *http.Request) {
	cookieValue := hash([]string{user.name, user.password})
	maxAge, _ := strconv.Atoi(flag.Lookup("T").Value.String())
	cookie := http.Cookie{Name: user.name, Value: cookieValue, MaxAge: maxAge, Expires: time.Now().Add(15 * time.Minute)}
	log.Println("Setting cookie")
	http.SetCookie(w, &cookie)
	log.Println("Redirecting to landrive")
	http.Redirect(w, req, "/landrive", http.StatusFound)
}

/*
Regestriert einen User. Prüft ob beide Passwörter gleich sind und ob es den User noch nicht gibt
*/
func newUserHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("User tried to register")
	username := req.FormValue("username")
	password := req.FormValue("password")
	password2 := req.FormValue("password2")

	if password == password2 {
		user := loadUser(username)
		if user.name != "" {
			http.Redirect(w, req, "?register=userfalse", http.StatusMovedPermanently)
		} else {
			user := createUser(username, password)
			loginUser(&user, w, req)
		}
	} else {
		http.Redirect(w, req, "?register=pwfalse", http.StatusMovedPermanently)
	}

}

/*
Fügt einen User der CSV Datei auf dem Server hinzu
*/
func createUser(username string, password string) user {
	salt := generateSalt()
	hashedPw := hash([]string{password, salt})
	log.Println("Creating user with parameters: ", username, hashedPw, salt)
	f, err := os.OpenFile(flag.Lookup("L").Value.String(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	writer := csv.NewWriter(f)
	defer f.Close()

	//writer.Write(username)
	//writer.Write(hashedPw)
	//writer.Write(salt)
	log.Println("Writing to csv")
	writer.Write([]string{username, hashedPw, salt})
	writer.Flush()
	err = writer.Error()
	if err != nil {
		log.Println(err)
	}

	createFolder(username)
	return user{name: username, password: hashedPw, salt: salt}
}

/*
Generiert einen zufälligen Wert. Den Salt
*/
func generateSalt() string {
	saltSize := 16
	buf := make([]byte, saltSize)
	_, err := rand.Read(buf)

	if err != nil {
		fmt.Printf("random read failed: %v", err)
	}

	return hex.EncodeToString(buf)
}

/*
Repräsentiert den User
*/
type user struct {
	name     string
	password string
	salt     string
}

/*
Läd den User aus der CSV Datei, die im Server Liegt
*/
func loadUser(username string) *user {
	f, _ := os.Open(flag.Lookup("L").Value.String())

	r := csv.NewReader(f)
	defer f.Close()
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if record[0] == username {
			return &user{name: record[0], password: record[1], salt: record[2]}
		}
		log.Println(record[0], username)
		log.Println(record[0] == username)
	}
	return &user{name: "", password: "", salt: ""}
}

func (a AuthenticatorFunc) Authenticate(user *user, password string) bool {
	//hasher := sha256.New()
	//hasher.Write([]byte(password))
	//hasher.Write([]byte(user.salt))
	hash := hash([]string{password, user.salt})

	if hash == user.password {
		return true
	}

	return false
}

func hash(strings []string) string {
	hasher := sha256.New()
	for _, value := range strings {
		hasher.Write([]byte(value))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

/*
Erstellt einen Ordner
*/
func createFolder(path string) {
	os.Mkdir(flag.Lookup("F").Value.String()+path, 0777)
}

/*
Repräsentiert einen Ordner
*/
type Folder struct {
	Name    string
	Files   []File
	Folders []Folder
}

/*
Repräsentiert eine Datei
*/
type File struct {
	Name string
	Date time.Time
	Size int64
}

/*
Gibt die Untere Ordnerstruktur eines Pfades zurück
 */
func getFolderStruct(path string) Folder {
	//log.Println(path)
	index := strings.Index(path, "/")
	var name string
	if index > 0 {
		name = path[index+1:]
	} else {
		name = path
	}
	files := make([]File, 0)
	folders := make([]Folder, 0)
	//log.Println(name + ": ")
	fileinfos, _ := ioutil.ReadDir(flag.Lookup("F").Value.String() + "/" + path)

	for _, file := range fileinfos {
		if file.IsDir() {
			folders = append(folders, getFolderStruct(path+"/"+file.Name()))
		} else {
			//log.Println(file.Name())
			fileStruct := File{Name: file.Name(), Date: file.ModTime(), Size: file.Size()}
			files = append(files, fileStruct)
		}
	}
	folder := Folder{Name: name, Files: files, Folders: folders}
	return folder
}

/*
Läd eine Datei die in FormFile und FormValue definiert ist hoch.
*/
func uploadFileHandler(w http.ResponseWriter, req *http.Request) {
	cookiecheck, user, _ := checkCookie(w, req)
	if cookiecheck {
		log.Println("Request to upload a file was made from user " + user.name)

		defer http.Redirect(w, req, "/landrive", http.StatusMovedPermanently)

		//Fileupload orientiert nach https://www.socketloop.com/tutorials/golang-upload-file
		file, header, err := req.FormFile("uploadFile")
		folderPath := req.FormValue("path")
		log.Println("Path where file is going to be placed: " + folderPath)
		if err != nil {
			log.Println(w, err)
			return
		}

		defer file.Close()

		filepath := flag.Lookup("F").Value.String() + user.name + "/" + folderPath + "/" + header.Filename

		out, err := os.Create(filepath)
		if err != nil {
			log.Println(w, "Unable to create the file for writing. Check your write access privilege. Path: "+filepath)
			return
		}

		defer out.Close()

		// write the content from POST to the file
		_, err = io.Copy(out, file)
		if err != nil {
			log.Println(w, err)
		}

		log.Println(w, "File uploaded successfully : ")
		log.Println(w, header.Filename)
	}
}

/*
Ändert ein Passwort. Überprüft das alte Passwort. Überprüft ob beide neuen Passwörter gleich sind. Wenn ja wird es in der CSV-Datei im Server niedergeschrieben.
*/
func changePasswordHandler(w http.ResponseWriter, req *http.Request) {

	log.Println("changePassword")
	cookiecheck, user, cookie := checkCookie(w, req)
	if cookiecheck {
		oldPW := req.FormValue("oldPassword")
		newPW := req.FormValue("newPassword")
		newPWToo := req.FormValue("newPassword2")
		if newPW == newPWToo {
			if AuthenticatorVar.Authenticate(&user, oldPW) {
				changePasswordInFile(&user, newPW)
				log.Println("Set newPW in Cookie: " + user.name)
				cookie.Expires = time.Now().Add(-1)
				cookie.Value = ""
				http.SetCookie(w, &cookie)

				user := loadUser(user.name)
				loginUser(user, w, req)
			} else {
				log.Println("ChangePW: Old PW is not Right")
				http.Redirect(w, req, "?change=oldPwFalse", http.StatusMovedPermanently)
			}
		} else {
			log.Println("ChangePW: PW1 is not PW2")
			http.Redirect(w, req, "?change=pwRepeatFalse", http.StatusMovedPermanently)
		}
	} else {
		http.Redirect(w, req, "/", http.StatusMovedPermanently)
	}
}

/*
Löscht einen User aus der CSV-Datei im Server und legt in daraufhin mit neuem Passwort wieder an
*/
func changePasswordInFile(user *user, newPassword string) {
	input, err := ioutil.ReadFile(flag.Lookup("L").Value.String())
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, user.name) {
			lines[i] = ""
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(flag.Lookup("L").Value.String(), []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
	createUser(user.name, newPassword)
}

func doRequestWithPassword(t *testing.T, url string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	req.SetBasicAuth("Andy", "1234")
	res, err := client.Do(req)
	assert.NoError(t, err)
	return res
}


func createServer(auth AuthenticatorFuncBasic) *httptest.Server {
	return httptest.NewServer(WrapperBasic(auth, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello client")
	}))
}

type Authenticator interface {
	Authenticate(user *user, password string) bool
}

type AuthenticatorFunc func(user *user, password string) bool

type AuthenticatorBasic interface {
	AuthenticateBasic(user, password string) bool
}

type AuthenticatorFuncBasic func(user, password string) bool

func (af AuthenticatorFuncBasic) AuthenticateBasic(user, password string) bool {
	return af(user, password)
}

func WrapperBasic(authenticator AuthenticatorBasic, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pswd, ok := r.BasicAuth()

		if ok && authenticator.AuthenticateBasic(user, pswd) {
			handler(w, r)
		} else {
			w.Header().Set("WWW-Authenticate",
				"Basic realm=\"My Simple Server\"")
			http.Error(w,
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized)
		}
	}
}
