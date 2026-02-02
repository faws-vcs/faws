package cas

func (pack *Pack) List(fn ListFunc) (err error) {
	pack.guard.RLock()
	err = pack.index.List(fn)
	pack.guard.RUnlock()
	return
}
