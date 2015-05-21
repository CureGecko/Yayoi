/*
Uploads.go
Yayoi

Created by Cure Gecko on 5/17/15.
Copyright 2015, Cure Gecko. All rights reserved.

Upload system.
*/

package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gographics/imagick/imagick"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var allowedExtensions []string = []string{"tif", "tiff", "gif", "jpeg", "jpg", "png", "bmp", "svg", "webp"}

func extneionAllowed(e string) bool {
	for _, a := range allowedExtensions {
		if a == e {
			return true
		}
	}
	return false
}

type Uploads struct {
	Server  *Iori
	Auth    *Auth
	Writer  http.ResponseWriter
	Request *http.Request
	Path    []string
}

func (u Uploads) Process() {
	if !u.Auth.Authenticated {
		t, _ := template.ParseFiles("resources/Upload/index.html")
		t.Execute(u.Writer, true)
		return
	}
	if len(u.Path) == 1 {
		t, _ := template.ParseFiles("resources/Upload/index.html")
		t.Execute(u.Writer, false)
	} else {
		switch u.Path[1] {
		case "tags":
			u.Tags()
		case "add":
			u.Add()
		case "view":
			u.View()
		case "submit":
			u.Submit()
		default:
			fmt.Fprint(u.Writer, "404 Not Found\n")
		}
	}
}

func (u Uploads) Tags() {
	type Info struct {
		Success bool
		Tags    string
		Source  string
	}
	info := Info{true, "", ""}
	var response map[string]interface{}

	MD5 := u.Request.Form.Get("MD5")
	if MD5 == "" {
		info.Success = false
	} else {
		resp, err := http.Get("https://cure.ninja/booru/api/json/md5/" + MD5)
		var bytes []byte
		if err != nil {
			log.Println(err)
			info.Success = false
		} else {
			bytes, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Println(err)
				info.Success = false
			} else {
				err = json.Unmarshal(bytes, &response)
				if err != nil {
					log.Println(err)
					info.Success = false
				}
			}
		}
		if info.Success {
			if response["success"].(bool) && len(response["results"].([]interface{})) != 0 {
				info.Tags = response["results"].([]interface{})[0].(map[string]interface{})["tag"].(string)
				source, ok := response["results"].([]interface{})[0].(map[string]interface{})["sourceURL"].(string)
				if source == "" || !ok {
					info.Source = response["results"].([]interface{})[0].(map[string]interface{})["page"].(string)
				} else {
					info.Source = source
				}
			}
		}
	}

	t, _ := template.ParseFiles("resources/Upload/tags.html")
	t.Execute(u.Writer, info)
}

