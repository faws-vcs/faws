package helpinfo

type CategoryEntry struct {
	CategoryID  string
	Description string
	Commands    []string
}

var Text = map[string]string{
	"id create":  "create a new identity for authoring commits",
	"id ls":      "list all identities in your ring",
	"id primary": "make one of your signing identities the primary",
	"id rm":      "remove an identity from the ring",
	"id set":     "alter various identity attributes",

	"pull":   "download a remote into the current directory",
	"shadow": "download the minimum portion of a remote necessary to checkout a specific commit or other object",
	"push":   "push an object and all its dependencies to the remote",

	"init":        "create an empty repository in the current directory",
	"add":         "add a file or directory to the index",
	"rm":          "remove a file from the index",
	"status":      "list files in the index yet to be committed",
	"write-tree":  "write cached files to a tree object",
	"commit-tree": "create a new commit object using an already-created tree object",
	"commit":      "create a new commit object using files from the index",
	"log":         "show commit logs",
	"checkout":    "export a tree, or a tree of a commit, into a directory",
	"cat-file":    "provide contents or details of repository objects",
	"ls-tag":      "list tags and their associated commit hashes",
	"ls-tree":     "list the contents of a tree object",
	"mass-revise": "perfom bulk edits of commits and trees",
}

var Categories = []CategoryEntry{
	{
		"id",
		"manage authorship identities",
		[]string{
			"id create",
			"id rm",
			"id primary",
			"id ls",
			"id set",
		},
	},

	{
		"remote",
		"sync objects between local and remote repositories",
		[]string{
			"pull",
			"shadow",
		},
	},

	{
		"repo",
		"manage repository state",
		[]string{
			"init",
			"add",
			"rm",
			"status",
			"write-tree",
			"ls-tree",
			"commit-tree",
			"commit",
			"log",
			"cat-file",
			"checkout",
			"mass-revise",
		},
	},
}
