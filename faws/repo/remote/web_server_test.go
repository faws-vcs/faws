package remote

import (
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestHtml(t *testing.T) {
	const msg = `<html>
<head><title>Index of /faws/common/objects/06/f2/</title></head>
<body bgcolor="white">
<h1>Index of /faws/common/objects/06/f2/</h1><hr><pre><a href="../">../</a>
<a href="6fa1b30d1be5e72ddd1fae5883b5ced0a21f">6fa1b30d1be5e72ddd1fae5883b5ced0a21f</a>               04-Feb-2025 11:37     66K
<a href="d866975fd7230d4a026ccc50dd6393d05888">d866975fd7230d4a026ccc50dd6393d05888</a>               04-Feb-2025 11:38     15K
</pre><hr></body>
</html>`

	node, err := html.Parse(strings.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}

	links, err := find_html_links(node)
	if err != nil {
		return
	}

	if !reflect.DeepEqual(links, []string{"../", "6fa1b30d1be5e72ddd1fae5883b5ced0a21f", "d866975fd7230d4a026ccc50dd6393d05888"}) {
		t.Fatal(links)
	}
}
