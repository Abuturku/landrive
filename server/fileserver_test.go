package server

import (
	"encoding/csv"
	"flag"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	//"log"
	//"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
	"github.com/stretchr/testify/assert"

)

func generateCookie() http.Cookie {
	cookieValue := hash([]string{"Andy", "a879518e72e3aa6d82126e52d6a641e66005d68b44a31ea5797d0e24f90fd759"})
	maxAge, _ := strconv.Atoi(flag.Lookup("T").Value.String())
	cookie := http.Cookie{Name: "Andy", Value: cookieValue, MaxAge: maxAge, Expires: time.Now().Add(15 * time.Minute)}
	return cookie
}

func init() {
	pathToFile := "./user_test.csv"
	if _, err := os.Stat(pathToFile); err == nil {
		os.Remove(pathToFile)
	}

	file, err := os.Create(pathToFile)

	writer := csv.NewWriter(file)
	defer file.Close()

	writer.Write([]string{"Andy", "a879518e72e3aa6d82126e52d6a641e66005d68b44a31ea5797d0e24f90fd759", "0912951feb016907a1b762c7f83de9b0"})
	writer.Flush()
	err = writer.Error()
	if err != nil {

	}

	os.Mkdir("test", 0777)

	flag.String("L", pathToFile, "Path to file, where usernames, passwords and salts are stored")
	flag.String("T", "900", "Session timeout given in seconds")
	flag.String("F", "test/", "Folder where all Userfiles are stored")

}

