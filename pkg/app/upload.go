package app

import (
	"bytes"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"starter/pkg/unique"
	"strconv"
	"strings"
	"sync"
)

type fileTypeMap struct {
	TypeMap sync.Map
	sync.Once
}

type SaveHandler interface {
	// 保存文件并返回文件最终路径
	Save(file *multipart.FileHeader, fileName string) string
}

// 默认文件保存器
type DefaultSaveHandler struct {
	prefix  string
	dst     string
	context *gin.Context
}

func (defaultSaveHandler *DefaultSaveHandler) SetDst(dst string) *DefaultSaveHandler {
	defaultSaveHandler.dst = dst
	return defaultSaveHandler
}

func (defaultSaveHandler *DefaultSaveHandler) SetPrefix(prefix string) *DefaultSaveHandler {
	defaultSaveHandler.prefix = prefix
	return defaultSaveHandler
}

func (defaultSaveHandler *DefaultSaveHandler) Save(file *multipart.FileHeader, fileName string) string {
	filePath := defaultSaveHandler.dst + fileName
	err := defaultSaveHandler.context.SaveUploadedFile(file, filePath)
	if err != nil {
		Get("logger").(*logrus.Logger).Println(err)
		return ""
	} else {
		return defaultSaveHandler.prefix + filePath
	}
}

var FileType = new(fileTypeMap)

func (fileTypeMap *fileTypeMap) lazyLoad() {
	fileTypeMap.Do(func() {
		fileTypeMap.TypeMap.Store("ffd8ffe", "jpg")                     // JPEG (jpg)
		fileTypeMap.TypeMap.Store("89504e470d0a1a0a0000", "png")        // PNG (png)
		fileTypeMap.TypeMap.Store("47494638396126026f01", "gif")        // GIF (gif)
		fileTypeMap.TypeMap.Store("49492a00227105008037", "tif")        // TIFF (tif)
		fileTypeMap.TypeMap.Store("424d228c010000000000", "bmp")        // 16色位图(bmp)
		fileTypeMap.TypeMap.Store("424d8240090000000000", "bmp")        // 24位位图(bmp)
		fileTypeMap.TypeMap.Store("424d8e1b030000000000", "bmp")        // 256色位图(bmp)
		fileTypeMap.TypeMap.Store("41433130313500000000", "dwg")        // CAD (dwg)
		fileTypeMap.TypeMap.Store("3c21444f435459504520", "html")       // HTML (html)   3c68746d6c3e0  3c68746d6c3e0
		fileTypeMap.TypeMap.Store("3c68746d6c3e0", "html")              // HTML (html)   3c68746d6c3e0  3c68746d6c3e0
		fileTypeMap.TypeMap.Store("3c21646f637479706520", "htm")        // HTM (htm)
		fileTypeMap.TypeMap.Store("48544d4c207b0d0a0942", "css")        // css
		fileTypeMap.TypeMap.Store("696b2e71623d696b2e71", "js")         // js
		fileTypeMap.TypeMap.Store("7b5c727466315c616e73", "rtf")        // Rich Text Format (rtf)
		fileTypeMap.TypeMap.Store("38425053000100000000", "psd")        // Photoshop (psd)
		fileTypeMap.TypeMap.Store("46726f6d3a203d3f6762", "eml")        // Email [Outlook Express 6] (eml)
		fileTypeMap.TypeMap.Store("d0cf11e0a1b11ae10000", "doc")        // MS Excel 注意：word、msi 和 excel的文件头一样
		fileTypeMap.TypeMap.Store("d0cf11e0a1b11ae10000", "vsd")        // Visio 绘图
		fileTypeMap.TypeMap.Store("5374616E64617264204A", "mdb")        // MS Access (mdb)
		fileTypeMap.TypeMap.Store("252150532D41646F6265", "ps")         //
		fileTypeMap.TypeMap.Store("255044462d312e350d0a", "pdf")        // Adobe Acrobat (pdf)
		fileTypeMap.TypeMap.Store("2e524d46000000120001", "rmvb")       // rmvb/rm相同
		fileTypeMap.TypeMap.Store("464c5601050000000900", "flv")        // flv与f4v相同
		fileTypeMap.TypeMap.Store("00000020667479706d70", "mp4")        //
		fileTypeMap.TypeMap.Store("49443303000000002176", "mp3")        //
		fileTypeMap.TypeMap.Store("000001ba210001000180", "mpg")        //
		fileTypeMap.TypeMap.Store("3026b2758e66cf11a6d9", "wmv")        // wmv与asf相同
		fileTypeMap.TypeMap.Store("52494646e27807005741", "wav")        // Wave (wav)
		fileTypeMap.TypeMap.Store("52494646d07d60074156", "avi")        //
		fileTypeMap.TypeMap.Store("4d546864000000060001", "mid")        // MIDI (mid)
		fileTypeMap.TypeMap.Store("504b0304140000000800", "zip")        //
		fileTypeMap.TypeMap.Store("526172211a0700cf9073", "rar")        //
		fileTypeMap.TypeMap.Store("235468697320636f6e66", "ini")        //
		fileTypeMap.TypeMap.Store("504b03040a0000000000", "jar")        //
		fileTypeMap.TypeMap.Store("4d5a9000030000000400", "exe")        // 可执行文件
		fileTypeMap.TypeMap.Store("3c25402070616765206c", "jsp")        // jsp文件
		fileTypeMap.TypeMap.Store("4d616e69666573742d56", "mf")         // MF文件
		fileTypeMap.TypeMap.Store("3c3f786d6c2076657273", "xml")        // xml文件
		fileTypeMap.TypeMap.Store("494e5345525420494e54", "sql")        // xml文件
		fileTypeMap.TypeMap.Store("7061636b616765207765", "java")       // java文件
		fileTypeMap.TypeMap.Store("406563686f206f66660d", "bat")        // bat文件
		fileTypeMap.TypeMap.Store("1f8b0800000000000000", "gz")         // gz文件
		fileTypeMap.TypeMap.Store("6c6f67346a2e726f6f74", "properties") // bat文件
		fileTypeMap.TypeMap.Store("cafebabe0000002e0041", "class")      // bat文件
		fileTypeMap.TypeMap.Store("49545346030000006000", "chm")        // bat文件
		fileTypeMap.TypeMap.Store("04000000010000001300", "mxp")        // bat文件
		fileTypeMap.TypeMap.Store("504b0304140006000800", "docx")       // docx文件
		fileTypeMap.TypeMap.Store("d0cf11e0a1b11ae10000", "wps")        // WPS文字wps、表格et、演示dps都是一样的
		fileTypeMap.TypeMap.Store("6431303a637265617465", "torrent")    //
		fileTypeMap.TypeMap.Store("6D6F6F76", "mov")                    // Quicktime (mov)
		fileTypeMap.TypeMap.Store("FF575043", "wpd")                    // WordPerfect (wpd)
		fileTypeMap.TypeMap.Store("CFAD12FEC5FD746F", "dbx")            // Outlook Express (dbx)
		fileTypeMap.TypeMap.Store("2142444E", "pst")                    // Outlook (pst)
		fileTypeMap.TypeMap.Store("AC9EBD8F", "qdf")                    // Quicken (qdf)
		fileTypeMap.TypeMap.Store("E3828596", "pwl")                    // Windows Password (pwl)
		fileTypeMap.TypeMap.Store("2E7261FD", "ram")                    // Real Audio (ram)
	})
}

