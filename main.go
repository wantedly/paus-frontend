package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/etcd/client"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

func appURL(uriScheme, identifier, baseDomain string) string {
	return uriScheme + "://" + identifier + "." + baseDomain
}

func main() {
	baseDomain := os.Getenv("BASE_DOMAIN")
	etcdEndpoint := os.Getenv("ETCD_ENDPOINT")
	uriScheme := os.Getenv("URI_SCHEME")

	if uriScheme == "" {
		uriScheme = "http"
	}

	config := client.Config{
		Endpoints:               []string{etcdEndpoint},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(config)

	if err != nil {
		log.Fatal(err)
	}

	keysAPI := client.NewKeysAPI(c)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"alert":   false,
			"error":   false,
			"message": "",
		})
	})

	r.GET("/urls", func(c *gin.Context) {
		// etcdctl ls --sort /vulcand/frontends
		resp, err := keysAPI.Get(context.Background(), "/vulcand/frontends/", &client.GetOptions{Sort: true})

		if err != nil {
			c.HTML(http.StatusInternalServerError, "urls.tmpl", gin.H{
				"error":   true,
				"message": strings.Join([]string{"error: ", err.Error()}, ""),
			})
		} else {
			appURLs := make([]string, 0)

			for _, node := range resp.Node.Nodes {
				identifier := strings.Replace(node.Key, "/vulcand/frontends/", "", 1)
				appURLs = append(appURLs, appURL(uriScheme, identifier, baseDomain))
			}

			c.HTML(http.StatusOK, "urls.tmpl", gin.H{
				"error":   false,
				"appURLs": appURLs,
			})
		}
	})

	r.GET("/urls/:name", func(c *gin.Context) {
		name := c.Param("name")
		cmd := exec.Command("sh", "etcd-ls.sh")
		stdout, err := cmd.StdoutPipe()

		if err != nil {
			c.HTML(http.StatusInternalServerError, "user.tmpl", gin.H{
				"error":   true,
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
				"user":  name,
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
	return regexp.MustCompile(`/vulcand/frontends/`).ReplaceAllString(str, "")
}

func check_username(reg, str string) bool {
	return regexp.MustCompile(reg).MatchString(str)
}
