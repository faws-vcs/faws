package helpinfo

type CategoryEntry struct {
	CategoryID  string
	Description string
	Commands    []string
}

var Text = map[string]string{
	"id create": "create a new identity for authoring commits",
	"id ls":     "list all identities in your ring",
	"id rm":     "remove an identity from the ring",
	"id set":    "alter various identity attributes",

	"pull":    "download tags or objects into the current repository",
	"clone":   "download an entire remote repository into a directory",
	"publish": "upload a manifest of the repository to the tracker server",
	"seed":    "connect directly with other computers and upload repository objects to them",

	"init":        "create an empty repository in the current directory",
	"add":         "add a file or directory to the index",
	"rm":          "remove a file from the index",
	"reset":       "remove all files from the index and, optionally, move an existing commit back into the index",
	"chmod":       "set the permission flag of a file in the index",
	"status":      "list files in the index yet to be committed",
	"write-tree":  "write cached files to a tree object",
	"commit-tree": "create a new commit object using an already-created tree object",
	"commit":      "create a new commit object using files from the index",
	"log":         "show commit logs",
	"checkout":    "export a tree, or a tree of a commit, into a directory",
	"cat-file":    "provide contents or details of repository objects",
	"tag":         "list tags and their associated commit hashes",
	"ls-tree":     "list the contents of a tree object",
	"prune":       "purge unreachable objects from the cache",
	"pack":        "compile many repository objects into larger files",
	"repack":      "recompile all repository objects into a pack with unreachable objects purged",
	"fsck":        "enumerate an object hierarchy (and optionally remove) corrupted objects",
	"mass-revise": "correct big mistakes across all tags",
}

var Categories = []CategoryEntry{
	{
		"about",
		"read information about your copy of Faws",
		[]string{
			"version",
		},
	},

	{
		"id",
		"manage authorship identities",
		[]string{
			"id create",
			"id rm",
			"id ls",
			"id set",
		},
	},

	{
		"remote",
		"sync objects between local and remote repositories",
		[]string{
			"pull",
			"clone",
			"seed",
			"publish",
		},
	},

	{
		"repo",
		"manage repository state",
		[]string{
			"init",
			"add",
			"rm",
			"chmod",
			"status",
			"write-tree",
			"ls-tree",
			"commit-tree",
			"commit",
			"log",
			"cat-file",
			"checkout",
			"prune",
			"pack",
			"repack",
			"fsck",
			"mass-revise",
			"tag",
		},
	},
}