//Der Zugang soll durch Benutzernamen und Passwort geschützt werden. Positives Beispiel
func TestAccess(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Andy")
	v.Add("password", "andy")
	req.Form = v

	rr := httptest.NewRecorder()

	loginHandler(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

}

//Der Zugang soll durch Benutzernamen und Passwort geschützt werden. Negatives Beispiel: FalschesPassword
func TestAccessWrongPassword(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Andy")
	v.Add("password", "andy1")
	req.Form = v

	rr := httptest.NewRecorder()

	loginHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

//Der Zugang soll durch Benutzernamen und Passwort geschützt werden. Negatives Beispiel: User exestiert nicht
func TestAccessUserDoesntExist(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Andy1")
	v.Add("password", "andy")
	req.Form = v

	rr := httptest.NewRecorder()

	loginHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

//Zur weiteren Identifikation des Nutzers soll ein Session-ID Cookie verwendet werden.
func TestValidCookie(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Andy")
	v.Add("password", "andy")
	req.Form = v

	rr := httptest.NewRecorder()

	loginHandler(rr, req)

	cookie := generateCookie()

	req.AddCookie(&cookie)

	isValid, _, _ := checkCookie(rr, req)

	if !isValid {
		t.Errorf("Cookie check failed. Expected true got %v", isValid)
	}

}

func TestUnvalidCookie(t *testing.T) {
	req, err := http.NewRequest("POST", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Andy")
	v.Add("password", "andy")
	req.Form = v

	rr := httptest.NewRecorder()

	loginHandler(rr, req)
	cookieValue := hash([]string{"Andy1", "a879518e72e3aa6d82126e52d6a641e66005d68b44a31ea5797d0e24f90fd759"})
	maxAge, _ := strconv.Atoi(flag.Lookup("T").Value.String())
	cookie := http.Cookie{Name: "Andy", Value: cookieValue, MaxAge: maxAge, Expires: time.Now().Add(15 * time.Minute)}

	req.AddCookie(&cookie)

	isValid, _, _ := checkCookie(rr, req)

	if isValid {
		t.Errorf("Cookie check failed. Expected false got %v", isValid)
	}
}

// Neue Nutzer sollen selbst einen Zugang anlegen können.
func TestCreateValidUser(t *testing.T) {
	req, err := http.NewRequest("POST", "/register", nil)

	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Niklas")
	v.Add("password", "niklas")
	v.Add("password2", "niklas")
	req.Form = v

	rr := httptest.NewRecorder()

	newUserHandler(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
}

func TestCreateUserPwFalse(t *testing.T) {
	req, err := http.NewRequest("POST", "/register", nil)

	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Niklas")
	v.Add("password", "niklas")
	v.Add("password2", "niklas1")
	req.Form = v

	rr := httptest.NewRecorder()

	newUserHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}

func TestCreateUserNameFalse(t *testing.T) {
	req, err := http.NewRequest("POST", "/register", nil)

	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("username", "Andy")
	v.Add("password", "niklas")
	v.Add("password2", "niklas")
	req.Form = v

	rr := httptest.NewRecorder()

	newUserHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}


// An https://golang.org/src/net/http/request_test.go orientiert
const testMessage =  `
 --MyBoundary
 Content-Disposition: form-data; name="uploadFile"; filename="filea.txt"
 Content-Type: text/plain
 This is a test file.
 --MyBoundary
 Content-Disposition: form-data; name="text"
 foo
 --MyBoundary--`

// Es soll möglich sein, Dateien ”hochzuladen“
func TestSaveFile(t *testing.T){
	postData := strings.NewReader(strings.Replace(testMessage, "\n", "\r\n", -1))

	req, err := http.NewRequest("POST", "/uploadFile", postData)

	if err != nil {

		t.Fatal("NewRequest:", err)

	}

	cookie := generateCookie()
	req.AddCookie(&cookie)
	req.PostForm = url.Values{}
	req.PostForm.Add("path", "")
	req.Header.Set("Content-type", `multipart/form-data; boundary="MyBoundary"`)

	rr := httptest.NewRecorder()
	,,c := req.FormFile("uploadFile")

	t.Log(c)

	uploadFileHandler(rr, req)

	if _, err := os.Stat(flag.Lookup("F").Value.String() + cookie.Name +"/filea.txt"); os.IsNotExist(err) {
		t.Error("Error while saving")
	}
}

// Es soll möglich sein, Dateien ”herunterzuladen“
func TestDownloadFile(t *testing.T) {
	req, err := http.NewRequest("POST", "/download", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	v := url.Values{}
	v.Add("path", "user_test.csv")
	req.Form = v

	rr := httptest.NewRecorder()

	downloadHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestDownloadFileNotLoggedIn(t *testing.T) {
	req, err := http.NewRequest("POST", "/download", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	downloadHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}

// Es soll möglich sein, Dateien ”herunterzuladen“ über wget
func TestDownloadFileWGETValidFile(t *testing.T) {
	req, err := http.NewRequest("GET", "/wget?path=./user_test.csv", nil)

	if err != nil {
		t.Fatal(err)
	}

	req.SetBasicAuth("Andy", "andy")

	rr := httptest.NewRecorder()
	wgetHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestDownloadFileWGETUnvalidFile(t *testing.T) {
	req, err := http.NewRequest("GET", "/wget?path=./user_test1.csv", nil)

	if err != nil {
		t.Fatal(err)
	}

	req.SetBasicAuth("Andy", "andy")

	rr := httptest.NewRecorder()
	wgetHandler(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestDownloadFileWGETUnvalidAccount(t *testing.T) {
	req, err := http.NewRequest("GET", "/wget?path=./user_test.csv", nil)

	if err != nil {
		t.Fatal(err)
	}

	req.SetBasicAuth("Andy", "andy1")

	rr := httptest.NewRecorder()
	wgetHandler(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestWithoutPW(t *testing.T) {
	ts := createServer(func(user, pwd string) bool {
		return true
	})
	defer ts.Close()
	res, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "wrong status")
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t,
		http.StatusText(http.StatusUnauthorized)+"\n",
		string(body), "wrong message")
}

func TestWithWrongPWBasicAuth(t *testing.T) {
	var receivedName, receivedPw string
	ts := createServer(func(user, pwd string) bool {
		receivedName = user
		receivedPw = pwd
		return false // <--- deny every request
	})
	defer ts.Close()
	res := doRequestWithPassword(t, ts.URL)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "wrong status")
	assert.Equal(t, "Andy", receivedName, "wrong username")
	assert.Equal(t, "1234", receivedPw, "wrong password")
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t,
		http.StatusText(http.StatusUnauthorized)+"\n",
		string(body), "wrong message")
}

func TestWithCorrectPW(t *testing.T) {
	var receivedName, receivedPwd string
	ts := createServer(func(user, pwd string) bool {
		receivedName = user
		receivedPwd = pwd
		return true // <--- accept every request
	})
	defer ts.Close()
	res := doRequestWithPassword(t, ts.URL)
	assert.Equal(t, http.StatusOK, res.StatusCode, "wrong status code")
	assert.Equal(t, "Andy", receivedName, "wrong username")
	assert.Equal(t, "1234", receivedPwd, "wrong password")
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Hello client\n", string(body), "wrong message")
}

//Auch in diese Unterordner sollen sich Dateien laden lassen.
func TestCreateFolder(t *testing.T) {
	req, err := http.NewRequest("POST", "/newFolder", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	v := url.Values{}
	v.Add("path", "")
	v.Add("newFolderName", "testFolder")
	req.Form = v

	rr := httptest.NewRecorder()

	createFolderHandler(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
}

func TestCreateFolderNotLoggedIn(t *testing.T) {
	req, err := http.NewRequest("POST", "/newFolder", nil)

	if err != nil {
		t.Fatal(err)
	}


	rr := httptest.NewRecorder()

	createFolderHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}

//Es möglich sein, Ordner zu löschen.
func TestDeleteFolder(t *testing.T) {
	req, err := http.NewRequest("POST", "/delete", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	v := url.Values{}
	v.Add("path", "testFolder")
	req.Form = v

	rr := httptest.NewRecorder()

	deleteHandler(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
}

func TestDeleteFolderNotLoggedIn(t *testing.T) {
	req, err := http.NewRequest("POST", "/delete", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	deleteHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}

func TestDeleteFolderPathEmpty(t *testing.T) {
	req, err := http.NewRequest("POST", "/delete", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	v := url.Values{}
	v.Add("path", "")
	req.Form = v

	rr := httptest.NewRecorder()

	deleteHandler(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
}

func TestIndexHandlerLoggedIn(t *testing.T){
	req, err := http.NewRequest("GET", "/", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	rr := httptest.NewRecorder()
	index(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}

func TestFolderStructHandlerLoggedIn(t *testing.T){
	req, err := http.NewRequest("GET", "/getFolderStruct", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	rr := httptest.NewRecorder()
	folderStructHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestFolderStructHandlerNotLoggedIn(t *testing.T){
	req, err := http.NewRequest("GET", "/getFolderStruct", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	folderStructHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}

func TestLogoutHandlerLoggedIn(t *testing.T){
	req, err := http.NewRequest("GET", "/logout", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	rr := httptest.NewRecorder()
	logoutHandler(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}

func TestLogoutHandlerNotLoggedIn(t *testing.T){
	req, err := http.NewRequest("GET", "/logout", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	logoutHandler(rr, req)

	if status := rr.Code; status != http.StatusNotModified {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotModified)
	}
}


// Ein Nutzer soll sein Passwort ändern können.
func TestChangePasswordValid(t *testing.T) {
	req, err := http.NewRequest("POST", "/changePw", nil)

	if err != nil {
		t.Fatal(err)
	}

	v := url.Values{}
	v.Add("oldPassword", "andy")
	v.Add("newPassword", "niklas")
	v.Add("newPassword2", "niklas")
	req.Form = v

	cookie := generateCookie()
	req.AddCookie(&cookie)

	rr := httptest.NewRecorder()

	changePasswordHandler(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
}

func TestLandriveHandlerLoggedIn(t *testing.T){
	req, err := http.NewRequest("GET", "/landrive", nil)

	if err != nil {
		t.Fatal(err)
	}

	cookie := generateCookie()
	req.AddCookie(&cookie)

	rr := httptest.NewRecorder()
	landrive(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMovedPermanently)
	}
}