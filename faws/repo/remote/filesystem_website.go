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

type filesystem_website struct {
	base_url url.URL
	client   http.Client
}

func (filesystem_website *filesystem_website) url(name string) (url string, err error) {
	u := filesystem_website.base_url
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

func (filesystem_website *filesystem_website) URI() (s string) {
	s = filesystem_website.base_url.String()
	return
}

func (filesystem_website *filesystem_website) ReadDir(name string) (entries []dir_entry, err error) {
	index_html_file, index_pull_err := filesystem_website.Pull(name + "/")
	if index_pull_err != nil {
		err = index_pull_err
		return
	}

	index_document, index_document_err := html.Parse(index_html_file)
	if index_document_err != nil {
		err = fmt.Errorf("error parsing html: %w", index_document_err)
		return
	}
	index_html_file.Close()

	links, links_err := find_html_links(index_document)
	if links_err != nil {
		err = fmt.Errorf("error finding html links: %w", links_err)
		return
	}

	for _, link := range links {
		var entry dir_entry
		entry.Name, entry.IsDir = strings.CutSuffix(link, "/")
		if !strings.HasPrefix(entry.Name, ".") {
			entries = append(entries, entry)
		}
	}

	return
}

func (filesystem_website *filesystem_website) Stat(name string) (size int64, err error) {
	var url string
	url, err = filesystem_website.url(name)
	if err != nil {
		return
	}

	var (
		request  *http.Request
		response *http.Response
	)
	request, err = http.NewRequest("HEAD", url, nil)
	if err != nil {
		return
	}

	request.Header.Set("User-Agent", fmt.Sprintf("Faws/%s", about.GetVersionString()))

	response, err = filesystem_website.client.Do(request)
	if err != nil {
		return
	}
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("faws/repo/remote: server returned %d %s", response.StatusCode, response.Status)
		return
	}
	response.Body.Close()
	size = response.ContentLength

	return
}

func (filesystem_website *filesystem_website) Pull(name string) (file io.ReadCloser, err error) {
	var url string
	url, err = filesystem_website.url(name)
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

	response, err = filesystem_website.client.Do(request)
	if err != nil {
		return
	}
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("faws/repo/remote: server returned %d %s", response.StatusCode, response.Status)
		return
	}

	file = response.Body
	return
}

func open_filesystem_website(website_url *url.URL) (origin Origin, err error) {
	filesystem_website_ := new(filesystem_website)
	filesystem_website_.base_url = *website_url

	if filesystem_website_.base_url.Scheme == "https" {
		var http_transport http.Transport
		http_transport.MaxConnsPerHost = 32
		http_transport.MaxIdleConnsPerHost = 16
		filesystem_website_.client.Transport = &http_transport
	}

	if !strings.HasSuffix(filesystem_website_.base_url.Path, "/") {
		filesystem_website_.base_url.Path += "/"
	}

	origin = filesystem_origin{filesystem_website_}

	return
}
