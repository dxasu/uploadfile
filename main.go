package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
)

var filePath string = ""
var myhost string = ""

const myport = "9999"

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("uploadHandler request time:%s host:%s header:%v \n", time.Now().Format("2006-01-02 15:04:05"), r.URL.Host, r.Header)
	err := r.ParseMultipartForm(1024 << 20) // 限制上传文件大小为1024MB
	if err != nil {
		fmt.Println("ParseMultipartForm failed err:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["file"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			fmt.Println("FileHeader.Open failed err:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		dst, err := os.Create(path.Join(filePath, fileHeader.Filename))
		if err != nil {
			fmt.Println("os.Create failed err:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			fmt.Println("os.Copy failed err:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("文件 %s 上传成功\n", fileHeader.Filename)
	}
	fmt.Fprintln(w, "所有文件上传成功")
	fmt.Println("All Success on time:", time.Now().Format("2006-01-02 15:04:05"))
}

func webIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("webIndex request time:%s host:%s header:%v \n", time.Now().Format("2006-01-02 15:04:05"), r.URL.Host, r.Header)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 将 HTML 内容写入响应体
	_, err := w.Write([]byte(htmlText))
	if err != nil {
		log.Println("Error writing response:", err)
	}

}

func showQrcode(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("showQrcode request time:%s host:%s header:%v \n", time.Now().Format("2006-01-02 15:04:05"), r.URL.Host, r.Header)

	qrBytes, err := qrcode.Encode(myhost+"/index", qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	qrImg, _, err := image.Decode(bytes.NewReader(qrBytes))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")

	err = png.Encode(w, qrImg)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func getLocalIp() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("获取网络接口错误:", err)
		return "localhost"
	}

	// 遍历网络接口，找到非 loopback 和非虚拟网卡的第一个 IPv4 地址
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("获取网络地址错误:", err)
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			if ipNet.IP.To4() != nil && strings.HasPrefix(ipNet.IP.String(), "192.168") {
				return ipNet.IP.String()
			}
		}
	}
	return "localhost"
}

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置允许跨域访问的来源，* 表示允许任意来源
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		log.Fatalln("Error opening browser:", err)
	}
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println(`empty filepath, command as:
uploadfile filepath`)
		return
	}
	myhost = "http://" + getLocalIp() + ":" + myport
	filePath = os.Args[1]
	http.HandleFunc("/", showQrcode)
	http.HandleFunc("/index", webIndex)
	http.HandleFunc("/upload", uploadHandler)
	fmt.Println("click link and upload with qrcode " + myhost)
	go func() {
		time.Sleep(time.Second)
		go openBrowser(myhost)
	}()
	http.ListenAndServe(":"+myport, Cors(http.DefaultServeMux))
}

var htmlText = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文件上传</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
        }

        .container {
            background-color: #ffffff;
            border-radius: 10px;
            box-shadow: 0px 5px 15px rgba(0, 0, 0, 0.1);
            padding: 30px;
            max-width: 500px;
            text-align: center;
        }

        h1 {
            font-size: 24px;
            margin-bottom: 20px;
        }

        .upload-btn {
            border: none;
            background: linear-gradient(to right, #ff7e5f, #feb47b);
            color: white;
            padding: 15px 30px;
            text-align: center;
            text-decoration: none;
            display: inline-block;
            font-size: 18px;
            border-radius: 8px;
            cursor: pointer;
            transition: background 0.3s ease;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        }

        .upload-btn:hover {
            background: linear-gradient(to right, #fe4a5a, #ff7e5f);
        }

        .loading-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.5);
            z-index: 9999;
            justify-content: center;
            align-items: center;
        }

        .loading-text {
            color: white;
            font-size: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>选择要上传的文件</h1>
        <input id="file-upload" type="file" name="file" accept=".jpg, .jpeg, .png, .gif" multiple style="display: none;">
        <button id="upload-btn" class="upload-btn">选择文件</button>
        <div id="loading-overlay" class="loading-overlay">
            <div class="loading-text">正在上传中...</div>
        </div>
    </div>

    <script>
        document.getElementById('upload-btn').addEventListener('click', function() {
            document.getElementById('file-upload').click();
        });

        document.getElementById('file-upload').addEventListener('change', function() {
            var fileInput = document.getElementById('file-upload');
            var files = fileInput.files;
            if (files.length > 0) {
                uploadFiles(files);
            }
        });

        function uploadFiles(files) {
            var formData = new FormData();
            for (var i = 0; i < files.length; i++) {
                formData.append('file', files[i]);
            }

            var xhr = new XMLHttpRequest();
            xhr.open('POST', '` + myhost + `/upload');
            xhr.onloadstart = function() {
                document.getElementById('loading-overlay').style.display = 'flex';
            };
            xhr.onload = function() {
                document.getElementById('loading-overlay').style.display = 'none';
                if (xhr.status === 200) {
                    alert('上传成功');
                } else {
                    alert('上传失败');
                }
            };
            xhr.onerror = function() {
                document.getElementById('loading-overlay').style.display = 'none';
                alert('发生错误，状态码：' + xhr.status);
            };
            xhr.send(formData);
        }
    </script>
</body>
</html>
`