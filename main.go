package main

import (
	"net/http"
	"os/exec"
	"strings"
  "bufio"
  "regexp"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"alert":   false,
			"error":   false,
			"message": "",
		})
	})

  // NOTE:
  // URL一覧を表示する。
  // go用のetcdクライアントは存在するがetcdctl lsが使えないため、
  // shellscriptで対応した。
  // https://github.com/coreos/etcd/tree/master/client
  r.GET("/urls", func(c *gin.Context) {
    cmd := exec.Command("sh", "etcd-ls.sh")
    stdout, err := cmd.StdoutPipe()

    if err != nil {
      c.HTML(http.StatusOK, "urls.tmpl", gin.H{
        "error":   true,
        "message": strings.Join([]string{"error: ", err.Error()}, ""),
      })
    } else {
      cmd.Start()
      scanner := bufio.NewScanner(stdout)
      url := make([]string, 0)
      for scanner.Scan() {
        url = append(url, extract_url(scanner.Text())
      }
      c.HTML(http.StatusOK, "urls.tmpl", gin.H{
        "error":   false,
        "url":     url,
      })
      cmd.Wait()
    }
  })

  r.GET("/urls/:name", func(c *gin.Context) {
    name := c.Param("name")
    cmd := exec.Command("sh", "etcd-ls.sh")
    stdout, err := cmd.StdoutPipe()

    if err != nil {
      c.HTML(http.StatusOK, "user.tmpl", gin.H{
        "error": true,
        "message": strings.Join([]string{"error: ", err.Error()}, ""),
      })
    } else {
      cmd.Start()
      scanner := bufio.NewScanner(stdout)
      url := make([]string, 0)
      for scanner.Scan() {
        result := extract_url(scanner.Text())
        if check_username(name, result) == true {
          url = append(url, result)
        }
      }
      c.HTML(http.StatusOK, "user.tmpl", gin.H{
        "error": false,
        "user": name,
        "url":   url,
      })
      cmd.Wait()
    }
  })

	r.POST("/submit", func(c *gin.Context) {
		username := c.PostForm("username")
		pubKey := c.PostForm("pubKey")

		// libcompose does not support `docker-compose run`...
		out, err := exec.Command("docker-compose", "-p", "paus", "run", "--rm", "gitreceive-upload-key", username, pubKey).Output()

		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.tmpl", gin.H{
				"alert":   true,
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})
		} else {
			c.HTML(http.StatusCreated, "index.tmpl", gin.H{
				"alert":   true,
				"error":   false,
				"message": strings.Join([]string{"fingerprint: ", string(out)}, ""),
			})
		}
	})

	r.Run()
}


func extract_url(str string) string {
  return regexp.MustCompile(`/vulcand/frontends`).ReplaceAllString(str, "")
}

func check_username(reg, str string) bool {
  return regexp.MustCompile(reg).MatchString(str)
}
