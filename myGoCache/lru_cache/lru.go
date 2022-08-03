package lru_cache

import "container/list"

/*
		** 实现缓存的淘汰算法 **
	常用的缓存淘汰策略有FIFO、LFU、LRU等
	LRU算法综合考虑了时间因素和访问频率,故选择LRU算法作为本cache的淘汰策略
	LRU(Least Recently Used), 即淘汰最近最少使用的缓存的策略
	如果数据最近被访问过，那么将来被访问的概率也会更高。
	LRU 算法的实现非常简单，维护一个队列，如果某条记录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可。
*/

// Cache 定义缓存结构，使用一个map维护键和值，再使用一个双向链表来实现队列(链表直接使用标准库的container/list包里已实现的链表)
type Cache struct {
	maxBytes  int64                         //允许使用的最大内存
	nBytes    int64                         //当前已使用的内存
	ll        *list.List                    //链表头指针
	cache     map[string]*list.Element      //缓存map,值是链表中对应节点的指针
	OnEvicted func(key string, value Value) //某条记录被淘汰时的回调函数
}

// Value 缓存中的值是实现该接口的任意类型,Len()函数返回值所占的内存大小
type Value interface {
	Len() int
}

// entry是一个键值对,是双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

// New Cache初始化构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 新增/修改 如果键已存在,则更新对应节点的值,并将节点移到队尾;不存在则新增节点
func (c *Cache) Add(key string, value Value) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element) //将该节点添加到队尾
		kv := element.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		element := c.ll.PushFront(&entry{key, value})
		c.cache[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Get 查找功能,找到链表中对应的节点，然后将该节点移动到队尾
func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element) //将查询的节点移到队尾
		kv := element.Value.(*entry)
		return kv.value, true
	} else {
		return nil, false
	}
}

// RemoveOldest 缓存淘汰函数,淘汰最近最少访问的节点
func (c *Cache) RemoveOldest() {
	element := c.ll.Back() //取队首节点
	if element != nil {
		c.ll.Remove(element) //移除
		kv := element.Value.(*entry)
		delete(c.cache, kv.key) //cacheMap中删除对应键值对
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
