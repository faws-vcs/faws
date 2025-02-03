![Faws](./doc/img/logo.png)

An experimental version control system for managing large volumes of game files, inspired by Git.

> [!WARNING]
> It might not be good idea to use Faws for serious archival purposes just yet. Your files could be lost.

```
manage authorship identities
  id create    create a new identity for authoring commits
  id rm        remove an identity from the ring
  id primary   make one of your signing identities the primary
  id ls        list all identities in your ring
  id set       alter various identity attributes

sync objects between local and remote repositories
  pull         download a remote into the current directory
  shadow       download the minimum number of objects to checkout a specific commit/tree/file object, creating a "shadow" of the repository

manage repository state
  init         create an empty repository in the current directory
  add          add a file or directory to the index
  rm           remove a file from the index
  status       list files in the index yet to be committed
  write-tree   write cached files to a tree object
  ls-tree      list the contents of a tree object
  commit-tree  create a new commit object using an already-created tree object
  commit       create a new commit object using files from the index
  cat-file     provide contents or details of repository objects
  checkout     export a tree, or a tree of a commit, into a directory

```
