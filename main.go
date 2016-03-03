package main

import (
	"net/http"
	"os/exec"
	"strings"

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
      rep := regexp.MustCompile(`/vulcand/frontends/`)
      url := make([]string, 0)
      for scanner.Scan() {
        url = append(url, rep.ReplaceAllString(scanner.Text(), ""))
      }
      c.HTML(http.StatusOK, "urls.tmpl", gin.H{
        "error":   false,
        "url":     url,
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