// 获取前面结果字节的二进制
func (fileTypeMap *fileTypeMap) bytesToHexString(src []byte) string {
	res := bytes.Buffer{}
	if src == nil || len(src) <= 0 {
		return ""
	}
	temp := make([]byte, 0)
	for _, v := range src {
		sub := v & 0xFF
		hv := hex.EncodeToString(append(temp, sub))
		if len(hv) < 2 {
			res.WriteString(strconv.FormatInt(int64(0), 10))
		}
		res.WriteString(hv)
	}
	return res.String()
}

// 用文件前面几个字节来判断
// fSrc: 文件字节流（就用前面几个字节）
func (fileTypeMap *fileTypeMap) GetFileType(fSrc []byte) string {
	fileTypeMap.lazyLoad()
	var fileType string
	fileCode := fileTypeMap.bytesToHexString(fSrc)
	fileTypeMap.TypeMap.Range(func(key, value interface{}) bool {
		k := key.(string)
		v := value.(string)
		if strings.HasPrefix(fileCode, strings.ToLower(k)) ||
			strings.HasPrefix(k, strings.ToLower(fileCode)) {
			fileType = v
			return false
		}
		return true
	})
	return fileType
}

// 文件上传公共方法
//  key 上传文件的表单name, 如果是多文件需要加上中括号[]
//  dst 存放路径 注意:无论这里传什么路径, 最后边都会追加 filename.xxx
func Upload(key string, saveHandler SaveHandler, allowedTyp ...string) gin.HandlerFunc {
	return func(context *gin.Context) {
		form, _ := context.MultipartForm()
		files := form.File[key]
		formKey := context.PostForm("key")

		var response = NewResponse(Success, nil, "SUCCESS")
		var data = make(map[string][]string, 0)
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				data[formKey] = make([]string, 0, 0)
			} else {
				byt, err := ioutil.ReadAll(f)
				if err != nil {
					data[formKey] = make([]string, 0, 0)
				} else {
					fileType := FileType.GetFileType(byt[:10])
					var typAllow = false
					for _, typ := range allowedTyp {
						typAllow = typAllow || typ == fileType
					}
					if typAllow {
						fileName := strconv.Itoa(int(unique.Id())) + "." + fileType
						data[formKey] = append(data[formKey], saveHandler.Save(file, fileName))
						//err := context.SaveUploadedFile(file, filePath)
						//if err != nil {
						//	log.Println(err)
						//} else {
						//	data[formKey] = append(data[formKey], prefix+filePath)
						//}
					}

				}
				_ = f.Close()
			}

		}

		response.Data = data
		context.JSON(http.StatusOK, response)
	}
}
