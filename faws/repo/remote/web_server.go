package remote

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/faws-vcs/faws/faws/app/about"
	"golang.org/x/net/html"
)

type web_server struct {
	base_url url.URL
	client   http.Client
}

func (ws *web_server) url(name string) (url string, err error) {
	u := ws.base_url
	u.Path += name
	url = u.String()
	return
}

func find_html_links(node *html.Node) (links []string, err error) {
	var find_link func(node *html.Node)
	find_link = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
					break
				}
			}
		} else {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				find_link(child)
			}
		}
	}

	find_link(node)
	return
}

func (ws *web_server) URL() (s string) {
	s = ws.base_url.String()
	return
}

func (ws *web_server) ReadDir(name string) (entries []DirEntry, err error) {
	index_html_file, index_pull_err := ws.Pull(name + "/")
	if index_pull_err != nil {
		err = index_pull_err
		return
	}

	index_document, index_document_err := html.Parse(index_html_file)
	if index_document_err != nil {
		err = index_document_err
		return
	}
	index_html_file.Close()

	links, links_err := find_html_links(index_document)
	if links_err != nil {
		err = links_err
		return
	}

	for _, link := range links {
		var entry DirEntry
		entry.Name, entry.IsDir = strings.CutSuffix(link, "/")
		if !strings.HasPrefix(entry.Name, ".") {
			entries = append(entries, entry)
		}
	}

	return
}

func (ws *web_server) Pull(name string) (file io.ReadCloser, err error) {
	var url string
	url, err = ws.url(name)
	if err != nil {
		return
	}

	var (
		request  *http.Request
		response *http.Response
	)
	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	request.Header.Set("User-Agent", fmt.Sprintf("Faws/%s", about.GetVersionString()))

	response, err = ws.client.Do(request)
	if err != nil {
		return
	}

	file = response.Body
	return
}

func open_web_server(u *url.URL) (fs Fs, err error) {
	ws := new(web_server)
	ws.base_url = *u

	if u.Scheme == "https" {
		ws.client.Transport = &http.Transport{
			MaxConnsPerHost:     32,
			MaxIdleConnsPerHost: 16,
		}
	}

	if !strings.HasSuffix(ws.base_url.Path, "/") {
		ws.base_url.Path += "/"
	}

	fs = ws

	return
}
