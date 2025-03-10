package page_writer

type DirtyPages interface {
	AddPage(offset int64, data []byte)
	FlushData() error
	ReadDirtyDataAt(data []byte, startOffset int64) (maxStop int64)
	GetStorageOptions() (collection, replication string)
	Destroy()
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
