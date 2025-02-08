![Faws](./doc/img/logo.png)

[![Go Reference](https://pkg.go.dev/badge/github.com/faws-vcs/faws.svg)](https://pkg.go.dev/github.com/faws-vcs/faws)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

An experimental version control system for archiving/merging/distributing large volumes of game files, inspired by Git.

> [!WARNING]
> It might not be good idea to use Faws for serious archival purposes just yet. Your files could be lost.


###### Install

```
git clone https://github.com/faws-vcs/faws faws && cd faws
go install github.com/faws-vcs/faws
```

###### Usage

```
manage authorship identities
  id create    create a new identity for authoring commits
  id rm        remove an identity from the ring
  id primary   make one of your signing identities the primary
  id ls        list all identities in your ring
  id set       alter various identity attributes

sync objects between local and remote repositories
  pull         download a remote into the current directory
  shadow       download the minimum portion of a remote necessary to checkout a specific commit or other object

manage repository state
  init         create an empty repository in the current directory
  add          add a file or directory to the index
  rm           remove a file from the index
  status       list files in the index yet to be committed
  write-tree   write cached files to a tree object
  ls-tree      list the contents of a tree object
  commit-tree  create a new commit object using an already-created tree object
  commit       create a new commit object using files from the index
  log          show commit logs
  cat-file     provide contents or details of repository objects
  checkout     export a tree, or a tree of a commit, into a directory
  mass-revise  perfom bulk edits of commits and trees

```
