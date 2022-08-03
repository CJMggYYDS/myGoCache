package myGoCache

/*
	抽象一个只读的数据结构ByteView用来表示缓存值
	方便为lru.Cache添加并发读写特性
*/

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

// ByteSlice 返回一个拷贝,防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
