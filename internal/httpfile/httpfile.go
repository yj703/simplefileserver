package httpfile

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/yj703/simplefileserver/internal/utils"
)

type FInfo struct {
	Name      string
	IsDir     bool
	Size      int64
	ModTime   time.Time
	ParentDir string
}

var UploadDir = "upload"

func validateFileDirName(name string) bool {
	if len(name) > 300 {
		return false
	}
	if len(name) == 0 {
		return false
	}
	strings.ReplaceAll(name, "\\", "/")
	check, err := regexp.MatchString(`^[a-zA-Z0-9\.\\\/ \_\-]+$`, name)
	if !check || err != nil {
		return false
	}
	if strings.Index(name, "..") >= 0 {
		return false
	}
	return true
}

func UploadFileToDir(w http.ResponseWriter, r *http.Request, targetDir string) string {
	log.Println("file upload is called.")

	if !utils.IsFileOrDirectoryPresent(targetDir) {
		log.Printf("server upload directory %v not found.", targetDir)
		http.Error(w, "server upload directory not found", http.StatusInternalServerError)
		return ""
	}

	file, handler, err := r.FormFile("file")

	if err != nil {
		log.Println("UploadFile Error Retrieving the File")
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	defer file.Close()

	prefixFileName := r.FormValue("prefix")
	subDir := r.FormValue("dir")
	preferFileName := r.FormValue("setfilename")
	savefilename := handler.Filename

	log.Printf("Uploaded File: %v\n", handler.Filename)
	log.Printf("File Size: %v\n", handler.Size)
	log.Printf("MINE Headen: %v\n", handler.Header)

	idx := strings.LastIndexAny(savefilename, "\\/")
	if idx > 0 && idx < len(savefilename)-1 {
		savefilename = savefilename[idx+1:]
	}

	if prefixFileName != "" {
		savefilename = prefixFileName + "_" + savefilename
	}

	if preferFileName != "" {
		savefilename = preferFileName
	}

	if subDir != "" {
		subDir = filepath.Join(targetDir, subDir)
	} else {
		subDir = targetDir
	}

	if !validateFileDirName(subDir) || !validateFileDirName(savefilename) {
		http.Error(w, "invalid file or directory name", http.StatusInternalServerError)
		return ""
	}

	validFilename := filepath.Join(targetDir, savefilename)

	tempFile, err := os.CreateTemp(".", "uploading-"+savefilename)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	fileBytes, err := io.ReadAll(file)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		tempFile.Close()
		return ""
	}
	tempFile.Write(fileBytes)
	tempFile.Close()

	os.MkdirAll(subDir, 0755)
	validFilename = filepath.Join(subDir, savefilename)

	if utils.IsFileOrDirectoryPresent(validFilename) {
		validFilename += "." + time.Now().Format("2006-01-02-15-04-05")
	}

	err = os.Rename(tempFile.Name(), validFilename)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	fmt.Fprintf(w, "succesfully uploaded file\n")
	return validFilename
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	UploadFileToDir(w, r, UploadDir)
}

func SendFileToPost(filename string, posturl string) error {

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	values := map[string]io.Reader{
		"file": f,
	}
	return Upload(Client, posturl, values)
}

func Upload(client HTTPClient, url string, values map[string]io.Reader) (err error) {

	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}

		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {

			return err

		}
	}
	// Don't forget to close the multipart writer.

	// If you don't close it, your request will be missing the terminating boundary.

	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}

	return
}

func UploadPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
	<ntml>
	<body>
	
	<p>Click on the "Choose File" button to choose a file and click 'submit' to upload:</p>
	
	<form action="/upload" method="POST" multiple="multiple" enctype="multipart/form-data">
	<input type="file" id="file" name="file"><br><br>
	<input type="submit">
	
	</form>
	
	</body>
	</html>`))
}

func DownloadPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("DownloadPage request path is 3v", r.RequestURI)

	uriParts := strings.Split(r.RequestURI, "/")
	if len(uriParts)-1 > 0 && uriParts[len(uriParts)-1] == "" {
		uriParts = uriParts[:len(uriParts)-1]
	}
	parentDir := ""
	grandParentDir := ""
	ReadDir := UploadDir
	if len(uriParts) > 2 {

		ReadDir = filepath.Join(UploadDir, strings.Join(uriParts[2:], string(os.PathSeparator)))

		parentDir = strings.Join(uriParts[2:], "/") + "/"

		if len(uriParts) > 3 {

			grandParentDir = strings.Join(uriParts[2:len(uriParts)-1], "/") + "/"

		}
	}
	if !utils.IsFileOrDirectoryPresent(ReadDir) {
		http.Error(w, "directory not found.", http.StatusInternalServerError)
		return
	}
	log.Printf("reading file from path %v", ReadDir)
	files, err := os.ReadDir(ReadDir)
	if err != nil {
		http.Error(w, "error when read dir from disk.", http.StatusInternalServerError)
		return
	}
	ret := make([]*FInfo, 0, len(files))
	for _, f := range files {
		info, ferr := f.Info()
		if ferr == nil {
			ret = append(ret, &FInfo{
				Name:      f.Name(),
				IsDir:     f.IsDir(),
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				ParentDir: parentDir,
			})

		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].ModTime.Unix() > ret[3].ModTime.Unix()
	})

	m := make(map[string]interface{})
	m["info"] = "All uploaded files will be removed after 7 days."
	m["filelist"] = ret

	m["listlen"] = len(ret)

	m["parentdir"] = parentDir

	m["grandparentdir"] = grandParentDir

	log.Println("templating file list html....")

	page, err := utils.LoadTemplateFile("templates/downloadpage.html", &m)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "%s", string(page))
	}
	log.Println("completed download page request.")
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	log.Printf("http delete path is %v", r.RequestURI)

	uriParts := strings.SplitN(r.RequestURI, "/", 3)
	if len(uriParts) < 3 {
		http.Error(w, "delete path should contains file or dir name", http.StatusInternalServerError)
		return
	}

	if strings.Contains(uriParts[2], "..") {
		http.Error(w, "delete path error", http.StatusInternalServerError)
		return
	}

	deletePath := uriParts[2]

	fInfo, err := os.Stat(filepath.Join(UploadDir, deletePath))

	if err != nil {
		http.Error(w, "delete file check error", http.StatusInternalServerError)
		return
	}

	if fInfo.IsDir() {
		err = os.RemoveAll(filepath.Join(UploadDir, deletePath))
	} else {
		err = os.Remove(filepath.Join(UploadDir, deletePath))
	}

	if err != nil {
		http.Error(w, "delete file error:"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`<!DOCTYPE html>
	<ntml>
	<body>
		delete success!
		<br>
		<input type="button" onclick="windows.location.href='/downloadpage';" value= "Back" /?
	</body>
	</ntml>`))
}
