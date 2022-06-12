package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"text/template"

	"github.com/gomarkdown/markdown"
)

type Post struct {
	Config          PostConfig
	Content         string
	RenderedContent string
}

type PostConfig struct {
	Title       string
	Slug        string
	Description string
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		panic("not enough arguments passed to cli")
	}

	if args[0] == "build" {
		posts, readErr := ReadPosts()

		if readErr != nil {
			log.Printf("there was a problem reading the posts directory: %+v", readErr)
			panic(readErr)
		}

		postCount, buildErr := BuildSite(posts)

		if buildErr != nil {
			log.Printf("there was a problem building the site: %+v", buildErr)
			panic(buildErr)
		}

		fmt.Printf("blog built. posts written: %s\n", strconv.Itoa(postCount))
	}
}

func ReadPosts() ([]Post, error) {
	var posts []Post

	items, err := ioutil.ReadDir("../posts")
	if err != nil {
		return nil, fmt.Errorf("error reading posts directory")
	}

	for _, item := range items {
		if item.IsDir() {
			post := Post{}
			postDirectory := fmt.Sprintf("../posts/%s", item.Name())
			subItems, err := ioutil.ReadDir(postDirectory)

			if err != nil {
				return nil, fmt.Errorf("error reading post directory: %v", item)
			}

			for _, subItem := range subItems {
				if subItem.Name() == "config.json" {
					filePath := fmt.Sprintf("../posts/%s/config.json", item.Name())
					contents, err := ioutil.ReadFile(filePath)
					if err != nil {
						return nil, fmt.Errorf("error reading config file: %s", filePath)
					}

					if err = json.Unmarshal([]byte(contents), &post.Config); err != nil {
						return nil, fmt.Errorf("error unmarshalling JSON: %s", filePath)
					}
				}

				if subItem.Name() == "post.md" {
					if err := RenderContent(fmt.Sprintf("../posts/%s/post.md", item.Name()), &post); err != nil {
						return nil, err
					}
				}
			}

			posts = append(posts, post)
		}
	}

	return posts, nil
}

func BuildSite(posts []Post) (int, error) {
	truncatePublicDir()

	for key, post := range posts {
		newDir := fmt.Sprintf("../docs/%s", post.Config.Slug)
		err := os.MkdirAll(newDir, os.ModePerm)
		if err != nil {
			return 0, err
		}

		newFilePath := fmt.Sprintf("%s/index.html", newDir)
		fp, err := os.OpenFile(newFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return 0, err
		}

		fp.WriteString(posts[key].RenderedContent)

		if err := fp.Close(); err != nil {
			return 0, err
		}
	}

	return len(posts), nil
}

func RenderContent(filePath string, post *Post) error {
	contents, pathErr := ioutil.ReadFile(filePath)

	if pathErr != nil {
		return fmt.Errorf("error reading file: %s", filePath)
	}

	post.Content = string(markdown.ToHTML(contents, nil, nil))

	template, parseErr := template.ParseFiles("./templates/base.html")
	if parseErr != nil {
		return fmt.Errorf("error reading template file base.html")
	}

	var content bytes.Buffer
	templateErr := template.Execute(&content, post)
	if templateErr != nil {
		return fmt.Errorf("error generating template")
	}

	post.RenderedContent = content.String()
	return nil
}

func truncatePublicDir() {
	os.RemoveAll("../docs")
	os.MkdirAll("../docs", os.ModePerm)
}