func (u Uploads) Add() {
	type Info struct {
		Success bool
		Reason  string
	}
	info := Info{true, ""}

	fileName := u.Request.Header.Get("fileName")
	log.Println(fileName)
	var extension string
	if len(fileName) != 0 && strings.LastIndex(fileName, ".") != -1 {
		extension = strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	}
	contentLength := u.Request.ContentLength
	if len(fileName) == 0 || contentLength == 0 || u.Request.Method != "POST" {
		info.Success = false
		info.Reason = "Invalid Request"
	}
	if info.Success && !extneionAllowed(extension) {
		info.Success = false
		info.Reason = "The extension " + extension + " is not allowed."
	}
	uploadFolder := "uploadTmp/" + strconv.FormatUint(u.Auth.User.Id, 10) + "/"
	uploadFile := uploadFolder + "upload." + extension
	if info.Success {
		err := os.MkdirAll(uploadFolder+"thumb/", 0755)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "OS Error"
		}
	}
	if info.Success {
		err := os.MkdirAll(uploadFolder+"iqdb/", 0755)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "OS Error"
		}
	}

	var SHA1 string
	var SHA256 string
	var SHA512 string
	var MD5 string
	if info.Success {
		file, err := os.Create(uploadFile)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "OS Error"
		} else {
			hasher1 := sha1.New()
			hasher256 := sha256.New()
			hasher512 := sha512.New()
			hasher5 := md5.New()
			for {
				bytes := make([]byte, 4096)
				read, err := u.Request.Body.Read(bytes)
				if err != nil && err != io.EOF {
					log.Println(err)
					info.Success = false
					info.Reason = "Error reading from browser."
					break
				}
				if read == 0 || err == io.EOF {
					break
				}
				readBytes := bytes[0:read]
				hasher1.Write(readBytes)
				hasher256.Write(readBytes)
				hasher512.Write(readBytes)
				hasher5.Write(readBytes)
				file.Write(readBytes)
			}
			SHA1 = hex.EncodeToString(hasher1.Sum(nil))
			SHA256 = hex.EncodeToString(hasher256.Sum(nil))
			SHA512 = hex.EncodeToString(hasher512.Sum(nil))
			MD5 = hex.EncodeToString(hasher5.Sum(nil))
			log.Println("SHA1:", SHA1)
			log.Println("SHA256:", SHA256)
			log.Println("SHA512:", SHA512)
			log.Println("MD5:", MD5)
			file.Close()
			err := os.Rename(uploadFile, uploadFolder+SHA256+"."+extension)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Reason = "OS Error"
			}
			uploadFile = uploadFolder + SHA256 + "." + extension
		}
	}

	if info.Success {
		upload, err := u.Server.DBmap.Get(Upload{}, SHA256)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "Error querying the database."
		}
		if upload != nil {
			info.Success = false
			info.Reason = "This image is already being uploaded."
		}
	}

	var width uint
	var height uint
	var thumnailWidth uint
	var thumnailHeight uint
	thumbnailExtension := extension
	if info.Success {
		mw := imagick.NewMagickWand()
		err := mw.ReadImage(uploadFile)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "Error reading image."
		} else {
			width = mw.GetImageWidth()
			height = mw.GetImageHeight()
			log.Println(width, height)

			newWidth := float64(width)
			newHeight := float64(height)
			if width > thumbnailSize || height > thumbnailSize {
				widthFactor := thumbnailSize / float64(width)
				heightFactor := thumbnailSize / float64(height)
				scaleFactor := 1.0

				if widthFactor < heightFactor {
					scaleFactor = widthFactor
				} else {
					scaleFactor = heightFactor
				}

				newWidth = math.Floor((float64(width) * scaleFactor) + 0.5)
				newHeight = math.Floor((float64(height) * scaleFactor) + 0.5)
			}
			thumnailWidth = uint(newWidth)
			thumnailHeight = uint(newHeight)
			log.Println(newWidth, newHeight)
			mw.ResizeImage(uint(newWidth), uint(newHeight), imagick.FILTER_LANCZOS2, 1)
			mw.SetImageColorspace(imagick.COLORSPACE_SRGB)

			if extension == "bmp" || extension == "webp" || extension == "svg" {
				thumbnailExtension = "jpg"
			}
			err := mw.WriteImage(uploadFolder + "thumb/" + SHA256 + "." + thumbnailExtension)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Reason = "Error saving thumbnail."
			}
		}
		mw.Destroy()
	}

	var fileSize int64
	var thumbnailFileSize int64
	if info.Success {
		fInfo, err := os.Stat(uploadFolder + SHA256 + "." + extension)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "Error retreving file info."
		} else {
			fileSize = fInfo.Size()
		}
		fInfo, err = os.Stat(uploadFolder + "thumb/" + SHA256 + "." + thumbnailExtension)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "Error retreving file info."
		} else {
			thumbnailFileSize = fInfo.Size()
		}
	}

	if info.Success {
		mw := imagick.NewMagickWand()
		err := mw.ReadImage(uploadFile)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "Error reading image."
		} else {
			mw.ResizeImage(128, 128, imagick.FILTER_LANCZOS2, 1)
			mw.SetImageColorspace(imagick.COLORSPACE_SRGB)

			iqdbFile := uploadFolder + "iqdb/" + SHA256 + ".png"

			err := mw.WriteImage(iqdbFile)
			if err != nil {
				log.Println(err)
				info.Success = false
				info.Reason = "Error saving IQDB thumb."
			}
		}
		mw.Destroy()
	}

	if info.Success {
		upload := new(Upload)
		upload.UserID = u.Auth.User.Id
		upload.MD5 = MD5
		upload.SHA1 = SHA1
		upload.SHA256 = SHA256
		upload.SHA512 = SHA512
		upload.Extension = extension
		upload.FileSize = fileSize
		upload.Width = int(width)
		upload.Height = int(height)
		upload.ThumbnailExtension = thumbnailExtension
		upload.ThumbnailFileSize = thumbnailFileSize
		upload.ThumnailWidth = int(thumnailWidth)
		upload.ThumnailHeight = int(thumnailHeight)
		upload.Time = time.Now().Unix()
		err := u.Server.DBmap.Insert(upload)
		if err != nil {
			log.Println(err)
			info.Success = false
			info.Reason = "Unable to add to database."
		}
	}

	u.Request.Body.Close()
	t, _ := template.ParseFiles("resources/Upload/add.html")
	t.Execute(u.Writer, info)
}

func (u Uploads) View() {
	SHA256 := u.Request.Form.Get("SHA256")

	obj, _ := u.Server.DBmap.Get(Upload{}, SHA256)

	if obj == nil {
		fmt.Fprint(u.Writer, "404 Not Found")
		return
	}
	upload := obj.(*Upload)

	http.ServeFile(u.Writer, u.Request, "uploadTmp/"+strconv.FormatUint(u.Auth.User.Id, 10)+"/"+upload.SHA256+"."+upload.Extension)
}

func (u Uploads) Submit() {
	type Info struct {
		Number int
		Image  string
		Upload *Upload
	}
	var info []Info

	results, _ := u.Server.DBmap.Select(Upload{}, "SELECT * FROM uploads WHERE UserID=?", u.Auth.User.Id)

	for i, result := range results {
		upload := result.(*Upload)
		thisInfo := Info{}
		thisInfo.Number = i
		thisInfo.Image = APIPath + "uploads/view?SHA256=" + upload.SHA256
		thisInfo.Upload = upload
		info = append(info, thisInfo)
	}

	t, _ := template.ParseFiles("resources/Upload/submit.html")
	t.Execute(u.Writer, info)
}
