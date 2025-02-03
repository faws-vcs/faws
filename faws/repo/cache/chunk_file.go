package cache

// func (index *Index) add_cache_object(object_hash cas.ContentID) {
// 	i := sort.Search(len(index.CacheObjects), func(i int) bool {
// 		return !index.CacheObjects[i].Less(object_hash)
// 	})
// 	if i < len(index.CacheObjects) && index.CacheObjects[i] == object_hash {
// 		return
// 	}
// 	index.CacheObjects = slices.Insert(index.CacheObjects, i, object_hash)
// }

// func (index *Index) is_cache_object()

// func (index *Index) track_file(origin string, objects *cas.Set) (file_id cas.ContentID, err error) {
// 	var (
// 		origin_file *os.File
// 		chunker     multipart.Chunker
// 	)
// 	origin_file, err = os.Open(origin)
// 	if err != nil {
// 		return
// 	}

// 	fmt.Println("scanning", origin)
// 	chunker, err = multipart.NewChunker(origin_file)
// 	if err != nil {
// 		return
// 	}

// 	var (
// 		chunk    []byte
// 		chunk_id cas.ContentID
// 		file     []byte
// 	)
// 	var new bool
// 	for {
// 		// section := chunker.Section()
// 		_, chunk, err = chunker.Next()
// 		if err != nil && errors.Is(err, io.EOF) {
// 			err = nil
// 			break
// 		} else if err != nil {
// 			return
// 		}

// 		// exists := false

// 		new, chunk_id, err = objects.Store(cas.Part, chunk)
// 		if err != nil {
// 			return
// 		}

// 		if new {
// 			index.add_cache_object(chunk_id)
// 		}

// 		// fmt.Println("scanning", cache_file.Origin, section, chunk_id, "already exists", exists)

// 		file = append(file, chunk_id[:]...)
// 	}

// 	new, file_id, err = objects.Store(cas.File, file)
// 	if err != nil {
// 		return
// 	}

// 	if new {
// 		index.add_cache_object(file_id)
// 	}
// 	return

// }
