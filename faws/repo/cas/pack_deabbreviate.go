package cas

func (pack *Pack) Deabbreviate(abbreviation string) (id ContentID, err error) {
	pack.guard.RLock()
	id, err = pack.index.Deabbreviate(abbreviation)
	pack.guard.RUnlock()
	return
}
