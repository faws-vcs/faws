![Faws](./doc/img/logo.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/faws-vcs/faws.svg)](https://pkg.go.dev/github.com/faws-vcs/faws)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

> [!WARNING]
> It might not be good idea to use Faws for serious archival purposes just yet. Your files could be lost.

###### Usage

```
manage authorship identities
  id create    create a new identity for authoring commits
  id rm        remove an identity from the ring
  id primary   make one of your signing identities the primary
  id ls        list all identities in your ring
  id set       alter various identity attributes

sync objects between local and remote repositories
  pull         download a ref (tag/commit/tree/file/part) into the current repository
  clone        download the entire remote repository into the current directory

manage repository state
  init         create an empty repository in the current directory
  add          add a file or directory to the index
  rm           remove a cached file from the index
  chmod        set the permission flag of a file in the index
  status       list files in the index yet to be committed
  write-tree   write cached files to a tree object
  ls-tree      list the contents of a tree object
  commit-tree  create a new commit object using an already-created tree object
  commit       create a new commit object using files from the index
  log          show commit logs
  cat-file     provide contents or details of repository objects
  checkout     export a tree, or a tree of a commit, into a directory
  fsck         enumerate an object hierarchy (and optionally remove) corrupted objects
  mass-revise  correct big mistakes across all tags
  tag          list tags and their associated commit hashes

```

###### Install from source 

```
git clone https://github.com/faws-vcs/faws.git faws && cd faws
go install github.com/faws-vcs/faws
```

###### Rationale

Faws aims to fulfill a very specific use-case: how can you store many different versions of a game without resorting to compression on a massive scale?

This can also be made more abstract: how could you organize many different directories that contain large files with mutually inclusive contents?

A few tools I know of that fill this niche do exist:

- [moonshadow565/rman](https://github.com/moonshadow565/rman)
- [bup/bup](https://github.com/bup/bup)

I could easily use rman, though I didn't necessarily want to download the entirety of a bundle to get only a small portion of the data contained within.

I wanted to have the option to download a single branch at a time, like with Git:

```sh
git clone https://git-example.org/repo.git --branch v0.1 --single-branch
```

Or the equivalent in Faws:
```sh
mkdir repo && cd repo
faws init https://faws-example.org/repo
faws pull v0.1
```

Bup was based on several internal Git systems, but I wanted to use the opposite tool: Something that more or less resembled the porcelain of Git, but wasn't married to any of its plumbing.